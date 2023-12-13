// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfresource_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestNotFound(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Name     string
		Err      error
		Expected bool
	}{
		{
			Name: "nil error",
			Err:  nil,
		},
		{
			Name: "other error",
			Err:  errors.New("test"),
		},
		{
			Name:     "not found error",
			Err:      &retry.NotFoundError{LastError: errors.New("test")},
			Expected: true,
		},
		{
			Name: "wrapped other error",
			Err:  fmt.Errorf("test: %w", errors.New("test")),
		},
		{
			Name:     "wrapped not found error",
			Err:      fmt.Errorf("test: %w", &retry.NotFoundError{LastError: errors.New("test")}),
			Expected: true,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()

			got := tfresource.NotFound(testCase.Err)

			if got != testCase.Expected {
				t.Errorf("got %t, expected %t", got, testCase.Expected)
			}
		})
	}
}

func TestTimedOut(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Name     string
		Err      error
		Expected bool
	}{
		{
			Name: "nil error",
			Err:  nil,
		},
		{
			Name: "other error",
			Err:  errors.New("test"),
		},
		{
			Name:     "timeout error",
			Err:      &retry.TimeoutError{},
			Expected: true,
		},
		{
			Name: "timeout error non-nil last error",
			Err:  &retry.TimeoutError{LastError: errors.New("test")},
		},
		{
			Name: "wrapped other error",
			Err:  fmt.Errorf("test: %w", errors.New("test")),
		},
		{
			Name: "wrapped timeout error",
			Err:  fmt.Errorf("test: %w", &retry.TimeoutError{}),
		},
		{
			Name: "wrapped timeout error non-nil last error",
			Err:  fmt.Errorf("test: %w", &retry.TimeoutError{LastError: errors.New("test")}),
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()

			got := tfresource.TimedOut(testCase.Err)

			if got != testCase.Expected {
				t.Errorf("got %t, expected %t", got, testCase.Expected)
			}
		})
	}
}

func TestSetLastError(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		Name     string
		Err      error
		LastErr  error
		Expected bool
	}{
		{
			Name: "nil error",
		},
		{
			Name:    "other error",
			Err:     errors.New("test"),
			LastErr: errors.New("last"),
		},
		{
			Name: "timeout error lastErr is nil",
			Err:  &retry.TimeoutError{},
		},
		{
			Name:     "timeout error",
			Err:      &retry.TimeoutError{},
			LastErr:  errors.New("lasttest"),
			Expected: true,
		},
		{
			Name: "timeout error non-nil last error lastErr is nil",
			Err:  &retry.TimeoutError{LastError: errors.New("test")},
		},
		{
			Name:    "timeout error non-nil last error no overwrite",
			Err:     &retry.TimeoutError{LastError: errors.New("test")},
			LastErr: errors.New("lasttest"),
		},
		{
			Name: "unexpected state error lastErr is nil",
			Err:  &retry.UnexpectedStateError{},
		},
		{
			Name:     "unexpected state error",
			Err:      &retry.UnexpectedStateError{},
			LastErr:  errors.New("lasttest"),
			Expected: true,
		},
		{
			Name: "unexpected state error non-nil last error lastErr is nil",
			Err:  &retry.UnexpectedStateError{LastError: errors.New("test")},
		},
		{
			Name:    "unexpected state error non-nil last error no overwrite",
			Err:     &retry.UnexpectedStateError{LastError: errors.New("test")},
			LastErr: errors.New("lasttest"),
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()

			tfresource.SetLastError(testCase.Err, testCase.LastErr)

			if testCase.Err != nil {
				got := testCase.Err.Error()
				contains := strings.Contains(got, "lasttest")

				if (testCase.Expected && !contains) || (!testCase.Expected && contains) {
					t.Errorf("got %s", got)
				}
			}
		})
	}
}
