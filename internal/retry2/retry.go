// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package retry2

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

// FixedDelay returns a delay that is the same through all iterations except the firts (when it is 0).
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

// ExponentialBackoffWithJitterDelay returns a duration of backoffMinDuration * backoffMultiplier**n, with added jitter.
func ExponentialBackoffWithJitterDelay(backoffMinDuration time.Duration, backoffMultiplier float64) DelayFunc {
	return func(n uint) time.Duration {
		if n == 0 {
			return 0
		}

		mult := math.Pow(backoffMultiplier, float64(n))
		d := time.Duration(float64(backoffMinDuration) * mult)
		const jitter = 0.4
		mult = 1 - jitter*rng.Float64() // Subtract up to 40%.
		return time.Duration(float64(d) * mult)
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
	return SDKv2HelperRetryCompatibleDelay(0, 0, 500*time.Millisecond)
}

// Config configures a retry loop.
type Config struct {
	delay       DelayFunc
	gracePeriod time.Duration
	timer       Timer
}

// Option represents an option for retry.
type Option func(*Config)

func emptyOption(c *Config) {}

func WithGracePeriod(d time.Duration) Option {
	return func(c *Config) {
		c.gracePeriod = d
	}
}

func WithDelay(d DelayFunc) Option {
	if d == nil {
		return emptyOption
	}

	return func(c *Config) {
		c.delay = d
	}
}

// WithTimer provides a way to swap out timer module implementations.
// This primarily is useful for mocking/testing, where you may not want to explicitly wait for a set duration
// for retries.
func WithTimer(t Timer) Option {
	return func(c *Config) {
		c.timer = t
	}
}

// Default timer is a wrapper around time.After
type timerImpl struct{}

func (t *timerImpl) After(d time.Duration) <-chan time.Time {
	return time.After(d)
}

// The default Config is backwards compatible with github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry.
func defaultConfig() Config {
	return Config{
		delay:       DefaultSDKv2HelperRetryCompatibleDelay(),
		gracePeriod: 30 * time.Second,
		timer:       &timerImpl{},
	}
}

// RetryWithTimeout holds state for managing retry loops with a timeout.
type RetryWithTimeout struct {
	attempt  uint
	config   Config
	deadline deadline
	timedOut bool
}

// BeginWithOptions returns a new retry loop configured with the provided options.
func BeginWithOptions(timeout time.Duration, opts ...Option) *RetryWithTimeout {
	config := defaultConfig()
	for _, opt := range opts {
		opt(&config)
	}

	return &RetryWithTimeout{
		config:   config,
		deadline: NewDeadline(timeout),
	}
}

// Begin returns a new retry loop configured with the default options.
func Begin(timeout time.Duration) *RetryWithTimeout {
	return BeginWithOptions(timeout)
}

// Continue sleeps between retry attempts.
// It returns false if the timeout has been exceeded.
// The deadline is not checked on the first call to Continue.
func (r *RetryWithTimeout) Continue(ctx context.Context) bool {
	if r.attempt != 0 && r.deadline.Remaining() == 0 {
		if r.config.gracePeriod == 0 {
			r.timedOut = true
			return false
		}

		r.deadline = NewDeadline(r.config.gracePeriod)
		r.config.gracePeriod = 0
	}

	r.sleep(ctx, r.config.delay(r.attempt))
	r.attempt++

	return true
}

// Reset resets a RetryWithTimeout to its initial state.
func (r *RetryWithTimeout) Reset() {
	r.attempt = 0
}

// TimedOut return whether the retry timed out.
func (r *RetryWithTimeout) TimedOut() bool {
	return r.timedOut
}

// Sleeps for the specified duration or until context is done, whichever occurs first.
func (r *RetryWithTimeout) sleep(ctx context.Context, d time.Duration) {
	if d == 0 {
		return
	}

	select {
	case <-ctx.Done():
		return
	case <-r.config.timer.After(d):
	}
}
