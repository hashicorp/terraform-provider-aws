// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package backoff

import (
	"context"
	"math"
	"math/rand"
	"time"
)

// Inspired by https://github.com/ServiceWeaver/weaver and https://github.com/avast/retry-go.

// Timer represents the timer used to track time for a retry.
type Timer interface {
	After(time.Duration) <-chan time.Time
}

// DelayFunc returns the duration to wait before the next retry attempt.
type DelayFunc func(uint) time.Duration

// FixedDelay returns a delay. The first retry attempt has no delay (0), and subsequent attempts use the fixed delay.
func FixedDelay(delay time.Duration) DelayFunc {
	return func(n uint) time.Duration {
		if n == 0 {
			return 0
		}

		return delay
	}
}

// Do not use the default RNG since we do not want different provider instances
// to pick the same deterministic random sequence.
var rng = rand.New(rand.NewSource(time.Now().UnixNano()))

// ExponentialJitterBackoff returns a duration of backoffMinDuration * backoffMultiplier**n, with added jitter.
func ExponentialJitterBackoff(backoffMinDuration time.Duration, backoffMultiplier float64) DelayFunc {
	return func(n uint) time.Duration {
		if n == 0 {
			return 0
		}

		mult := math.Pow(backoffMultiplier, float64(n))
		return applyJitter(time.Duration(float64(backoffMinDuration) * mult))
	}
}

func applyJitter(base time.Duration) time.Duration {
	const jitterFactor = 0.4
	jitter := 1 - jitterFactor*rng.Float64() // Subtract up to 40%.
	return time.Duration(float64(base) * jitter)
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

// RetryConfig configures a retry loop.
type RetryConfig struct {
	delay       DelayFunc
	gracePeriod time.Duration
	timer       Timer
}

// Option represents an option for retry.
type Option func(*RetryConfig)

func emptyOption(c *RetryConfig) {}

func WithGracePeriod(d time.Duration) Option {
	return func(c *RetryConfig) {
		c.gracePeriod = d
	}
}

func WithDelay(d DelayFunc) Option {
	if d == nil {
		return emptyOption
	}

	return func(c *RetryConfig) {
		c.delay = d
	}
}

// WithTimer provides a way to swap out timer module implementations.
// This primarily is useful for mocking/testing, where you may not want to explicitly wait for a set duration
// for retries.
func WithTimer(t Timer) Option {
	return func(c *RetryConfig) {
		c.timer = t
	}
}

// Default timer is a wrapper around time.After
type timerImpl struct{}

func (t *timerImpl) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}

// The default RetryConfig is backwards compatible with github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry.
func defaultRetryConfig() RetryConfig {
	return RetryConfig{
		delay:       DefaultSDKv2HelperRetryCompatibleDelay(),
		gracePeriod: 30 * time.Second,
		timer:       &timerImpl{},
	}
}

// RetryLoop holds state for managing retry loops with a timeout.
type RetryLoop struct {
	attempt  uint
	config   RetryConfig
	deadline deadline
	timedOut bool
}

// NewRetryLoopWithOptions returns a new retry loop configured with the provided options.
func NewRetryLoopWithOptions(timeout time.Duration, opts ...Option) *RetryLoop {
	config := defaultRetryConfig()
	for _, opt := range opts {
		opt(&config)
	}

	return &RetryLoop{
		config:   config,
		deadline: NewDeadline(timeout + config.gracePeriod),
	}
}

// NewRetryLoop returns a new retry loop configured with the default options.
func NewRetryLoop(timeout time.Duration) *RetryLoop {
	return NewRetryLoopWithOptions(timeout)
}

// Continue sleeps between retry attempts.
// It returns false if the timeout has been exceeded.
// The deadline is not checked on the first call to Continue.
func (r *RetryLoop) Continue(ctx context.Context) bool {
	if r.attempt != 0 && r.deadline.Remaining() == 0 {
		r.timedOut = true

		return false
	}

	r.sleep(ctx, r.config.delay(r.attempt))
	r.attempt++

	return true
}

// Reset resets a RetryLoop to its initial state.
func (r *RetryLoop) Reset() {
	r.attempt = 0
}

// TimedOut return whether the retry timed out.
func (r *RetryLoop) TimedOut() bool {
	return r.timedOut
}

// sleep sleeps for the specified duration or until the context is canceled, whichever occurs first.
func (r *RetryLoop) sleep(ctx context.Context, d time.Duration) {
	if d == 0 {
		return
	}

	select {
	case <-ctx.Done():
		return
	case <-r.config.timer.After(d):
	}
}
