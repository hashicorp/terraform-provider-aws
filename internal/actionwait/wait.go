// Copyright (c) HashiCorp, Inc.
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
)

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
type IntervalStrategy interface { // nolint:interfacebloat // single method interface
	NextPoll(attempt uint) time.Duration
}

// FixedInterval implements IntervalStrategy with a constant delay.
type FixedInterval time.Duration

// NextPoll returns the fixed duration.
func (fi FixedInterval) NextPoll(uint) time.Duration { return time.Duration(fi) }

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

// ErrTimeout is returned when the operation does not reach a success state within Timeout.
type ErrTimeout struct {
	LastStatus Status
	Timeout    time.Duration
}

func (e *ErrTimeout) Error() string {
	return "timeout waiting for target status after " + e.Timeout.String()
}

// ErrFailureState indicates the operation entered a declared failure state.
type ErrFailureState struct {
	Status Status
}

func (e *ErrFailureState) Error() string {
	return "operation entered failure state: " + string(e.Status)
}

// ErrUnexpectedState indicates the operation entered a state outside success/transitional/failure sets.
type ErrUnexpectedState struct {
	Status  Status
	Allowed []Status
}

func (e *ErrUnexpectedState) Error() string {
	return "operation entered unexpected state: " + string(e.Status)
}

// sentinel errors helpers
var (
	_ error = (*ErrTimeout)(nil)
	_ error = (*ErrFailureState)(nil)
	_ error = (*ErrUnexpectedState)(nil)
)

// WaitForStatus polls using fetch until a success state, failure state, timeout, unexpected state,
// context cancellation, or fetch error occurs.
// On success, the final FetchResult is returned with nil error.
func WaitForStatus[T any](ctx context.Context, fetch FetchFunc[T], opts Options[T]) (FetchResult[T], error) { // nolint:cyclop
	var zero FetchResult[T]

	if opts.Timeout <= 0 {
		return zero, errors.New("actionwait: Timeout must be > 0")
	}
	if len(opts.SuccessStates) == 0 {
		return zero, errors.New("actionwait: at least one SuccessState required")
	}
	if opts.ConsecutiveSuccess <= 0 {
		opts.ConsecutiveSuccess = 1
	}
	if opts.Interval == nil {
		opts.Interval = FixedInterval(30 * time.Second)
	}

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
		if ctx.Err() != nil {
			return last, ctx.Err()
		}
		now := time.Now()
		if now.After(deadline) {
			return last, &ErrTimeout{LastStatus: last.Status, Timeout: opts.Timeout}
		}

		fr, err := fetch(ctx)
		if err != nil {
			return fr, err
		}
		last = fr

		// Classification precedence: failure -> success -> transitional -> unexpected
		if contains(opts.FailureStates, fr.Status) {
			return fr, &ErrFailureState{Status: fr.Status}
		}
		if contains(opts.SuccessStates, fr.Status) {
			successStreak++
			if successStreak >= opts.ConsecutiveSuccess {
				return fr, nil
			}
		} else {
			successStreak = 0
			if len(opts.TransitionalStates) > 0 {
				if !contains(opts.TransitionalStates, fr.Status) {
					return fr, &ErrUnexpectedState{Status: fr.Status, Allowed: allowedTransient}
				}
			}
		}

		// Progress callback throttling
		if opts.ProgressSink != nil && opts.ProgressInterval > 0 {
			if lastProgress.IsZero() || time.Since(lastProgress) >= opts.ProgressInterval {
				nextPoll := opts.Interval.NextPoll(attempt)
				opts.ProgressSink(anyFetchResult(fr), ProgressMeta{
					Attempt:    attempt,
					Elapsed:    time.Since(start),
					Remaining:  maxDuration(0, time.Until(deadline)), // time.Until for clarity
					Deadline:   deadline,
					NextPollIn: nextPoll,
				})
				lastProgress = time.Now()
			}
		}

		// Sleep until next attempt
		sleep := opts.Interval.NextPoll(attempt)
		if sleep > 0 {
			timer := time.NewTimer(sleep)
			select {
			case <-ctx.Done():
				timer.Stop()
				return last, ctx.Err()
			case <-timer.C:
			}
		}
		attempt++
	}
}

// anyFetchResult converts a typed FetchResult[T] into FetchResult[any] for ProgressSink.
func anyFetchResult[T any](fr FetchResult[T]) FetchResult[any] {
	return FetchResult[any]{Status: fr.Status, Value: any(fr.Value)}
}

// contains tests membership in a slice of Status.
func contains(haystack []Status, needle Status) bool {
	return slices.Contains(haystack, needle)
}

func maxDuration(a, b time.Duration) time.Duration {
	if a > b {
		return a
	}
	return b
}
