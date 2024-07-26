// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tfresource

import (
	"errors"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func TestEmptyResultErrorAsNotFoundError(t *testing.T) {
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
	if nfe.LastRequest != lastRequest {
		t.Errorf("unexpected value for LastRequest")
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
			err: &EmptyResultError{
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
			err: fmt.Errorf("test: %w", &EmptyResultError{
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
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			err := &EmptyResultError{}
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

	var nfe *retry.NotFoundError
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
		testCase := testCase
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

func TestAssertSinglePtrResult(t *testing.T) {
	t.Parallel()

	type x struct {
		A int
	}

	x1 := &x{A: 1}
	x2 := &x{A: 2}
	testCases := []struct {
		name   string
		input  []*x
		fs     []foundFunc[x]
		output *x
		err    bool
	}{
		{
			name: "nil input",
			err:  true,
		},
		{
			name:  "empty input",
			input: []*x{},
			err:   true,
		},
		{
			name:  "single nil input",
			input: []*x{nil},
			err:   true,
		},
		{
			name:   "single non-nil input",
			input:  []*x{x1},
			output: x1,
		},
		{
			name:  "multiple inputs",
			input: []*x{x1, x2},
			err:   true,
		},
		{
			name:   "single non-nil input, with found",
			input:  []*x{x1},
			fs:     []foundFunc[x]{func(v *x) bool { return v.A == 1 }},
			output: x1,
		},
		{
			name:  "single non-nil input, with not found",
			input: []*x{x1},
			fs:    []foundFunc[x]{func(v *x) bool { return v.A == 2 }},
			err:   true,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			output, err := AssertSinglePtrResult(testCase.input, testCase.fs...)
			if got, want := err != nil, testCase.err; got != want {
				t.Fatalf("AssertSinglePtrResult err %t, want %t", got, want)
			}
			if diff := cmp.Diff(output, testCase.output); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}
