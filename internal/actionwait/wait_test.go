// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package actionwait

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// fastFixedInterval returns a very small fixed interval to speed tests.
const fastFixedInterval = 5 * time.Millisecond

// makeCtx creates a context with generous overall test timeout safeguard.
func makeCtx(t *testing.T) context.Context { // test helper
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	t.Cleanup(cancel)
	return ctx
}

func TestWaitForStatus_ValidationErrors(t *testing.T) {
	t.Parallel()
	// Subtests parallelized; each uses its own context with timeout.
	cases := map[string]Options[struct{}]{
		"missing timeout":            {SuccessStates: []Status{"ok"}},
		"missing success":            {Timeout: time.Second},
		"negative consecutive":       {Timeout: time.Second, SuccessStates: []Status{"ok"}, ConsecutiveSuccess: -1},
		"negative progress interval": {Timeout: time.Second, SuccessStates: []Status{"ok"}, ProgressInterval: -time.Second},
	}

	for name, opts := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := makeCtx(t)
			_, err := WaitForStatus(ctx, func(context.Context) (FetchResult[struct{}], error) {
				return FetchResult[struct{}]{Status: "irrelevant"}, nil
			}, opts)
			if err == nil {
				t.Fatalf("expected validation error")
			}
		})
	}
}

func TestWaitForStatus_SuccessImmediate(t *testing.T) {
	t.Parallel()
	ctx := makeCtx(t)
	fr, err := WaitForStatus(ctx, func(context.Context) (FetchResult[int], error) {
		return FetchResult[int]{Status: "DONE", Value: 42}, nil
	}, Options[int]{
		Timeout:       250 * time.Millisecond,
		SuccessStates: []Status{"DONE"},
		Interval:      FixedInterval(fastFixedInterval),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fr.Value != 42 || fr.Status != "DONE" {
		t.Fatalf("unexpected result: %#v", fr)
	}
}

func TestWaitForStatus_SuccessAfterTransitions(t *testing.T) {
	t.Parallel()
	ctx := makeCtx(t)
	var calls int32
	fr, err := WaitForStatus(ctx, func(context.Context) (FetchResult[string], error) {
		c := atomic.AddInt32(&calls, 1)
		switch c {
		case 1, 2:
			return FetchResult[string]{Status: "IN_PROGRESS", Value: "step"}, nil
		default:
			return FetchResult[string]{Status: "COMPLETE", Value: "done"}, nil
		}
	}, Options[string]{
		Timeout:            500 * time.Millisecond,
		SuccessStates:      []Status{"COMPLETE"},
		TransitionalStates: []Status{"IN_PROGRESS"},
		Interval:           FixedInterval(fastFixedInterval),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fr.Status != "COMPLETE" || fr.Value != "done" {
		t.Fatalf("unexpected final result: %#v", fr)
	}
}

func TestWaitForStatus_FailureState(t *testing.T) {
	t.Parallel()
	ctx := makeCtx(t)
	fr, err := WaitForStatus(ctx, func(context.Context) (FetchResult[struct{}], error) {
		return FetchResult[struct{}]{Status: "FAILED"}, nil
	}, Options[struct{}]{
		Timeout:       200 * time.Millisecond,
		SuccessStates: []Status{"SUCCEEDED"},
		FailureStates: []Status{"FAILED"},
		Interval:      FixedInterval(fastFixedInterval),
	})
	if err == nil {
		t.Fatal("expected failure error")
	}
	if _, ok := err.(*FailureStateError); !ok { //nolint:errorlint // direct type assertion adequate in tests
		t.Fatalf("expected FailureStateError, got %T", err)
	}
	if fr.Status != "FAILED" {
		t.Fatalf("unexpected status: %v", fr.Status)
	}
}

func TestWaitForStatus_UnexpectedState_WithTransitional(t *testing.T) {
	t.Parallel()
	ctx := makeCtx(t)
	_, err := WaitForStatus(ctx, func(context.Context) (FetchResult[int], error) {
		return FetchResult[int]{Status: "UNKNOWN"}, nil
	}, Options[int]{
		Timeout:            200 * time.Millisecond,
		SuccessStates:      []Status{"OK"},
		TransitionalStates: []Status{"PENDING"},
		Interval:           FixedInterval(fastFixedInterval),
	})
	if err == nil {
		t.Fatal("expected unexpected state error")
	}
	if _, ok := err.(*UnexpectedStateError); !ok { //nolint:errorlint // direct type assertion adequate in tests
		t.Fatalf("expected UnexpectedStateError, got %T", err)
	}
}

func TestWaitForStatus_NoTransitionalListAllowsAnyUntilTimeout(t *testing.T) {
	t.Parallel()
	ctx := makeCtx(t)
	start := time.Now()
	_, err := WaitForStatus(ctx, func(context.Context) (FetchResult[struct{}], error) {
		return FetchResult[struct{}]{Status: "WHATEVER"}, nil
	}, Options[struct{}]{
		Timeout:       50 * time.Millisecond,
		SuccessStates: []Status{"DONE"},
		Interval:      FixedInterval(10 * time.Millisecond),
	})
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if _, ok := err.(*TimeoutError); !ok { //nolint:errorlint // direct type assertion adequate in tests
		t.Fatalf("expected TimeoutError, got %T", err)
	}
	if time.Since(start) < 40*time.Millisecond { // sanity that we actually waited
		t.Fatalf("timeout returned too early")
	}
}

func TestWaitForStatus_ContextCancel(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(makeCtx(t))
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()
	_, err := WaitForStatus(ctx, func(context.Context) (FetchResult[struct{}], error) {
		return FetchResult[struct{}]{Status: "PENDING"}, nil
	}, Options[struct{}]{
		Timeout:       500 * time.Millisecond,
		SuccessStates: []Status{"DONE"},
		Interval:      FixedInterval(fastFixedInterval),
	})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestWaitForStatus_FetchErrorPropagation(t *testing.T) {
	t.Parallel()
	ctx := makeCtx(t)
	testErr := errors.New("boom")
	_, err := WaitForStatus(ctx, func(context.Context) (FetchResult[int], error) {
		return FetchResult[int]{}, testErr
	}, Options[int]{
		Timeout:       200 * time.Millisecond,
		SuccessStates: []Status{"OK"},
		Interval:      FixedInterval(fastFixedInterval),
	})
	if !errors.Is(err, testErr) {
		t.Fatalf("expected fetch error, got %v", err)
	}
}

func TestWaitForStatus_ConsecutiveSuccess(t *testing.T) {
	t.Parallel()
	ctx := makeCtx(t)
	var toggle int32
	// alternate success / transitional until two consecutive successes happen
	fr, err := WaitForStatus(ctx, func(context.Context) (FetchResult[string], error) {
		n := atomic.AddInt32(&toggle, 1)
		// Pattern: BUILDING, READY, READY, READY ... ensures at least two consecutive successes by third attempt
		if n == 1 {
			return FetchResult[string]{Status: "BUILDING", Value: "val"}, nil
		}
		return FetchResult[string]{Status: "READY", Value: "val"}, nil
	}, Options[string]{
		Timeout:            750 * time.Millisecond,
		SuccessStates:      []Status{"READY"},
		TransitionalStates: []Status{"BUILDING"},
		ConsecutiveSuccess: 2,
		Interval:           FixedInterval(2 * time.Millisecond),
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fr.Status != "READY" {
		t.Fatalf("expected READY, got %v", fr.Status)
	}
	if atomic.LoadInt32(&toggle) < 3 { // at least three fetches required (BUILDING, READY, READY)
		t.Fatalf("expected multiple attempts, got %d", toggle)
	}
}

func TestWaitForStatus_ProgressSinkThrottling(t *testing.T) {
	t.Parallel()
	ctx := makeCtx(t)
	var progressCalls int32
	var fetchCalls int32
	_, _ = WaitForStatus(ctx, func(context.Context) (FetchResult[int], error) {
		atomic.AddInt32(&fetchCalls, 1)
		if fetchCalls >= 5 {
			return FetchResult[int]{Status: "DONE"}, nil
		}
		return FetchResult[int]{Status: "WORKING"}, nil
	}, Options[int]{
		Timeout:            500 * time.Millisecond,
		SuccessStates:      []Status{"DONE"},
		TransitionalStates: []Status{"WORKING"},
		Interval:           FixedInterval(5 * time.Millisecond),
		ProgressInterval:   15 * time.Millisecond, // should group roughly 3 polls
		ProgressSink: func(fr FetchResult[any], meta ProgressMeta) {
			atomic.AddInt32(&progressCalls, 1)
			if fr.Status != "WORKING" && fr.Status != "DONE" {
				t.Fatalf("unexpected status in progress sink: %v", fr.Status)
			}
			if meta.NextPollIn <= 0 {
				t.Fatalf("expected positive NextPollIn")
			}
		},
	})
	// With 5 fetch calls and 15ms progress vs 5ms poll, expect fewer progress events than fetches
	if progressCalls <= 1 || progressCalls >= fetchCalls {
		t.Fatalf("unexpected progress call count: %d (fetches %d)", progressCalls, fetchCalls)
	}
}

func TestWaitForStatus_ConsecutiveSuccessDefault(t *testing.T) {
	t.Parallel()
	ctx := makeCtx(t)
	fr, err := WaitForStatus(ctx, func(context.Context) (FetchResult[struct{}], error) {
		return FetchResult[struct{}]{Status: "READY"}, nil
	}, Options[struct{}]{
		Timeout:       100 * time.Millisecond,
		SuccessStates: []Status{"READY"},
		Interval:      FixedInterval(fastFixedInterval),
		// ConsecutiveSuccess left zero to trigger defaulting logic
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if fr.Status != "READY" {
		t.Fatalf("unexpected status: %v", fr.Status)
	}
}

func TestWaitForStatus_ProgressSinkDisabled(t *testing.T) {
	t.Parallel()
	ctx := makeCtx(t)
	var progressCalls int32
	_, err := WaitForStatus(ctx, func(context.Context) (FetchResult[int], error) {
		return FetchResult[int]{Status: "DONE"}, nil
	}, Options[int]{
		Timeout:          100 * time.Millisecond,
		SuccessStates:    []Status{"DONE"},
		Interval:         FixedInterval(fastFixedInterval),
		ProgressInterval: 0, // disabled
		ProgressSink: func(FetchResult[any], ProgressMeta) {
			atomic.AddInt32(&progressCalls, 1)
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if progressCalls != 0 { // should not be invoked when ProgressInterval <= 0
		t.Fatalf("expected zero progress sink calls, got %d", progressCalls)
	}
}

func TestWaitForStatus_UnexpectedStateErrorMessage(t *testing.T) {
	t.Parallel()
	ctx := makeCtx(t)
	_, err := WaitForStatus(ctx, func(context.Context) (FetchResult[int], error) {
		return FetchResult[int]{Status: "UNKNOWN"}, nil
	}, Options[int]{
		Timeout:            200 * time.Millisecond,
		SuccessStates:      []Status{"OK"},
		TransitionalStates: []Status{"PENDING", "IN_PROGRESS"},
		Interval:           FixedInterval(fastFixedInterval),
	})
	if err == nil {
		t.Fatal("expected unexpected state error")
	}
	unexpectedErr, ok := err.(*UnexpectedStateError)
	if !ok { //nolint:errorlint // direct type assertion adequate in tests
		t.Fatalf("expected UnexpectedStateError, got %T", err)
	}
	errMsg := unexpectedErr.Error()
	if !strings.Contains(errMsg, "UNKNOWN") {
		t.Errorf("error message should contain status 'UNKNOWN', got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "allowed:") {
		t.Errorf("error message should list allowed states, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "PENDING") {
		t.Errorf("error message should contain allowed state 'PENDING', got: %s", errMsg)
	}
}
