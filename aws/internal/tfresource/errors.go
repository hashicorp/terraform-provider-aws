package tfresource

import (
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// NotFound returns true if the error represents a "resource not found" condition.
// Specifically, NotFound returns true if the error or a wrapped error is of type
// resource.NotFoundError.
func NotFound(err error) bool {
	var e *resource.NotFoundError // nosemgrep: is-not-found-error
	return errors.As(err, &e)
}

// TimedOut returns true if the error represents a "wait timed out" condition.
// Specifically, TimedOut returns true if the error matches all these conditions:
//  * err is of type resource.TimeoutError
//  * TimeoutError.LastError is nil
func TimedOut(err error) bool {
	// This explicitly does *not* match wrapped TimeoutErrors
	timeoutErr, ok := err.(*resource.TimeoutError) // nolint:errorlint
	return ok && timeoutErr.LastError == nil
}

// SetLastError sets the LastError field on the error if supported.
// If lastErr is nil it is ignored.
func SetLastError(err, lastErr error) {
	switch err := err.(type) {
	case *resource.TimeoutError:
		if err.LastError == nil {
			err.LastError = lastErr
		}

	case *resource.UnexpectedStateError:
		if err.LastError == nil {
			err.LastError = lastErr
		}
	}
}
