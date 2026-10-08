// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package sync_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/sync"
)

// Adapted from TestOnceValue in sync/once_test.go in the Go standard library

func TestOnceValueCtx(t *testing.T) { //nolint:paralleltest // t.Parallel() cannot be used with testing.AllocsPerRun
	ctx := t.Context()
	calls := 0
	of := func(_ context.Context) int {
		calls++
		return calls
	}
	f := sync.OnceValueCtx(of)
	allocs := testing.AllocsPerRun(10, func() { f(ctx) })
	value := f(ctx)
	if calls != 1 {
		t.Errorf("want calls==1, got %d", calls)
	}
	if value != 1 {
		t.Errorf("want value==1, got %d", value)
	}
	if allocs != 0 {
		t.Errorf("want 0 allocations per call to f, got %v", allocs)
	}
	allocs = testing.AllocsPerRun(10, func() {
		f = sync.OnceValueCtx(of)
	})
	if allocs > 2 {
		t.Errorf("want at most 2 allocations per call to OnceValue, got %v", allocs)
	}
}

func TestOnceValueCtxPanic(t *testing.T) {
	t.Parallel()

	calls := 0
	f := sync.OnceValueCtx(func(_ context.Context) int {
		calls++
		panic("x")
	})
	testOncePanicX(t, &calls, func() { f(t.Context()) })
}

func testOncePanicX(t *testing.T, calls *int, f func()) {
	testOncePanicWith(t, calls, f, func(label string, p any) {
		if p != "x" {
			t.Fatalf("%s: want panic %v, got %v", label, "x", p)
		}
	})
}

func testOncePanicWith(t *testing.T, calls *int, f func(), check func(label string, p any)) {
	// Check that the each call to f panics with the same value, but the
	// underlying function is only called once.
	for _, label := range []string{"first time", "second time"} {
		var p any
		panicked := true
		func() {
			defer func() {
				p = recover()
			}()
			f()
			panicked = false
		}()
		if !panicked {
			t.Fatalf("%s: f did not panic", label)
		}
		check(label, p)
	}
	if *calls != 1 {
		t.Errorf("want calls==1, got %d", *calls)
	}
}
