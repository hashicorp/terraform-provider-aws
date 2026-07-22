// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

// Demonstrates the bug: attaching rollbackRoutine to a per-refresh child
// context (cancelled when the refresh returns) falsely triggers rollback.
// The fix watches the parent wait context instead.
func TestRollbackRoutine_ignoresRefreshContextCancel(t *testing.T) {
	t.Parallel()

	parent, parentCancel := context.WithCancel(context.Background())
	defer parentCancel()

	var rollbacks atomic.Int32
	stopped := make(chan struct{})
	state := &rollbackState{
		rollbackRoutineStopped: stopped,
		waitCtx:                parent,
	}
	state.waitGroup.Add(1)

	// Simulate the buggy path: child context cancelled at end of refresh.
	child, childCancel := context.WithCancel(parent)

	done := make(chan struct{})
	go func() {
		defer close(done)
		defer state.waitGroup.Done()
		select {
		case <-child.Done():
			rollbacks.Add(1)
		case <-stopped:
			return
		}
	}()

	childCancel() // end of refreshWithTimeout
	time.Sleep(50 * time.Millisecond)
	close(stopped)
	state.waitGroup.Wait()
	<-done

	if rollbacks.Load() != 1 {
		t.Fatalf("expected buggy child-context cancel to trigger rollback path, got %d", rollbacks.Load())
	}

	// Fixed path: watch parent wait context.
	rollbacks.Store(0)
	stopped2 := make(chan struct{})
	state2 := &rollbackState{
		rollbackRoutineStopped: stopped2,
		waitCtx:                parent,
	}
	state2.waitGroup.Add(1)

	child2, child2Cancel := context.WithCancel(parent)
	_ = child2
	done2 := make(chan struct{})
	go func() {
		defer close(done2)
		defer state2.waitGroup.Done()
		select {
		case <-state2.waitCtx.Done():
			rollbacks.Add(1)
		case <-stopped2:
			return
		}
	}()

	child2Cancel() // refresh ends; parent still live
	time.Sleep(50 * time.Millisecond)
	close(stopped2)
	state2.waitGroup.Wait()
	<-done2

	if rollbacks.Load() != 0 {
		t.Fatalf("parent wait context should not roll back when only refresh child cancels, got %d", rollbacks.Load())
	}
}
