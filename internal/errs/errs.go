// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package errs

import (
	"errors"
	"strings"
)

// errorMessager is a simple interface for types with ErrorMessage().
type errorMessager interface {
	ErrorMessage() string
}

type ErrorWithErrorMessage interface {
	error
	errorMessager
}

// IsAErrorMessageContains returns whether or not the specified error is of the specified type
// and its ErrorMessage() value contains the specified needle.
func IsAErrorMessageContains[T ErrorWithErrorMessage](err error, needle string) bool {
	as, ok := errors.AsType[T](err)
	if ok {
		return strings.Contains(as.ErrorMessage(), needle)
	}
	return false
}

// Contains returns true if the error matches all these conditions:
//   - err as string contains needle
func Contains(err error, needle string) bool {
	if err != nil && strings.Contains(err.Error(), needle) {
		return true
	}
	return false
}

// IsA indicates whether an error matches an error type
func IsA[T error](err error) bool {
	_, ok := errors.AsType[T](err)
	return ok
}

var _ ErrorWithErrorMessage = &MessageError{}

// MessageError is a simple error type that implements the errorMessager
type MessageError struct {
	error
}

func (e *MessageError) ErrorMessage() string {
	if e == nil || e.error == nil {
		return ""
	}
	return e.Error()
}

// NewMessageError returns a new MessageError
func NewMessageError(err error) *MessageError {
	return &MessageError{error: err}
}
