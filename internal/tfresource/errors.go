// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfresource

import (
	"errors"

	"github.com/hashicorp/terraform-provider-aws/internal/retry"
)

// NotFound returns true if the error represents a "resource not found" condition.
// Specifically, NotFound returns true if the error or a wrapped error is of type
// retry.NotFoundError.
//
// Deprecated: NotFound is an alias to a function of the same name in internal/retry
// which handles both Plugin SDK V2 and internal error types. For net-new usage,
// prefer calling retry.NotFound directly.
var NotFound = retry.NotFound

// TimedOut returns true if the error represents a "wait timed out" condition.
// Specifically, TimedOut returns true if the error matches all these conditions:
//   - err is of type retry.TimeoutError
//   - TimeoutError.LastError is nil
//
// Deprecated: TimedOut is an alias to a function of the same name in internal/retry
// which handles both Plugin SDK V2 and internal error types. For net-new usage,
// prefer calling retry.TimedOut directly.
var TimedOut = retry.TimedOut

// SetLastError sets the LastError field on the error if supported.
// If lastErr is nil it is ignored.
//
// Deprecated: SetLastError is an alias to a function of the same name in internal/retry
// which handles both Plugin SDK V2 and internal error types. For net-new usage,
// prefer calling retry.SetLastError directly.
var SetLastError = retry.SetLastError

// From github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry:

// RetryError forces client code to choose whether or not a given error is retryable.
type RetryError struct {
	err         error
	isRetryable bool
}

func (e *RetryError) Error() string {
	return e.err.Error()
}

func (e *RetryError) Unwrap() error {
	return e.err
}

// RetryableError is a helper to create a RetryError that's retryable from a
// given error. To prevent logic errors, will return an error when passed a
// nil error.
func RetryableError(err error) *RetryError {
	if err == nil {
		return &RetryError{
			err: errors.New("empty retryable error received. " +
				"This is a bug with the Terraform AWS Provider and should be " +
				"reported as a GitHub issue in the provider repository."),
			isRetryable: false,
		}
	}
	return &RetryError{err: err, isRetryable: true}
}

// NonRetryableError is a helper to create a RetryError that's _not_ retryable
// from a given error. To prevent logic errors, will return an error when
// passed a nil error.
func NonRetryableError(err error) *RetryError {
	if err == nil {
		return &RetryError{
			err: errors.New("empty non-retryable error received. " +
				"This is a bug with the Terraform AWS Provider and should be " +
				"reported as a GitHub issue in the provider repository."),
			isRetryable: false,
		}
	}
	return &RetryError{err: err, isRetryable: false}
}
