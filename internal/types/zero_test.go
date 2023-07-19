// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"testing"
)

type AIsZero struct {
	Key   string
	Value int
}

func TestIsZero(t *testing.T) {
	t.Parallel()

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
			Ptr:      &AIsZero{},
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
		testCase := testCase

		t.Run(testCase.Name, func(t *testing.T) {
			t.Parallel()

			got := IsZero(testCase.Ptr)

			if got != testCase.Expected {
				t.Errorf("got %t, expected %t", got, testCase.Expected)
			}
		})
	}
}
