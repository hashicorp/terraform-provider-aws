// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package retry2

import (
	"fmt"
	"strings"
	"time"

	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
)

//
// Based on https://github.com/hashicorp/terraform-plugin-sdk/helper/retry/error.go.
//

type NotFoundError struct {
	LastError    error
	LastRequest  any
	LastResponse any
	Message      string
	Retries      int
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

type UnexpectedStateError[S comparable] struct {
	LastError     error
	State         S
	ExpectedState []S
}

func (e *UnexpectedStateError[S]) Error() string {
	return fmt.Sprintf(
		"unexpected state '%s', wanted target '%s'. last error: %s",
		e.State,
		strings.Join(toStrings(e.ExpectedState), ", "),
		e.LastError,
	)
}

func (e *UnexpectedStateError[S]) Unwrap() error {
	return e.LastError
}

type TimeoutError[S comparable] struct {
	LastError     error
	LastState     S
	Timeout       time.Duration
	ExpectedState []S
}

func (e *TimeoutError[S]) Error() string {
	expectedState := "resource to be gone"
	if len(e.ExpectedState) > 0 {
		expectedState = fmt.Sprintf("state to become '%s'", strings.Join(toStrings(e.ExpectedState), ", "))
	}

	var zero S
	extraInfo := make([]string, 0)
	if e.LastState != zero {
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

func (e *TimeoutError[S]) Unwrap() error {
	return e.LastError
}

func toStrings[S comparable](a []S) []string {
	return tfslices.ApplyToAll(a, func(s S) string {
		return fmt.Sprintf("%v", s)
	})
}
