// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package errs

import (
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go/aws/awserr"
)

// errorMessager is a simple interface for types with ErrorMessage().
type errorMessager interface {
	ErrorMessage() string
}

func AsContains(err error, target any, message string) bool {
	if errors.As(err, target) {
		if v, ok := target.(errorMessager); ok && strings.Contains(v.ErrorMessage(), message) {
			return true
		}
	}
	return false
}

type ErrorWithErrorMessage interface {
	error
	errorMessager
}

// IsAErrorMessageContains returns whether or not the specified error is of the specified type
// and its ErrorMessage() value contains the specified needle.
func IsAErrorMessageContains[T ErrorWithErrorMessage](err error, needle string) bool {
	as, ok := As[T](err)
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

// MessageContains unwraps the error and returns true if the error matches
// all these conditions:
//   - err is of type awserr.Error, Error.Code() equals code, and Error.Message() contains message
//   - OR err if not of type awserr.Error as string contains both code and message
func MessageContains(err error, code string, message string) bool {
	var awsErr awserr.Error
	if AsContains(err, &awsErr, message) {
		return true
	}

	if Contains(err, code) && Contains(err, message) {
		return true
	}

	return false
}

// IsA indicates whether an error matches an error type
func IsA[T error](err error) bool {
	_, ok := As[T](err)
	return ok
}

// As is equivalent to errors.As(), but returns the value in-line
func As[T error](err error) (T, bool) {
	var as T
	ok := errors.As(err, &as)
	return as, ok
}
