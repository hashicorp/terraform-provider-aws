// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package apicall

import (
	"context"
	"errors"
	"sync"
	"testing"

	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/smithy-go/middleware"
)

func TestRecorder_RecordAndCalls(t *testing.T) {
	t.Parallel()

	r := NewRecorder()
	r.Record("Pinpoint", "GetApplicationSettings", nil)
	r.Record("Pinpoint", "UpdateApplicationSettings", errors.New("boom"))

	calls := r.Calls()
	if got, want := len(calls), 2; got != want {
		t.Fatalf("len(Calls()) = %d, want %d", got, want)
	}
	if got, want := calls[0].Service, "Pinpoint"; got != want {
		t.Errorf("calls[0].Service = %q, want %q", got, want)
	}
	if got, want := calls[0].Operation, "GetApplicationSettings"; got != want {
		t.Errorf("calls[0].Operation = %q, want %q", got, want)
	}
	if calls[0].Err != nil {
		t.Errorf("calls[0].Err = %v, want nil", calls[0].Err)
	}
	if calls[1].Err == nil || calls[1].Err.Error() != "boom" {
		t.Errorf("calls[1].Err = %v, want boom", calls[1].Err)
	}
}

func TestRecorder_CallsReturnsSnapshot(t *testing.T) {
	t.Parallel()

	r := NewRecorder()
	r.Record("S3", "GetObject", nil)

	snap := r.Calls()
	r.Record("S3", "PutObject", nil)

	if got, want := len(snap), 1; got != want {
		t.Errorf("snapshot mutated by later Record: len=%d, want %d", got, want)
	}
}

func TestRecorder_MarkAndContainsSince(t *testing.T) {
	t.Parallel()

	r := NewRecorder()
	r.Record("Pinpoint", "GetApp", nil)
	r.Record("Pinpoint", "UpdateApplicationSettings", nil)

	mark := r.Mark()

	r.Record("Pinpoint", "GetApp", nil)

	if r.ContainsSince(mark, "Pinpoint", "UpdateApplicationSettings") {
		t.Error("ContainsSince(mark) reported pre-mark call as post-mark")
	}
	if !r.ContainsSince(mark, "Pinpoint", "GetApp") {
		t.Error("ContainsSince(mark) failed to find post-mark call")
	}
	if !r.Contains("Pinpoint", "UpdateApplicationSettings") {
		t.Error("Contains failed to find pre-mark call")
	}
}

func TestRecorder_CallsSince(t *testing.T) {
	t.Parallel()

	r := NewRecorder()
	r.Record("A", "X", nil)
	mark := r.Mark()
	r.Record("B", "Y", nil)
	r.Record("C", "Z", nil)

	got := r.CallsSince(mark)
	if len(got) != 2 || got[0].Service != "B" || got[1].Service != "C" {
		t.Errorf("CallsSince(mark) = %+v, want [B/Y, C/Z]", got)
	}

	if got := r.CallsSince(r.Mark()); got != nil {
		t.Errorf("CallsSince(end) = %+v, want nil", got)
	}
}

func TestRecorder_Reset(t *testing.T) {
	t.Parallel()

	r := NewRecorder()
	r.Record("A", "X", nil)
	r.Record("B", "Y", nil)
	r.Reset()

	if got := r.Calls(); len(got) != 0 {
		t.Errorf("after Reset, Calls() = %+v, want empty", got)
	}
}

func TestRecorder_ConcurrentRecord(t *testing.T) {
	t.Parallel()

	r := NewRecorder()
	const goroutines = 32
	const perGoroutine = 64

	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := range goroutines {
		go func() {
			defer wg.Done()
			for range perGoroutine {
				r.Record("S", "Op", nil)
			}
			_ = i
		}()
	}
	wg.Wait()

	if got, want := len(r.Calls()), goroutines*perGoroutine; got != want {
		t.Errorf("len(Calls()) = %d, want %d (data race or lost record?)", got, want)
	}
}

func TestNewContext_NilRecorder(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	out := NewContext(ctx, nil)
	if out != ctx {
		t.Error("NewContext(ctx, nil) should return ctx unchanged")
	}
	if _, ok := FromContext(out); ok {
		t.Error("FromContext(ctx) found recorder when none was set")
	}
}

func TestNewContext_FromContextRoundTrip(t *testing.T) {
	t.Parallel()

	r := NewRecorder()
	ctx := NewContext(context.Background(), r)
	got, ok := FromContext(ctx)
	if !ok {
		t.Fatal("FromContext failed to find recorder")
	}
	if got != r {
		t.Errorf("FromContext returned %p, want %p", got, r)
	}
}

// TestMiddleware_RecordsServiceAndOperation drives a real smithy stack with
// RegisterServiceMetadata + the recording middleware to verify wiring.
func TestMiddleware_RecordsServiceAndOperation(t *testing.T) {
	t.Parallel()

	r := NewRecorder()
	ctx := NewContext(context.Background(), r)

	stack := middleware.NewStack("test", smithyRequestBuilder)
	if err := stack.Initialize.Add(&awsmiddleware.RegisterServiceMetadata{
		ServiceID:     "Pinpoint",
		OperationName: "GetApplicationSettings",
	}, middleware.Before); err != nil {
		t.Fatalf("adding RegisterServiceMetadata: %v", err)
	}
	if err := Middleware()(stack); err != nil {
		t.Fatalf("adding recorder middleware: %v", err)
	}

	if _, _, err := middleware.DecorateHandler(noopHandler{}, stack).Handle(ctx, nil); err != nil {
		t.Fatalf("stack.Handle: %v", err)
	}

	calls := r.Calls()
	if len(calls) != 1 {
		t.Fatalf("len(Calls()) = %d, want 1", len(calls))
	}
	if got, want := calls[0].Service, "Pinpoint"; got != want {
		t.Errorf("Service = %q, want %q", got, want)
	}
	if got, want := calls[0].Operation, "GetApplicationSettings"; got != want {
		t.Errorf("Operation = %q, want %q", got, want)
	}
	if calls[0].Err != nil {
		t.Errorf("Err = %v, want nil", calls[0].Err)
	}
}

func TestMiddleware_RecordsErrorFromInner(t *testing.T) {
	t.Parallel()

	r := NewRecorder()
	ctx := NewContext(context.Background(), r)
	wantErr := errors.New("simulated failure")

	stack := middleware.NewStack("test", smithyRequestBuilder)
	if err := stack.Initialize.Add(&awsmiddleware.RegisterServiceMetadata{
		ServiceID:     "S3",
		OperationName: "GetObject",
	}, middleware.Before); err != nil {
		t.Fatalf("adding RegisterServiceMetadata: %v", err)
	}
	if err := Middleware()(stack); err != nil {
		t.Fatalf("adding recorder middleware: %v", err)
	}

	_, _, err := middleware.DecorateHandler(failingHandler{err: wantErr}, stack).Handle(ctx, nil)
	if !errors.Is(err, wantErr) {
		t.Fatalf("stack.Handle err = %v, want %v", err, wantErr)
	}

	calls := r.Calls()
	if len(calls) != 1 {
		t.Fatalf("len(Calls()) = %d, want 1", len(calls))
	}
	if !errors.Is(calls[0].Err, wantErr) {
		t.Errorf("recorded Err = %v, want %v", calls[0].Err, wantErr)
	}
}

func TestMiddleware_NoRecorderInContextIsNoop(t *testing.T) {
	t.Parallel()

	stack := middleware.NewStack("test", smithyRequestBuilder)
	if err := stack.Initialize.Add(&awsmiddleware.RegisterServiceMetadata{
		ServiceID:     "S3",
		OperationName: "GetObject",
	}, middleware.Before); err != nil {
		t.Fatalf("adding RegisterServiceMetadata: %v", err)
	}
	if err := Middleware()(stack); err != nil {
		t.Fatalf("adding recorder middleware: %v", err)
	}

	if _, _, err := middleware.DecorateHandler(noopHandler{}, stack).Handle(context.Background(), nil); err != nil {
		t.Fatalf("stack.Handle: %v", err)
	}
}

// smithyRequestBuilder is a minimal stack request builder. The recording
// middleware operates on Initialize parameters which can be anything; the
// request value is irrelevant for these tests.
func smithyRequestBuilder() any { return nil }

type noopHandler struct{}

func (noopHandler) Handle(_ context.Context, _ any) (any, middleware.Metadata, error) {
	return nil, middleware.Metadata{}, nil
}

type failingHandler struct{ err error }

func (h failingHandler) Handle(_ context.Context, _ any) (any, middleware.Metadata, error) {
	return nil, middleware.Metadata{}, h.err
}
