// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package retry

import (
	"context"
	"math"
	"testing"
	"time"
)

// Inspired by "github.com/ServiceWeaver/weaver/runtime/retry".

func TestRetry(t *testing.T) {
	t.Parallel()

	ctx, cf := context.WithDeadline(context.Background(), time.Now().Add(time.Second))
	defer cf()
	var gaps []time.Duration
	last := time.Now()
	for r := Begin(); r.Continue(ctx); {
		now := time.Now()
		gap := now.Sub(last)
		t.Logf("gap %v", gap)
		gaps = append(gaps, gap)
		last = now
	}
	if len(gaps) >= 100 {
		t.Fatalf("too many retries (%d) in one second", len(gaps))
	}
	if gaps[0] >= 100*time.Millisecond {
		t.Fatalf("first retry took too long: %v", gaps[0])
	}
	if gaps[len(gaps)-1] < 5*gaps[1] {
		t.Fatalf("retries did not increase significantly: %v", gaps)
	}
}

/*
** Comment out for now due to flakiness.
func TestSleepFor(t *testing.T) {
	t.Parallel()

	const N = 20
	const delay = time.Millisecond * 10
	const minDelay = delay - delay/4
	const maxDelay = delay * 2

	over := 0
	for i := 0; i < N; i++ {
		start := time.Now()
		sleep(context.Background(), delay)
		elapsed := time.Since(start)
		t.Logf("sleep duration: %v", elapsed)
		if elapsed < minDelay {
			t.Errorf("sleep returned too early: %v, expecting %v", elapsed, delay)
		} else if elapsed > maxDelay {
			// Allow a couple of over-shoots to reduce flakiness due to slowness.
			over++
			if over > 2 {
				t.Errorf("sleep returned too late: %v, expecting %v", elapsed, delay)
			}
		}
	}
}
*/

func TestSleepCancellation(t *testing.T) {
	t.Parallel()

	const cancelDelay = time.Millisecond * 10
	const sleepDelay = time.Second
	ctx, cf := context.WithTimeout(context.Background(), cancelDelay)
	defer cf()
	start := time.Now()
	sleep(ctx, sleepDelay)
	elapsed := time.Since(start)
	if elapsed >= sleepDelay {
		t.Errorf("sleep not cancelled")
	}
}

func TestRandomization(t *testing.T) {
	t.Parallel()

	const N = 20
	const delay = time.Millisecond * 10
	sumSquares := 0.0
	var sum time.Duration
	for i := 0; i < N; i++ {
		start := time.Now()
		randomizedSleep(context.Background(), delay)
		elapsed := time.Since(start)
		t.Logf("sleep duration: %v", elapsed)
		diff := float64(elapsed - delay)
		sum += elapsed
		sumSquares += diff * diff
	}
	mean := time.Duration(float64(sum) / N)
	if mean < delay/2 || mean >= delay*2 {
		t.Errorf("average sleep interval was too different from specified delay (%v vs. %v)", mean, delay)
	}
	stdDevFraction := math.Sqrt(sumSquares/N) / float64(delay)
	if stdDevFraction < 0.05 {
		t.Errorf("sleep interval was too consistent (+- %.1f%%)", stdDevFraction*100)
	}
}
