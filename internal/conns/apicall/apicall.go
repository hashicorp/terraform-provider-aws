// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// Package apicall provides a recorder and Smithy middleware that captures
// every AWS SDK for Go v2 API operation invocation made through the provider's
// service clients. The recorder is opt-in per request via context: when no
// recorder is attached to the operation context, the middleware is a no-op.
//
// The package is intended primarily for acceptance tests, where it lets
// authors assert that a given resource did or did not make a particular API
// call (for example, "Read must not call GetApplicationSettings when the
// optional settings block is absent"). It can also be used by tooling to
// build coverage data over which AWS operations the test suite exercises.
//
// The middleware runs at the end of the Smithy Initialize stage. By that
// point the AWS SDK Go v2 RegisterServiceMetadata middleware has populated
// ServiceID and OperationName on the operation context, and the call to
// next.HandleInitialize delivers the final post-retry error to the recorder.
// As a result, each logical SDK operation is recorded exactly once even if
// the SDK retried the underlying HTTP request multiple times.
package apicall

import (
	"context"
	"slices"
	"sync"
	"time"

	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/smithy-go/middleware"
)

// Call captures a single AWS SDK for Go v2 API operation invocation.
type Call struct {
	// Service is the Smithy ServiceID, e.g. "Pinpoint", "S3".
	Service string
	// Operation is the operation name, e.g. "GetApplicationSettings".
	Operation string
	// Err is the final error returned by the SDK call after retries, or nil
	// on success.
	Err error
	// At is the time at which recording occurred, after the SDK call returned.
	At time.Time
}

// Cursor is an opaque position into a Recorder's call log, suitable for
// scoping assertions to a window of operations (typically a single test
// step).
type Cursor int

// Recorder collects API call records from middleware. The zero value is not
// usable; construct via NewRecorder.
//
// Recorder is safe for concurrent use.
type Recorder struct {
	mu    sync.Mutex
	calls []Call
}

// NewRecorder returns a new, empty Recorder.
func NewRecorder() *Recorder {
	return &Recorder{}
}

// Record appends a call to the log. Intended to be invoked by the middleware
// or by tests; not generally called by user code directly.
func (r *Recorder) Record(service, operation string, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.calls = append(r.calls, Call{
		Service:   service,
		Operation: operation,
		Err:       err,
		At:        time.Now(),
	})
}

// Calls returns a snapshot of all recorded calls.
func (r *Recorder) Calls() []Call {
	r.mu.Lock()
	defer r.mu.Unlock()
	return slices.Clone(r.calls)
}

// Mark returns a cursor pointing at the current end of the call log. Use it
// with CallsSince/ContainsSince to assert behavior in a window starting now.
func (r *Recorder) Mark() Cursor {
	r.mu.Lock()
	defer r.mu.Unlock()
	return Cursor(len(r.calls))
}

// CallsSince returns a snapshot of calls recorded after the given cursor.
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

// Contains reports whether the recorder has captured at least one call to the
// given service operation.
func (r *Recorder) Contains(service, operation string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return containsLocked(r.calls, service, operation)
}

// ContainsSince reports whether the recorder has captured at least one call
// to the given service operation after the given cursor.
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

type contextKeyType int

var contextKey contextKeyType

// NewContext returns ctx with r attached. Middleware installed by Middleware
// will record calls against r when the operation's context descends from the
// returned context.
//
// Passing a nil Recorder returns ctx unchanged so callers can write
// `ctx = apicall.NewContext(ctx, c.CallRecorder())` without a nil check.
func NewContext(ctx context.Context, r *Recorder) context.Context {
	if r == nil {
		return ctx
	}
	return context.WithValue(ctx, contextKey, r)
}

// FromContext extracts the Recorder attached to ctx, if any.
func FromContext(ctx context.Context) (*Recorder, bool) {
	r, ok := ctx.Value(contextKey).(*Recorder)
	return r, ok
}

// middlewareID is the stable identifier under which the recorder middleware
// is registered with the Smithy stack. Exported as a constant for tests and
// for any code that needs to remove or reference the middleware.
const middlewareID = "TerraformProviderAWSCallRecorder"

// recorderMiddleware records the service operation invocation against the
// Recorder attached to the operation context, if any.
//
// It runs at the end of the Smithy Initialize stage so that:
//   - awsmiddleware.RegisterServiceMetadata has already set ServiceID and
//     OperationName on the context.
//   - next.HandleInitialize returns after the rest of the stack (Serialize,
//     Build, Finalize/retry, Deserialize) has run, giving us the final error.
type recorderMiddleware struct{}

func (recorderMiddleware) ID() string { return middlewareID }

func (recorderMiddleware) HandleInitialize(
	ctx context.Context,
	in middleware.InitializeInput,
	next middleware.InitializeHandler,
) (middleware.InitializeOutput, middleware.Metadata, error) {
	out, metadata, err := next.HandleInitialize(ctx, in)

	if rec, ok := FromContext(ctx); ok {
		rec.Record(awsmiddleware.GetServiceID(ctx), awsmiddleware.GetOperationName(ctx), err)
	}

	return out, metadata, err
}

// Middleware returns a stack mutator that registers the call-recording
// middleware on a Smithy stack. Append it to aws.Config.APIOptions (or to a
// per-service Options.APIOptions) once at construction time; the middleware
// itself is the gate, so it imposes no cost on requests whose context has no
// recorder attached.
func Middleware() func(*middleware.Stack) error {
	return func(stack *middleware.Stack) error {
		return stack.Initialize.Add(recorderMiddleware{}, middleware.After)
	}
}
