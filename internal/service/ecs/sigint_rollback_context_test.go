// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"
	"sync"
	"testing"
	"time"
)

// Documents the refresh-vs-wait context failure mode. This does not call
// production wait/refresh code (no fake ECS client); the acceptance test
// TestAccECSService_LaunchTypeFargate_waitForSteadyState_sigintRollbackNoSignal
// is the regression guard for the service.go call site.
func TestRollbackRoutine_parentWaitContextSurvivesRefreshCancel(t *testing.T) {
	t.Parallel()

	parent := t.Context()
	child, childCancel := context.WithCancel(parent)

	var (
		mu        sync.Mutex
		rollbacks int
	)
	childDoneSeen := make(chan struct{})
	parentDoneSeen := make(chan struct{})
	stop := make(chan struct{})

	// Buggy attach: watch per-refresh child.
	var buggyWG sync.WaitGroup
	buggyWG.Add(1)
	go func() {
		defer buggyWG.Done()
		select {
		case <-child.Done():
			mu.Lock()
			rollbacks++
			mu.Unlock()
			close(childDoneSeen)
		case <-stop:
		}
	}()

	childCancel()
	select {
	case <-childDoneSeen:
	case <-t.Context().Done():
		t.Fatal("timed out waiting for buggy child-context cancel path")
	}
	buggyWG.Wait()

	mu.Lock()
	if rollbacks != 1 {
		mu.Unlock()
		t.Fatalf("expected buggy child-context cancel to trigger rollback path, got %d", rollbacks)
	}
	rollbacks = 0
	mu.Unlock()

	// Fixed attach: watch parent wait context (closed over).
	_, child2Cancel := context.WithCancel(parent)
	waitCtx := parent
	var fixedWG sync.WaitGroup
	fixedWG.Go(func() {
		select {
		case <-waitCtx.Done():
			mu.Lock()
			rollbacks++
			mu.Unlock()
			close(parentDoneSeen)
		case <-stop:
		}
	})

	child2Cancel()

	select {
	case <-parentDoneSeen:
		t.Fatal("parent wait context should not cancel when only refresh child cancels")
	case <-time.After(50 * time.Millisecond):
	}

	close(stop)
	fixedWG.Wait()

	mu.Lock()
	defer mu.Unlock()
	if rollbacks != 0 {
		t.Fatalf("parent wait context should not roll back when only refresh child cancels, got %d", rollbacks)
	}
}
