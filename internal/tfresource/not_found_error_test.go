// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package tfresource

import (
	"errors"
	"fmt"
	"iter"
	"testing"

	sdkretry "github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	tfiter "github.com/hashicorp/terraform-provider-aws/internal/iter"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
)

func TestEmptyResultErrorAsSdkNotFoundError(t *testing.T) {
	t.Parallel()

	lastRequest := 123
	err := NewEmptyResultError(lastRequest)

	var nfe *sdkretry.NotFoundError
	ok := errors.As(err, &nfe)

	if !ok {
		t.Fatal("expected errors.As() to return true")
	}
	if nfe.Message != "empty result" {
		t.Errorf(`expected Message to be "empty result", got %q`, nfe.Message)
	}
	if nfe.LastRequest != lastRequest {
		t.Errorf("unexpected value for LastRequest")
	}
}

func TestEmptyResultErrorAsRetryNotFoundError(t *testing.T) {
	t.Parallel()

	lastRequest := 123
	err := NewEmptyResultError(lastRequest)

	var nfe *retry.NotFoundError
	ok := errors.As(err, &nfe)

	if !ok {
		t.Fatal("expected errors.As() to return true")
	}
	if nfe.Message != "empty result" {
		t.Errorf(`expected Message to be "empty result", got %q`, nfe.Message)
	}
}

func TestEmptyResultErrorErrorsIs(t *testing.T) {
	t.Parallel()

	if !errors.Is(&emptyResultError{}, ErrEmptyResult) {
		t.Error("Expected `errors.Is` to match EmptyResultError")
	}
}

func TestEmptyResultErrorIs(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name: "compare to nil",
			err:  nil,
		},
		{
			name: "other error",
			err:  errors.New("test"),
		},
		{
			name: "EmptyResultError with LastRequest",
			err: &emptyResultError{
				LastRequest: 123,
			},
			expected: true,
		},
		{
			name:     "ErrEmptyResult",
			err:      ErrEmptyResult,
			expected: true,
		},
		{
			name: "wrapped other error",
			err:  fmt.Errorf("test: %w", errors.New("test")),
		},
		{
			name: "wrapped EmptyResultError with LastRequest",
			err: fmt.Errorf("test: %w", &emptyResultError{
				LastRequest: 123,
			}),
			expected: true,
		},
		{
			name:     "wrapped ErrEmptyResult",
			err:      fmt.Errorf("test: %w", ErrEmptyResult),
			expected: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := &emptyResultError{}
			ok := errors.Is(testCase.err, err)
			if ok != testCase.expected {
				t.Errorf("got %t, expected %t", ok, testCase.expected)
			}
		})
	}
}

func TestTooManyResultsErrorAsNotFoundError(t *testing.T) {
	t.Parallel()

	count := 2
	lastRequest := 123
	err := NewTooManyResultsError(count, lastRequest)

	var nfe *sdkretry.NotFoundError
	ok := errors.As(err, &nfe)

	if !ok {
		t.Fatal("expected errors.As() to return true")
	}
	if expected := fmt.Sprintf("too many results: wanted 1, got %d", count); nfe.Message != expected {
		t.Errorf(`expected Message to be %q, got %q`, expected, nfe.Message)
	}
	if nfe.LastRequest != lastRequest {
		t.Errorf("unexpected value for LastRequest")
	}
}

func TestTooManyResultsErrorIs(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name: "compare to nil",
			err:  nil,
		},
		{
			name: "other error",
			err:  errors.New("test"),
		},
		{
			name: "TooManyResultsError with LastRequest",
			err: &TooManyResultsError{
				LastRequest: 123,
			},
			expected: true,
		},
		{
			name:     "ErrTooManyResults",
			err:      ErrTooManyResults,
			expected: true,
		},
		{
			name: "wrapped other error",
			err:  fmt.Errorf("test: %w", errors.New("test")),
		},
		{
			name: "wrapped TooManyResultsError with LastRequest",
			err: fmt.Errorf("test: %w", &TooManyResultsError{
				LastRequest: 123,
			}),
			expected: true,
		},
		{
			name:     "wrapped ErrTooManyResults",
			err:      fmt.Errorf("test: %w", ErrTooManyResults),
			expected: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := &TooManyResultsError{}
			ok := errors.Is(testCase.err, err)
			if ok != testCase.expected {
				t.Errorf("got %t, expected %t", ok, testCase.expected)
			}
		})
	}
}

func TestAssertSingleValueResult(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input         []int
		expectedValue int
		expectedError error
	}{
		"empty slice": {
			input:         []int{},
			expectedError: NewEmptyResultError(nil),
		},
		"single element": {
			input:         []int{42},
			expectedValue: 42,
		},
		"multiple elements": {
			input:         []int{42, 43},
			expectedError: NewTooManyResultsError(2, nil),
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			result, err := AssertSingleValueResult(testCase.input)

			if testCase.expectedError != nil {
				if err == nil {
					t.Errorf("expected error: %v, got nil", testCase.expectedError)
				} else if err.Error() != testCase.expectedError.Error() {
					t.Errorf("expected error: %v, got %v", testCase.expectedError, err)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if result == nil {
				if testCase.expectedError == nil {
					t.Errorf("expected %d, got nil", testCase.expectedValue)
				}
				return
			} else if *result != testCase.expectedValue {
				t.Errorf("expected %d, got %d", testCase.expectedValue, *result)
			}
		})
	}
}

func TestAssertSingleValueResultIterErr(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		input         iter.Seq2[int, error]
		expectedValue int
		expectedError error
	}{
		"empty slice": {
			input:         tfiter.Null2[int, error](),
			expectedError: NewEmptyResultError(nil),
		},
		"single element": {
			input:         valuesWithErrors([]int{42}),
			expectedValue: 42,
		},
		"multiple elements": {
			input:         valuesWithErrors([]int{42, 43}),
			expectedError: NewTooManyResultsError(2, nil),
		},
		"with error": {
			input:         valueError(errors.New("test error")),
			expectedError: errors.New("test error"),
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			result, err := AssertSingleValueResultIterErr(testCase.input)

			if testCase.expectedError != nil {
				if err == nil {
					t.Errorf("expected error: %v, got nil", testCase.expectedError)
				} else if err.Error() != testCase.expectedError.Error() {
					t.Errorf("expected error: %v, got %v", testCase.expectedError, err)
				}
			} else if err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if result == nil {
				if testCase.expectedError == nil {
					t.Errorf("expected %d, got nil", testCase.expectedValue)
				}
				return
			} else if *result != testCase.expectedValue {
				t.Errorf("expected %d, got %d", testCase.expectedValue, *result)
			}
		})
	}
}

func valuesWithErrors(values []int) iter.Seq2[int, error] {
	return func(yield func(int, error) bool) {
		for _, v := range values {
			if !yield(v, nil) {
				break
			}
		}
	}
}

func valueError(err error) iter.Seq2[int, error] {
	return func(yield func(int, error) bool) {
		if !yield(0, err) {
			return
		}
	}
}
