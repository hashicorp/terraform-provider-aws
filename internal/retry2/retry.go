// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package retry2

import (
	"context"
	"time"
)

// Inspired by https://github.com/ServiceWeaver/weaver and https://github.com/avast/retry-go.

// Timer represents the timer used to track time for a retry.
type Timer interface {
	After(time.Duration) <-chan time.Time
}

// Config configures a retry loop.
type Config struct {
	delay    time.Duration
	maxDelay time.Duration
	timer    Timer
}

// Option represents an option for retry.
type Option func(*Config)

func emptyOption(c *Config) {}

// WithDelay sets the delay between retries.
func WithDelay(delay time.Duration) Option {
	return func(c *Config) {
		c.delay = delay
	}
}

// WithMaxDelay sets the maximum delay between retries.
func WithMaxDelay(maxDelay time.Duration) Option {
	return func(c *Config) {
		c.maxDelay = maxDelay
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
func defaultConfig() *Config {
	return &Config{
		timer: &timerImpl{},
	}
}

// RetryWithTimeout holds state for managing retry loops with a timeout.
type RetryWithTimeout struct {
	config   *Config
	attempt  int
	deadline deadline
}

// BeginWithOptions returns a new retry loop configured with the provided options.
func BeginWithOptions(timeout time.Duration, opts ...Option) *RetryWithTimeout {
	config := defaultConfig()
	for _, opt := range opts {
		opt(config)
	}

	if timeout <= 0 {
		timeout = 1<<63 - 1 // time.maxDuration
	}

	// TODO Backwards compatibel grace period.

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
// The first call does not sleep.
func (r *RetryWithTimeout) Continue(ctx context.Context) bool {
	if r.deadline.Remaining() == 0 {
		return false
	}

	if r.attempt != 0 {
		//randomizedSleep(ctx, r.backoffDelay())
	}
	r.attempt++

	return true
}

// Reset resets a Retry to its initial state.
func (r *RetryWithTimeout) Reset() {
	r.attempt = 0
}
