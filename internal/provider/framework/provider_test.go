// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package framework

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/provider/sdkv2"
)

// To run these benchmarks:
// go test -bench=. -benchmem -run=^$ -v ./internal/provider/framework

// This logs Initialization an annoying number of times
// func BenchmarkFrameworkProviderInitialization(b *testing.B) {
// 	ctx := b.Context()
// 	primary, err := sdkv2.NewProvider(ctx)
// 	if err != nil {
// 		b.Fatalf("Initializing SDKv2 provider: %s", err)
// 	}

// 	// Reset memory counters to zero, so that we only measure the Framework provider initialization.
// 	b.ResetTimer()
// 	for b.Loop() {
// 		_, err := NewProvider(ctx, primary)
// 		if err != nil {
// 			b.Fatalf("Initializing Framework provider: %s", err)
// 		}
// 	}
// }

func BenchmarkFrameworkProvider_validateResourceSchemas(b *testing.B) {
	ctx := b.Context()
	primary, err := sdkv2.NewProvider(ctx)
	if err != nil {
		b.Fatalf("Initializing SDKv2 provider: %s", err)
	}
	p, err := NewProvider(ctx, primary)
	if err != nil {
		b.Fatalf("Initializing Framework provider: %s", err)
	}

	provider := p.(*frameworkProvider)

	// Reset memory counters to zero, so that we only measure the schema validation.
	b.ResetTimer()
	for b.Loop() {
		err := provider.validateResourceSchemas(ctx)
		if err != nil {
			b.Fatalf("Validating resource schemas: %s", err)
		}
	}
}

func BenchmarkFrameworkProvider_SchemaInitialization_DataSource(b *testing.B) {
	ctx := b.Context()
	primary, err := sdkv2.NewProvider(ctx)
	if err != nil {
		b.Fatalf("Initializing SDKv2 provider: %s", err)
	}
	provider, err := NewProvider(ctx, primary)
	if err != nil {
		b.Fatalf("Initializing Framework provider: %s", err)
	}

	// Reset memory counters to zero, so that we only measure the Data Source schema initialization.
	b.ResetTimer()
	for b.Loop() {
		datasources := provider.DataSources(ctx)
		for _, f := range datasources {
			ds := f()
			ds.Schema(ctx, datasource.SchemaRequest{}, &datasource.SchemaResponse{})
		}
	}
}

func BenchmarkFrameworkProvider_SchemaInitialization_Resource(b *testing.B) {
	ctx := b.Context()
	primary, err := sdkv2.NewProvider(ctx)
	if err != nil {
		b.Fatalf("Initializing SDKv2 provider: %s", err)
	}
	provider, err := NewProvider(ctx, primary)
	if err != nil {
		b.Fatalf("Initializing Framework provider: %s", err)
	}

	// Reset memory counters to zero, so that we only measure the Resource schema initialization.
	b.ResetTimer()
	for b.Loop() {
		resources := provider.Resources(ctx)
		for _, f := range resources {
			r := f()
			r.Schema(ctx, resource.SchemaRequest{}, &resource.SchemaResponse{})
		}
	}
}
