// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfresource

import (
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

// NotFound returns true if the error represents a "resource not found" condition.
// Specifically, NotFound returns true if the error or a wrapped error is of type
// retry.NotFoundError.
func NotFound(err error) bool {
	var e *retry.NotFoundError // nosemgrep:ci.is-not-found-error
	return errors.As(err, &e)
}

// TimedOut returns true if the error represents a "wait timed out" condition.
// Specifically, TimedOut returns true if the error matches all these conditions:
//   - err is of type retry.TimeoutError
//   - TimeoutError.LastError is nil
func TimedOut(err error) bool {
	timeoutErr, ok := err.(*retry.TimeoutError) //nolint:errorlint // Explicitly does *not* match wrapped TimeoutErrors
	return ok && timeoutErr.LastError == nil
}

// SetLastError sets the LastError field on the error if supported.
// If lastErr is nil it is ignored.
func SetLastError(err, lastErr error) {
	switch err := err.(type) { //nolint:errorlint // Explicitly does *not* match down the error tree
	case *retry.TimeoutError:
		if err.LastError == nil {
			err.LastError = lastErr
		}

	case *retry.UnexpectedStateError:
		if err.LastError == nil {
			err.LastError = lastErr
		}
	}
}
