// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package actionwait

import (
	"errors"
	"strings"
	"time"
)

// TimeoutError is returned when the operation does not reach a success state within Timeout.
type TimeoutError struct {
	LastStatus Status
	Timeout    time.Duration
}

func (e *TimeoutError) Error() string {
	return "timeout waiting for target status after " + e.Timeout.String()
}

// FailureStateError indicates the operation entered a declared failure state.
type FailureStateError struct {
	Status Status
}

func (e *FailureStateError) Error() string {
	return "operation entered failure state: " + string(e.Status)
}

// UnexpectedStateError indicates the operation entered a state outside success/transitional/failure sets.
type UnexpectedStateError struct {
	Status  Status
	Allowed []Status
}

func (e *UnexpectedStateError) Error() string {
	if len(e.Allowed) == 0 {
		return "operation entered unexpected state: " + string(e.Status)
	}
	allowedStr := make([]string, len(e.Allowed))
	for i, s := range e.Allowed {
		allowedStr[i] = string(s)
	}
	return "operation entered unexpected state: " + string(e.Status) + " (allowed: " +
		strings.Join(allowedStr, ", ") + ")"
}

// Error type assertions for compile-time verification
var (
	_ error = (*TimeoutError)(nil)
	_ error = (*FailureStateError)(nil)
	_ error = (*UnexpectedStateError)(nil)
)

// Helper functions for error type checking
func IsTimeout(err error) bool {
	var timeoutErr *TimeoutError
	return errors.As(err, &timeoutErr)
}

func IsFailureState(err error) bool {
	var failureErr *FailureStateError
	return errors.As(err, &failureErr)
}

func IsUnexpectedState(err error) bool {
	var unexpectedErr *UnexpectedStateError
	return errors.As(err, &unexpectedErr)
}
