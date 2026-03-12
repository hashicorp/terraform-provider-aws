// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// Package actionwait provides a lightweight, action-focused polling helper
// for imperative Terraform actions which need to await asynchronous AWS
// operation completion with periodic user progress events.
package actionwait

import (
	"context"
	"errors"
	"slices"
	"time"

	"github.com/hashicorp/terraform-provider-aws/internal/backoff"
)

// DefaultPollInterval is the default fixed polling interval used when no custom IntervalStrategy is provided.
const DefaultPollInterval = 30 * time.Second

// Status represents a string status value returned from a polled API.
type Status string

// FetchResult wraps the latest status (and optional value) from a poll attempt.
// Value may be a richer SDK structure (pointer) or zero for simple cases.
type FetchResult[T any] struct {
	Status Status
	Value  T
}

// FetchFunc retrieves the latest state of an asynchronous operation.
// It should be side-effect free aside from the remote read.
type FetchFunc[T any] func(context.Context) (FetchResult[T], error)

// IntervalStrategy allows pluggable poll interval behavior (fixed, backoff, etc.).
type IntervalStrategy interface { //nolint:interfacebloat // single method interface (tiny intentional interface)
	NextPoll(attempt uint) time.Duration
}

// FixedInterval implements IntervalStrategy with a constant delay.
type FixedInterval time.Duration

// NextPoll returns the fixed duration.
func (fi FixedInterval) NextPoll(uint) time.Duration { return time.Duration(fi) }

// BackoffInterval implements IntervalStrategy using a backoff.Delay strategy.
// This allows actionwait to leverage sophisticated backoff algorithms while
// maintaining the declarative status-based polling approach.
type BackoffInterval struct {
	delay backoff.Delay
}

// NextPoll returns the next polling interval using the wrapped backoff delay strategy.
func (bi BackoffInterval) NextPoll(attempt uint) time.Duration {
	return bi.delay.Next(attempt)
}

// WithBackoffDelay creates an IntervalStrategy that uses the provided backoff.Delay.
// This bridges actionwait's IntervalStrategy interface with the backoff package's
// delay strategies (fixed, exponential, SDK-compatible, etc.).
//
// Example usage:
//
//	opts := actionwait.Options[MyType]{
//	    Interval: actionwait.WithBackoffDelay(backoff.FixedDelay(time.Second)),
//	    // ... other options
//	}
func WithBackoffDelay(delay backoff.Delay) IntervalStrategy {
	return BackoffInterval{delay: delay}
}

// Options configure the WaitForStatus loop.
type Options[T any] struct {
	Timeout            time.Duration    // Required total timeout.
	Interval           IntervalStrategy // Poll interval strategy (default: 30s fixed).
	ProgressInterval   time.Duration    // Throttle for ProgressSink (default: disabled if <=0).
	SuccessStates      []Status         // Required (>=1) terminal success states.
	TransitionalStates []Status         // Optional allowed in-flight states.
	FailureStates      []Status         // Optional explicit failure states.
	ConsecutiveSuccess int              // Number of consecutive successes required (default 1).
	ProgressSink       func(fr FetchResult[any], meta ProgressMeta)
}

// ProgressMeta supplies metadata for progress callbacks.
type ProgressMeta struct {
	Attempt    uint
	Elapsed    time.Duration
	Remaining  time.Duration
	Deadline   time.Time
	NextPollIn time.Duration
}

// WaitForStatus polls using fetch until a success state, failure state, timeout, unexpected state,
// context cancellation, or fetch error occurs.
// On success, the final FetchResult is returned with nil error.
func WaitForStatus[T any](ctx context.Context, fetch FetchFunc[T], opts Options[T]) (FetchResult[T], error) { //nolint:cyclop // complexity driven by classification/state machine; readability preferred
	if err := validateOptions(opts); err != nil {
		var zero FetchResult[T]
		return zero, err
	}

	normalizeOptions(&opts)

	start := time.Now()
	deadline := start.Add(opts.Timeout)
	var lastProgress time.Time
	var attempt uint
	var successStreak int
	var last FetchResult[T]

	// Precompute allowed states for unexpected classification (success + transitional + failure)
	// Failure states are excluded from Allowed to ensure they classify distinctly.
	allowedTransient := append([]Status{}, opts.SuccessStates...)
	allowedTransient = append(allowedTransient, opts.TransitionalStates...)

	for {
		// Early return: context cancelled
		if ctx.Err() != nil {
			return last, ctx.Err()
		}

		// Early return: timeout exceeded
		if time.Now().After(deadline) {
			return last, &TimeoutError{LastStatus: last.Status, Timeout: opts.Timeout}
		}

		// Fetch current status
		fr, err := fetch(ctx)
		if err != nil {
			return fr, err // Early return: fetch error
		}
		last = fr

		// Classify status and determine if we should terminate
		isTerminal, classifyErr := classifyStatus(fr, opts, &successStreak, allowedTransient)
		if isTerminal {
			return fr, classifyErr // Early return: terminal state (success or failure)
		}

		// Handle progress reporting
		handleProgressReport(opts, fr, start, deadline, attempt, &lastProgress)

		// Sleep until next attempt, with context cancellation check
		if err := sleepWithContext(ctx, opts.Interval.NextPoll(attempt)); err != nil {
			return last, err // Early return: context cancelled during sleep
		}

		attempt++
	}
}

// anyFetchResult converts a typed FetchResult[T] into FetchResult[any] for ProgressSink.
func anyFetchResult[T any](fr FetchResult[T]) FetchResult[any] {
	return FetchResult[any]{Status: fr.Status, Value: any(fr.Value)}
}

func maxDuration(a, b time.Duration) time.Duration {
	if a > b {
		return a
	}
	return b
}

// validateOptions performs early validation of required options.
func validateOptions[T any](opts Options[T]) error {
	if opts.Timeout <= 0 {
		return errors.New("actionwait: Timeout must be > 0")
	}
	if len(opts.SuccessStates) == 0 {
		return errors.New("actionwait: at least one SuccessState required")
	}
	if opts.ConsecutiveSuccess < 0 {
		return errors.New("actionwait: ConsecutiveSuccess cannot be negative")
	}
	if opts.ProgressInterval < 0 {
		return errors.New("actionwait: ProgressInterval cannot be negative")
	}
	return nil
}

// normalizeOptions sets defaults for optional configuration.
func normalizeOptions[T any](opts *Options[T]) {
	if opts.ConsecutiveSuccess <= 0 {
		opts.ConsecutiveSuccess = 1
	}
	if opts.Interval == nil {
		opts.Interval = FixedInterval(DefaultPollInterval)
	}
}

// classifyStatus determines the next action based on the current status.
// Returns: (isTerminal, error) - if isTerminal is true, polling should stop.
func classifyStatus[T any](fr FetchResult[T], opts Options[T], successStreak *int, allowedTransient []Status) (bool, error) {
	// Classification precedence: failure -> success -> transitional -> unexpected
	if slices.Contains(opts.FailureStates, fr.Status) {
		return true, &FailureStateError{Status: fr.Status}
	}

	if slices.Contains(opts.SuccessStates, fr.Status) {
		*successStreak++
		if *successStreak >= opts.ConsecutiveSuccess {
			return true, nil // Success!
		}
		return false, nil // Continue polling for consecutive successes
	}

	// Not a success state, reset streak
	*successStreak = 0

	// Check if transitional state is allowed
	// If TransitionalStates is specified, status must be in that list
	// If TransitionalStates is empty, any non-success/non-failure state is allowed
	if len(opts.TransitionalStates) > 0 && !slices.Contains(opts.TransitionalStates, fr.Status) {
		return true, &UnexpectedStateError{Status: fr.Status, Allowed: allowedTransient}
	}

	return false, nil // Continue polling
}

// handleProgressReport sends progress updates if conditions are met.
func handleProgressReport[T any](opts Options[T], fr FetchResult[T], start time.Time, deadline time.Time, attempt uint, lastProgress *time.Time) {
	if opts.ProgressSink == nil || opts.ProgressInterval <= 0 {
		return
	}

	if lastProgress.IsZero() || time.Since(*lastProgress) >= opts.ProgressInterval {
		nextPoll := opts.Interval.NextPoll(attempt)
		opts.ProgressSink(anyFetchResult(fr), ProgressMeta{
			Attempt:    attempt,
			Elapsed:    time.Since(start),
			Remaining:  maxDuration(0, time.Until(deadline)),
			Deadline:   deadline,
			NextPollIn: nextPoll,
		})
		*lastProgress = time.Now()
	}
}

// sleepWithContext sleeps for the specified duration while respecting context cancellation.
func sleepWithContext(ctx context.Context, duration time.Duration) error {
	if duration <= 0 {
		return nil
	}

	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}
