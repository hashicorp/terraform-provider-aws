// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package backoff

import (
	"context"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestDefaultSDKv2HelperRetryCompatibleDelay(t *testing.T) {
	t.Parallel()

	delay := DefaultSDKv2HelperRetryCompatibleDelay()
	want := []time.Duration{
		0,
		500 * time.Millisecond,
		1 * time.Second,
		2 * time.Second,
		4 * time.Second,
		8 * time.Second,
		10 * time.Second,
		10 * time.Second,
		10 * time.Second,
		10 * time.Second,
	}
	var got []time.Duration
	for i := range len(want) {
		got = append(got, delay(uint(i)))
	}

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("unexpected diff (+wanted, -got): %s", diff)
	}
}

func TestSDKv2HelperRetryCompatibleDelay(t *testing.T) {
	t.Parallel()

	delay := SDKv2HelperRetryCompatibleDelay(200*time.Millisecond, 0, 3*time.Second)
	want := []time.Duration{
		200 * time.Millisecond,
		3 * time.Second,
		6 * time.Second,
		10 * time.Second,
		10 * time.Second,
	}
	var got []time.Duration
	for i := range len(want) {
		got = append(got, delay(uint(i)))
	}

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("unexpected diff (+wanted, -got): %s", diff)
	}
}

func TestSDKv2HelperRetryCompatibleDelayWithPollTimeout(t *testing.T) {
	t.Parallel()

	delay := SDKv2HelperRetryCompatibleDelay(200*time.Millisecond, 20*time.Second, 3*time.Second)
	want := []time.Duration{
		200 * time.Millisecond,
		20 * time.Second,
		20 * time.Second,
		20 * time.Second,
		20 * time.Second,
	}
	var got []time.Duration
	for i := range len(want) {
		got = append(got, delay(uint(i)))
	}

	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("unexpected diff (+wanted, -got): %s", diff)
	}
}

func TestLoopWithTimeoutDefaultGracePeriod(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var n int
	for r := NewLoopWithOptions(1*time.Minute, WithDelay(FixedDelay(1*time.Second))); r.Continue(ctx); {
		time.Sleep(35 * time.Second)
		n++
	}

	// Want = 3 because of default 30s grace period.
	if got, want := n, 3; got != want {
		t.Errorf("Iterations = %v, want %v", got, want)
	}
}

func TestLoopWithTimeoutNoGracePeriod(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	var n int
	for r := NewLoopWithOptions(1*time.Minute, WithDelay(FixedDelay(1*time.Second)), WithGracePeriod(0)); r.Continue(ctx); {
		time.Sleep(35 * time.Second)
		n++
	}

	// Want = 2 because of no grace period.
	if got, want := n, 2; got != want {
		t.Errorf("Iterations = %v, want %v", got, want)
	}
}
