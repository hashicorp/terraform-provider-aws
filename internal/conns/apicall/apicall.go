// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// Package apicall captures AWS SDK for Go v2 operation invocations made
// through the provider's service clients, for use in tests.
//
// The Smithy middleware is opt-in per request: when no Recorder is attached
// to the operation context (see NewContext), it is a no-op.
//
// The middleware runs at the end of Initialize, after RegisterServiceMetadata
// populates ServiceID and OperationName, and captures the final post-retry
// error. Each logical SDK operation is recorded once regardless of retries.
//
// For richer observability — coverage tooling across the suite, latency
// histograms, OTEL export — prefer the smithy-go observability surface
// (TracerProvider / MeterProvider) wired via aws.Config.ServiceOptions.
// smithyoteltracing.Adapt and smithyotelmetrics.Adapt bridge those to a
// real OTEL SDK. This package's recorder is intentionally limited to the
// "did this operation happen" assertion use case.
package apicall

import (
	"context"
	"slices"
	"sync"
	"time"

	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/smithy-go/middleware"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

// Call captures one AWS SDK for Go v2 operation invocation.
type Call struct {
	Service   string        // Smithy ServiceID, e.g. "Pinpoint".
	Operation string        // Operation name, e.g. "GetApplicationSettings".
	Err       error         // Final error after retries, or nil.
	At        time.Time     // Time of recording (after the call returned).
	Duration  time.Duration // Wall-clock time spent in the SDK stack, including retries.
	RequestID string        // AWS request ID from the response, when available.
}

// Cursor is an opaque position into a Recorder's call log. Use Mark to obtain
// one and CallsSince/ContainsSince to scope assertions to a window.
type Cursor int

// Recorder collects API call records. Construct via NewRecorder.
// Safe for concurrent use.
type Recorder struct {
	mu    sync.Mutex
	calls []Call
}

// NewRecorder returns an empty Recorder.
func NewRecorder() *Recorder {
	return &Recorder{}
}

// Record appends a call to the log. Convenience wrapper that fills in only
// the service, operation, error, and timestamp; tests typically use this.
// Middleware uses RecordCall to populate richer fields.
func (r *Recorder) Record(service, operation string, err error) {
	r.RecordCall(Call{
		Service:   service,
		Operation: operation,
		Err:       err,
		At:        time.Now(),
	})
}

// RecordCall appends c to the log. If c.At is zero, it is set to time.Now().
func (r *Recorder) RecordCall(c Call) {
	if c.At.IsZero() {
		c.At = time.Now()
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls = append(r.calls, c)
}

// Calls returns a snapshot of all recorded calls.
func (r *Recorder) Calls() []Call {
	r.mu.Lock()
	defer r.mu.Unlock()
	return slices.Clone(r.calls)
}

// Mark returns a cursor at the current end of the log.
//
// A cursor obtained before Reset points past the end of the post-Reset log;
// CallsSince and ContainsSince treat that as an empty window.
func (r *Recorder) Mark() Cursor {
	r.mu.Lock()
	defer r.mu.Unlock()
	return Cursor(len(r.calls))
}

// CallsSince returns calls recorded at or after c.
func (r *Recorder) CallsSince(c Cursor) []Call {
	r.mu.Lock()
	defer r.mu.Unlock()
	if int(c) >= len(r.calls) {
		return nil
	}
	if c < 0 {
		c = 0
	}
	return slices.Clone(r.calls[c:])
}

// Reset clears the call log.
func (r *Recorder) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls = r.calls[:0]
}

// Contains reports whether service.operation has been recorded.
func (r *Recorder) Contains(service, operation string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return containsLocked(r.calls, service, operation)
}

// ContainsSince reports whether service.operation has been recorded at or
// after c.
func (r *Recorder) ContainsSince(c Cursor, service, operation string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	if int(c) >= len(r.calls) {
		return false
	}
	if c < 0 {
		c = 0
	}
	return containsLocked(r.calls[c:], service, operation)
}

func containsLocked(calls []Call, service, operation string) bool {
	for i := range calls {
		if calls[i].Service == service && calls[i].Operation == operation {
			return true
		}
	}
	return false
}

// recorderKey is the typed context key under which a *Recorder is stored.
var recorderKey = inttypes.NewContextKey[*Recorder]()

// NewContext returns ctx with r attached. The middleware records against r
// for any operation whose context descends from the returned context.
//
// A nil r returns ctx unchanged.
func NewContext(ctx context.Context, r *Recorder) context.Context {
	if r == nil {
		return ctx
	}
	return recorderKey.NewContext(ctx, r)
}

// FromContext extracts the Recorder attached to ctx, if any.
func FromContext(ctx context.Context) (*Recorder, bool) {
	r := recorderKey.FromContext(ctx)
	return r, r != nil
}

// MiddlewareID is the Smithy stack identifier of the recording middleware.
const MiddlewareID = "TerraformProviderAWSCallRecorder"

// recorderMiddleware records each operation against the Recorder attached
// to its context. Runs at Initialize.After: after RegisterServiceMetadata
// populates ctx, and after the rest of the stack returns the final error.
type recorderMiddleware struct{}

func (recorderMiddleware) ID() string { return MiddlewareID }

func (recorderMiddleware) HandleInitialize(
	ctx context.Context,
	in middleware.InitializeInput,
	next middleware.InitializeHandler,
) (middleware.InitializeOutput, middleware.Metadata, error) {
	start := time.Now()
	out, metadata, err := next.HandleInitialize(ctx, in)

	if rec, ok := FromContext(ctx); ok {
		end := time.Now()
		reqID, _ := awsmiddleware.GetRequestIDMetadata(metadata)
		rec.RecordCall(Call{
			Service:   awsmiddleware.GetServiceID(ctx),
			Operation: awsmiddleware.GetOperationName(ctx),
			Err:       err,
			At:        end,
			Duration:  end.Sub(start),
			RequestID: reqID,
		})
	}

	return out, metadata, err
}

// Middleware returns a stack mutator that registers the recording middleware
// on a Smithy stack. Idempotent. Append once to aws.Config.APIOptions; the
// middleware gates itself on a Recorder in the request context.
func Middleware() func(*middleware.Stack) error {
	return func(stack *middleware.Stack) error {
		if _, ok := stack.Initialize.Get(MiddlewareID); ok {
			return nil
		}
		return stack.Initialize.Add(recorderMiddleware{}, middleware.After)
	}
}
