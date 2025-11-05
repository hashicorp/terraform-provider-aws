// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package actionwait

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestTimeoutError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      *TimeoutError
		wantMsg  string
		wantType string
	}{
		{
			name: "with last status",
			err: &TimeoutError{
				LastStatus: "CREATING",
				Timeout:    5 * time.Minute,
			},
			wantMsg:  "timeout waiting for target status after 5m0s",
			wantType: "*actionwait.TimeoutError",
		},
		{
			name: "with empty status",
			err: &TimeoutError{
				LastStatus: "",
				Timeout:    30 * time.Second,
			},
			wantMsg:  "timeout waiting for target status after 30s",
			wantType: "*actionwait.TimeoutError",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.err.Error(); got != tt.wantMsg {
				t.Errorf("TimeoutError.Error() = %q, want %q", got, tt.wantMsg)
			}

			// Verify it implements error interface
			var err error = tt.err
			if got := err.Error(); got != tt.wantMsg {
				t.Errorf("TimeoutError as error.Error() = %q, want %q", got, tt.wantMsg)
			}
		})
	}
}

func TestFailureStateError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		err     *FailureStateError
		wantMsg string
	}{
		{
			name: "with status",
			err: &FailureStateError{
				Status: "FAILED",
			},
			wantMsg: "operation entered failure state: FAILED",
		},
		{
			name: "with empty status",
			err: &FailureStateError{
				Status: "",
			},
			wantMsg: "operation entered failure state: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.err.Error(); got != tt.wantMsg {
				t.Errorf("FailureStateError.Error() = %q, want %q", got, tt.wantMsg)
			}
		})
	}
}

func TestUnexpectedStateError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		err     *UnexpectedStateError
		wantMsg string
	}{
		{
			name: "no allowed states",
			err: &UnexpectedStateError{
				Status:  "UNKNOWN",
				Allowed: nil,
			},
			wantMsg: "operation entered unexpected state: UNKNOWN",
		},
		{
			name: "empty allowed states",
			err: &UnexpectedStateError{
				Status:  "UNKNOWN",
				Allowed: []Status{},
			},
			wantMsg: "operation entered unexpected state: UNKNOWN",
		},
		{
			name: "single allowed state",
			err: &UnexpectedStateError{
				Status:  "UNKNOWN",
				Allowed: []Status{"AVAILABLE"},
			},
			wantMsg: "operation entered unexpected state: UNKNOWN (allowed: AVAILABLE)",
		},
		{
			name: "multiple allowed states",
			err: &UnexpectedStateError{
				Status:  "UNKNOWN",
				Allowed: []Status{"CREATING", "AVAILABLE", "UPDATING"},
			},
			wantMsg: "operation entered unexpected state: UNKNOWN (allowed: CREATING, AVAILABLE, UPDATING)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := tt.err.Error(); got != tt.wantMsg {
				t.Errorf("UnexpectedStateError.Error() = %q, want %q", got, tt.wantMsg)
			}
		})
	}
}

func TestErrorTypeChecking(t *testing.T) {
	t.Parallel()

	// Create instances of each error type
	timeoutErr := &TimeoutError{LastStatus: "CREATING", Timeout: time.Minute}
	failureErr := &FailureStateError{Status: "FAILED"}
	unexpectedErr := &UnexpectedStateError{Status: "UNKNOWN", Allowed: []Status{"AVAILABLE"}}
	genericErr := errors.New("generic error")

	tests := []struct {
		name             string
		err              error
		wantIsTimeout    bool
		wantIsFailure    bool
		wantIsUnexpected bool
	}{
		{
			name:             "TimeoutError",
			err:              timeoutErr,
			wantIsTimeout:    true,
			wantIsFailure:    false,
			wantIsUnexpected: false,
		},
		{
			name:             "FailureStateError",
			err:              failureErr,
			wantIsTimeout:    false,
			wantIsFailure:    true,
			wantIsUnexpected: false,
		},
		{
			name:             "UnexpectedStateError",
			err:              unexpectedErr,
			wantIsTimeout:    false,
			wantIsFailure:    false,
			wantIsUnexpected: true,
		},
		{
			name:             "generic error",
			err:              genericErr,
			wantIsTimeout:    false,
			wantIsFailure:    false,
			wantIsUnexpected: false,
		},
		{
			name:             "nil error",
			err:              nil,
			wantIsTimeout:    false,
			wantIsFailure:    false,
			wantIsUnexpected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := IsTimeout(tt.err); got != tt.wantIsTimeout {
				t.Errorf("IsTimeout(%v) = %v, want %v", tt.err, got, tt.wantIsTimeout)
			}

			if got := IsFailureState(tt.err); got != tt.wantIsFailure {
				t.Errorf("IsFailureState(%v) = %v, want %v", tt.err, got, tt.wantIsFailure)
			}

			if got := IsUnexpectedState(tt.err); got != tt.wantIsUnexpected {
				t.Errorf("IsUnexpectedState(%v) = %v, want %v", tt.err, got, tt.wantIsUnexpected)
			}
		})
	}
}

func TestWrappedErrors(t *testing.T) {
	t.Parallel()

	// Test that error type checking works with wrapped errors
	baseErr := &TimeoutError{LastStatus: "CREATING", Timeout: time.Minute}
	wrappedErr := errors.New("wrapped: " + baseErr.Error())

	// Direct error should be detected
	if !IsTimeout(baseErr) {
		t.Errorf("IsTimeout should detect direct TimeoutError")
	}

	// Wrapped string error should NOT be detected (this is expected behavior)
	if IsTimeout(wrappedErr) {
		t.Errorf("IsTimeout should not detect string-wrapped error")
	}

	// But wrapped with errors.Join should work
	joinedErr := errors.Join(baseErr, errors.New("additional context"))
	if !IsTimeout(joinedErr) {
		t.Errorf("IsTimeout should detect error in errors.Join")
	}
}

func TestErrorMessages(t *testing.T) {
	t.Parallel()

	// Verify error messages contain expected components for debugging
	timeoutErr := &TimeoutError{
		LastStatus: "PENDING",
		Timeout:    2 * time.Minute,
	}

	msg := timeoutErr.Error()
	if !strings.Contains(msg, "timeout") {
		t.Errorf("TimeoutError message should contain 'timeout', got: %q", msg)
	}
	if !strings.Contains(msg, "2m0s") {
		t.Errorf("TimeoutError message should contain timeout duration, got: %q", msg)
	}

	failureErr := &FailureStateError{Status: "ERROR"}
	msg = failureErr.Error()
	if !strings.Contains(msg, "failure state") {
		t.Errorf("FailureStateError message should contain 'failure state', got: %q", msg)
	}
	if !strings.Contains(msg, "ERROR") {
		t.Errorf("FailureStateError message should contain status, got: %q", msg)
	}

	unexpectedErr := &UnexpectedStateError{
		Status:  "WEIRD",
		Allowed: []Status{"GOOD", "BETTER"},
	}
	msg = unexpectedErr.Error()
	if !strings.Contains(msg, "unexpected state") {
		t.Errorf("UnexpectedStateError message should contain 'unexpected state', got: %q", msg)
	}
	if !strings.Contains(msg, "WEIRD") {
		t.Errorf("UnexpectedStateError message should contain actual status, got: %q", msg)
	}
	if !strings.Contains(msg, "GOOD, BETTER") {
		t.Errorf("UnexpectedStateError message should contain allowed states, got: %q", msg)
	}
}
