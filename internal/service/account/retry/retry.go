// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package retry

import (
	"context"
	"math"
	"math/rand"
	"time"
)

// Inspired by "github.com/ServiceWeaver/weaver/runtime/retry".

// Options configure a retry loop.
// Before the ith iteration of the loop, retry.Continue() sleeps for a duraion of BackoffMinDuration * BackoffMultiplier**i, with added jitter.
type Options struct {
	BackoffMinDuration time.Duration
	BackoffMultiplier  float64 // If specified, must be at least 1.
}

var defaultOptions = Options{
	BackoffMinDuration: 10 * time.Millisecond,
	BackoffMultiplier:  1.3,
}

// Retry holds state for managing retry loops with exponential backoff and jitter.
type Retry struct {
	options Options
	attempt int
}

// BeginWithOptions returns a new retry loop configured with the provided options.
func BeginWithOptions(options Options) *Retry {
	return &Retry{options: options}
}

// Begin returns a new retry loop configured with the default options.
func Begin() *Retry {
	return BeginWithOptions(defaultOptions)
}

// Continue sleeps for an exponentially increasing interval (with jitter).
// It stops its sleep early and returns false if context becomes done.
// If the return value is false, ctx.Err() is guaranteed to be non-nil.
// The first call does not sleep.
func (r *Retry) Continue(ctx context.Context) bool {
	if r.attempt != 0 {
		randomizedSleep(ctx, r.backoffDelay())
	}
	r.attempt++
	return ctx.Err() == nil
}

// Reset resets a Retry to its initial state.
// It's useful when you only want to retry a failing operation.
func (r *Retry) Reset() {
	r.attempt = 0
}

func (r *Retry) backoffDelay() time.Duration {
	mult := math.Pow(r.options.BackoffMultiplier, float64(r.attempt))
	return time.Duration(float64(r.options.BackoffMinDuration) * mult)
}

func randomizedSleep(ctx context.Context, d time.Duration) {
	const jitter = 0.4
	mult := 1 - jitter*rand.Float64() // Subtract up to 40%.
	sleep(ctx, time.Duration(float64(d)*mult))
}

func sleep(ctx context.Context, d time.Duration) {
	t := time.NewTimer(d)
	select {
	case <-ctx.Done():
		t.Stop()
	case <-t.C:
	}
}
