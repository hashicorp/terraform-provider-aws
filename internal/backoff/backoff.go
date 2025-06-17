// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package backoff

import (
	"context"
	"time"

	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

// Inspired by https://github.com/ServiceWeaver/weaver and https://github.com/avast/retry-go.

// Timer represents the timer used to track time.
type Timer interface {
	After(time.Duration) <-chan time.Time
}

// DelayFunc returns the duration to wait before the next attempt.
type DelayFunc func(uint) time.Duration

// FixedDelay returns a delay. The first attempt has no delay (0), and subsequent attempts use the fixed delay.
func FixedDelay(delay time.Duration) DelayFunc {
	return func(n uint) time.Duration {
		if n == 0 {
			return 0
		}

		return delay
	}
}

type sdkv2HelperRetryCompatibleDelay struct {
	minTimeout   time.Duration
	pollInterval time.Duration
	wait         time.Duration
}

func (d *sdkv2HelperRetryCompatibleDelay) delay() time.Duration {
	wait := d.wait

	// First round had no wait.
	if wait == 0 {
		wait = 100 * time.Millisecond
	}

	wait *= 2

	// If a poll interval has been specified, choose that interval.
	// Otherwise bound the default value.
	if d.pollInterval > 0 && d.pollInterval < 180*time.Second {
		wait = d.pollInterval
	} else {
		if wait < d.minTimeout {
			wait = d.minTimeout
		} else if wait > 10*time.Second {
			wait = 10 * time.Second
		}
	}

	d.wait = wait

	return wait
}

// SDKv2HelperRetryCompatibleDelay returns a Terraform Plugin SDK v2 helper/retry-compatible delay.
func SDKv2HelperRetryCompatibleDelay(initialDelay, pollInterval, minTimeout time.Duration) DelayFunc {
	delay := &sdkv2HelperRetryCompatibleDelay{
		minTimeout:   minTimeout,
		pollInterval: pollInterval,
	}

	return func(n uint) time.Duration {
		if n == 0 {
			return initialDelay
		}

		return delay.delay()
	}
}

// DefaultSDKv2HelperRetryCompatibleDelay returns a Terraform Plugin SDK v2 helper/retry-compatible delay
// with default values (from the `RetryContext` function).
func DefaultSDKv2HelperRetryCompatibleDelay() DelayFunc {
	return SDKv2HelperRetryCompatibleDelay(0, 0, 500*time.Millisecond) //nolint:mnd // 500ms is the Plugin SDKv2 default
}

// LoopConfig configures a loop.
type LoopConfig struct {
	delay       DelayFunc
	gracePeriod time.Duration
	timer       Timer
}

// Option represents a loop option.
type Option func(*LoopConfig)

func emptyOption(c *LoopConfig) {}

func WithGracePeriod(d time.Duration) Option {
	return func(c *LoopConfig) {
		c.gracePeriod = d
	}
}

func WithDelay(d DelayFunc) Option {
	if d == nil {
		return emptyOption
	}

	return func(c *LoopConfig) {
		c.delay = d
	}
}

// WithTimer provides a way to swap out timer module implementations.
// This primarily is useful for mocking/testing, where you may not want to explicitly wait for a set duration
// for retries.
func WithTimer(t Timer) Option {
	return func(c *LoopConfig) {
		c.timer = t
	}
}

// Default timer is a wrapper around time.After
type timerImpl struct{}

func (t *timerImpl) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}

// The default RetryConfig is backwards compatible with github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry.
func defaultLoopConfig() LoopConfig {
	return LoopConfig{
		delay:       DefaultSDKv2HelperRetryCompatibleDelay(),
		gracePeriod: 30 * time.Second,
		timer:       &timerImpl{},
	}
}

// Loop holds state for managing loops with a timeout.
type Loop struct {
	attempt  uint
	config   LoopConfig
	deadline inttypes.Deadline
	timedOut bool
}

// NewLoopWithOptions returns a new loop configured with the provided options.
func NewLoopWithOptions(timeout time.Duration, opts ...Option) *Loop {
	config := defaultLoopConfig()
	for _, opt := range opts {
		opt(&config)
	}

	return &Loop{
		config:   config,
		deadline: inttypes.NewDeadline(timeout + config.gracePeriod),
	}
}

// NewLoop returns a new loop configured with the default options.
func NewLoop(timeout time.Duration) *Loop {
	return NewLoopWithOptions(timeout)
}

// Continue sleeps between attempts.
// It returns false if the timeout has been exceeded.
// The deadline is not checked on the first call to Continue.
func (r *Loop) Continue(ctx context.Context) bool {
	if r.attempt != 0 && r.deadline.Remaining() == 0 {
		r.timedOut = true

		return false
	}

	r.sleep(ctx, r.config.delay(r.attempt))
	r.attempt++

	return context.Cause(ctx) == nil
}

// Reset resets a Loop to its initial state.
func (r *Loop) Reset() {
	r.attempt = 0
}

// TimedOut return whether the loop timed out.
func (r *Loop) TimedOut() bool {
	return r.timedOut
}

// sleep sleeps for the specified duration or until the context is canceled, whichever occurs first.
func (r *Loop) sleep(ctx context.Context, d time.Duration) {
	if d == 0 {
		return
	}

	select {
	case <-ctx.Done():
		return
	case <-r.config.timer.After(d):
	}
}
