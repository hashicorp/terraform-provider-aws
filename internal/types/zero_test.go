// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"testing"
)

type AIsZero struct {
	Key   string
	Value int
}

func TestIsZeroPtr(t *testing.T) {
	t.Parallel()

	zero := Zero[AIsZero]()
	testCases := []struct {
		Name     string
		Ptr      *AIsZero
		Expected bool
	}{
		{
			Name:     "nil pointer",
			Expected: true,
		},
		{
			Name:     "pointer to zero value",
			Ptr:      &zero,
			Expected: true,
		},
		{
			Name: "pointer to non-zero value Key",
			Ptr:  &AIsZero{Key: "test"},
		},
		{
			Name: "pointer to non-zero value Value",
			Ptr:  &AIsZero{Value: 42},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()

			got := IsZero(testCase.Ptr)

			if got != testCase.Expected {
				t.Errorf("got %t, expected %t", got, testCase.Expected)
			}
		})
	}
}

func TestIsZeroStructValue(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		Value    AIsZero
		Expected bool
	}{
		"zero value struct": {
			Value:    AIsZero{},
			Expected: true,
		},
		"non-zero value struct": {
			Value:    AIsZero{Value: 42},
			Expected: false,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := IsZero(testCase.Value)

			if got != testCase.Expected {
				t.Errorf("got %t, expected %t", got, testCase.Expected)
			}
		})
	}
}

func TestIsZeroPointerToAny(t *testing.T) {
	t.Parallel()

	var (
		a, b, c any
	)
	b = 0
	c = 42
	testCases := []struct {
		Name     string
		Ptr      *any
		Expected bool
	}{
		{
			Name:     "nil pointer",
			Expected: true,
		},
		{
			Name:     "pointer to nil",
			Ptr:      &a,
			Expected: true,
		},
		{
			Name:     "pointer to zero value",
			Ptr:      &b,
			Expected: true,
		},
		{
			Name: "pointer to non-zero value",
			Ptr:  &c,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()

			got := IsZero(testCase.Ptr)

			if got != testCase.Expected {
				t.Errorf("got %t, expected %t", got, testCase.Expected)
			}
		})
	}
}

func TestIsZeroAnyValue(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		Value    any
		Expected bool
	}{
		"nil value": {
			Expected: true,
		},
		"zero int value": {
			Value:    0,
			Expected: true,
		},
		"zero string value": {
			Value:    "",
			Expected: true,
		},
		"non-zero int value": {
			Value:    1,
			Expected: false,
		},
		"non-zero string value": {
			Value:    "string",
			Expected: false,
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := IsZero(testCase.Value)

			if got != testCase.Expected {
				t.Errorf("got %t, expected %t", got, testCase.Expected)
			}
		})
	}
}
