// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package retry

import (
	"errors"
	"fmt"
	"strings"
	"time"

	sdkretry "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

//
// Based on https://github.com/hashicorp/terraform-plugin-sdk/helper/retry/error.go.
//

var ErrFoundResource = errors.New(`found resource`)

// NotFound returns true if the error represents a "resource not found" condition.
// Specifically, NotFound returns true if the error or a wrapped error is of type
// retry.NotFoundError
func NotFound(err error) bool {
	// Handle both internal and Plugin SDK V2 error variants
	var e1 *NotFoundError          // nosemgrep:ci.is-not-found-error
	var e2 *sdkretry.NotFoundError // nosemgrep:ci.is-not-found-error
	return errors.As(err, &e1) || errors.As(err, &e2)
}

// TimedOut returns true if the error represents a "wait timed out" condition.
// Specifically, TimedOut returns true if the error matches all these conditions:
//   - err is of type retry.TimeoutError
//   - TimeoutError.LastError is nil
func TimedOut(err error) bool {
	// Handle both internal and Plugin SDK V2 error variants
	timeoutErr, ok := err.(*TimeoutError)                //nolint:errorlint // Explicitly does *not* match wrapped TimeoutErrors
	sdkTimeoutErr, sdkOk := err.(*sdkretry.TimeoutError) //nolint:errorlint // Explicitly does *not* match wrapped TimeoutErrors
	return (ok && timeoutErr.LastError == nil) || (sdkOk && sdkTimeoutErr.LastError == nil)
}

// SetLastError sets the LastError field on the error if supported.
// If lastErr is nil it is ignored.
func SetLastError(err, lastErr error) {
	// Handle both internal and Plugin SDK V2 error variants
	switch err := err.(type) { //nolint:errorlint // Explicitly does *not* match down the error tree
	case *TimeoutError:
		if err.LastError == nil {
			err.LastError = lastErr
		}
	case *sdkretry.TimeoutError:
		if err.LastError == nil {
			err.LastError = lastErr
		}

	case *UnexpectedStateError:
		if err.LastError == nil {
			err.LastError = lastErr
		}
	case *sdkretry.UnexpectedStateError:
		if err.LastError == nil {
			err.LastError = lastErr
		}
	}
}

type NotFoundError struct {
	LastError error
	Message   string
	Retries   int
}

func (e *NotFoundError) Error() string {
	if e.Message != "" {
		return e.Message
	}

	if e.Retries > 0 {
		return fmt.Sprintf("couldn't find resource (%d retries)", e.Retries)
	}

	return "couldn't find resource"
}

func (e *NotFoundError) Unwrap() error {
	return e.LastError
}

// UnexpectedStateError is returned when Refresh returns a state that's neither in Target nor Pending.
type UnexpectedStateError struct {
	LastError     error
	State         string
	ExpectedState []string
}

func (e *UnexpectedStateError) Error() string {
	message := fmt.Sprintf(
		"unexpected state '%s', wanted target '%s'",
		e.State,
		strings.Join(e.ExpectedState, ", "),
	)
	if e.LastError != nil {
		message += fmt.Sprintf(". last error: %s",
			e.LastError,
		)
	}
	return message
}

func (e *UnexpectedStateError) Unwrap() error {
	return e.LastError
}

// TimeoutError is returned when WaitForState times out.
type TimeoutError struct {
	LastError     error
	LastState     string
	Timeout       time.Duration
	ExpectedState []string
}

func (e *TimeoutError) Error() string {
	expectedState := "resource to be gone"
	if len(e.ExpectedState) > 0 {
		expectedState = fmt.Sprintf("state to become '%s'", strings.Join(e.ExpectedState, ", "))
	}

	extraInfo := make([]string, 0)
	if e.LastState != "" {
		extraInfo = append(extraInfo, fmt.Sprintf("last state: '%s'", e.LastState))
	}
	if e.Timeout > 0 {
		extraInfo = append(extraInfo, fmt.Sprintf("timeout: %s", e.Timeout.String()))
	}

	suffix := ""
	if len(extraInfo) > 0 {
		suffix = fmt.Sprintf(" (%s)", strings.Join(extraInfo, ", "))
	}

	if e.LastError != nil {
		return fmt.Sprintf("timeout while waiting for %s%s: %s",
			expectedState, suffix, e.LastError)
	}

	return fmt.Sprintf("timeout while waiting for %s%s",
		expectedState, suffix)
}

func (e *TimeoutError) Unwrap() error {
	return e.LastError
}
