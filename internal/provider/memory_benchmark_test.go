// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package provider_test

import (
	"runtime"
	"runtime/debug"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov5"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
)

// These benchmarks measure the provider's memory footprint in two phases:
// constructing the muxed server, and serving the first GetProviderSchema call.
//
// Both depend on a cold schema cache, and the cache is process-global: once any
// GetProviderSchema call builds it, every later call in the same process hits a
// warm cache. A looping benchmark would therefore under-report, and the two
// benchmarks cannot share a process. Each must be run with a single iteration,
// in its own fresh process:
//
//	go test -bench=BenchmarkProviderInitialization -benchtime=1x -benchmem -run=^$ -v ./internal/provider
//	go test -bench=BenchmarkProviderGetProviderSchema -benchtime=1x -benchmem -run=^$ -v ./internal/provider
//
// Statistical significance comes from invoking each benchmark across multiple
// fresh processes and aggregating the results (for example, with benchstat), not
// from increasing the iteration count within a single process.
//
// The standard -benchmem columns (B/op, allocs/op) report the bytes and number
// of allocations for the measured phase, including memory allocated and then
// discarded. The custom heap metrics report live heap after a full garbage
// collection; they are absolute, including the process baseline heap and memory
// retained by package init(), because those are part of the provider's resident
// footprint. Two measures are reported: heap-alloc-bytes (exact live-object
// bytes, the lowest-noise measure of live data) and heap-inuse-bytes (whole
// spans in use, which tracks resident memory more closely). Their difference is
// heap fragmentation.

// BenchmarkProviderInitialization measures constructing the muxed provider
// server. B/op and allocs/op report the construction churn; heap-alloc-bytes and
// heap-inuse-bytes report the live heap afterward, with only the server
// referenced.
//
// Custom metrics:
//   - heap-alloc-bytes: exact live-object bytes after construction.
//   - heap-inuse-bytes: whole spans in use after construction.
func BenchmarkProviderInitialization(b *testing.B) {
	if b.N != 1 {
		b.Fatalf("must be run with -benchtime=1x for an accurate cold-start measurement; got b.N = %d", b.N)
	}

	restore := pinRuntime()
	defer restore()

	ctx := b.Context()
	b.ReportAllocs()

	factory, _, err := provider.ProtoV5ProviderServerFactory(ctx)
	if err != nil {
		b.Fatal(err)
	}

	// Instantiate the muxed server so that the measurement includes everything built
	// during construction.
	server := factory()

	heapAlloc, heapInuse := heapStats()
	runtime.KeepAlive(server)

	b.ReportMetric(float64(heapAlloc), "heap-alloc-bytes")
	b.ReportMetric(float64(heapInuse), "heap-inuse-bytes")
}

// BenchmarkProviderGetProviderSchema measures the GetProviderSchema call,
// which builds and serializes every schema. Excludes the construction of the
// muxed server.
//
// Custom metrics:
//   - peak-heap-inuse-bytes: peak memory usage while both the server and the response are referenced.
//   - heap-alloc-bytes: exact live-object bytes after the response has been released.
//   - heap-inuse-bytes: whole spans in use after the response has been released.
func BenchmarkProviderGetProviderSchema(b *testing.B) {
	if b.N != 1 {
		b.Fatalf("must be run with -benchtime=1x for an accurate cold-start measurement; got b.N = %d", b.N)
	}

	restore := pinRuntime()
	defer restore()

	ctx := b.Context()

	factory, _, err := provider.ProtoV5ProviderServerFactory(ctx)
	if err != nil {
		b.Fatal(err)
	}
	server := factory()

	// Exclude construction from the measurement. ResetTimer also resets the
	// memory allocation counters, so B/op and allocs/op reflect only the
	// GetProviderSchema call below. The server must not have had
	// GetProviderSchema called before this point, or the cache would be warm.
	b.ReportAllocs()
	b.ResetTimer()

	resp, err := server.GetProviderSchema(ctx, &tfprotov5.GetProviderSchemaRequest{})
	if err != nil {
		b.Fatalf("calling GetProviderSchema: %s", err)
	}

	for _, diag := range resp.Diagnostics {
		if diag.Severity == tfprotov5.DiagnosticSeverityError {
			b.Fatalf("GetProviderSchema returned an error diagnostic: %s: %s", diag.Summary, diag.Detail)
		}
	}

	// Fail if any category of schema-bearing type is empty, so that a regression
	// that silently drops a whole category from the serialized response does not
	// pass as a memory improvement.
	if len(resp.ResourceSchemas) == 0 {
		b.Error("GetProviderSchema response contains no resource schemas")
	}
	if len(resp.DataSourceSchemas) == 0 {
		b.Error("GetProviderSchema response contains no data source schemas")
	}
	if len(resp.EphemeralResourceSchemas) == 0 {
		b.Error("GetProviderSchema response contains no ephemeral resource schemas")
	}
	if len(resp.ListResourceSchemas) == 0 {
		b.Error("GetProviderSchema response contains no list resource schemas")
	}
	if len(resp.ActionSchemas) == 0 {
		b.Error("GetProviderSchema response contains no action schemas")
	}
	if len(resp.Functions) == 0 {
		b.Error("GetProviderSchema response contains no function schemas")
	}
	if b.Failed() {
		return
	}

	// Peak: live heap in use while both the server and the response are
	// referenced, the high-water mark of a schema request. The KeepAlive marks
	// the last point the response must be live; after it, the response is
	// unreachable and the next collection may reclaim it.
	_, peakInuse := heapStats()
	runtime.KeepAlive(resp)

	// Primary: with the response no longer referenced, measure live heap with
	// only the server referenced. This isolates what the provider holds onto
	// internally after a GetProviderSchema call completes.
	heapAlloc, heapInuse := heapStats()
	runtime.KeepAlive(server)

	b.ReportMetric(float64(peakInuse), "peak-heap-inuse-bytes")
	b.ReportMetric(float64(heapAlloc), "heap-alloc-bytes")
	b.ReportMetric(float64(heapInuse), "heap-inuse-bytes")
}

// pinRuntime makes the memory measurement deterministic regardless of how the
// benchmark is invoked: it disables automatic garbage collection (explicit
// runtime.GC() calls still run) so collections do not fire at unpredictable
// points, and pins GOMAXPROCS to remove scheduling variance. It returns a
// function that restores the previous settings.
func pinRuntime() func() {
	prevGC := debug.SetGCPercent(-1)
	prevProcs := runtime.GOMAXPROCS(1)
	return func() {
		runtime.GOMAXPROCS(prevProcs)
		debug.SetGCPercent(prevGC)
	}
}

// heapStats returns the live heap after settling the collector. GC is run
// twice: the first collection reclaims dead objects, the second settles spans
// freed by the first so that the reading reflects only live memory.
//
// Both measures are returned. heapAlloc is exact live-object bytes: the
// lowest-noise measure of how much live data the provider holds. heapInuse
// counts whole spans in use, including span fragmentation, and so tracks
// resident memory more closely. Their difference is heap fragmentation.
func heapStats() (heapAlloc, heapInuse uint64) {
	runtime.GC()
	runtime.GC()
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	return m.HeapAlloc, m.HeapInuse
}
