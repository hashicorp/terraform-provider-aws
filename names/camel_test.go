// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package names

import (
	"testing"
)

func TestToCamelCase(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty",
			input:    "",
			expected: "",
		},
		{
			name:     "All lower",
			input:    "lower",
			expected: "Lower",
		},
		{
			name:     "Initial upper",
			input:    "Lower",
			expected: "Lower",
		},
		{
			name:     "Two words",
			input:    "Two words",
			expected: "TwoWords",
		},
		{
			name:     "Three words",
			input:    "TheseThreeWords",
			expected: "TheseThreeWords",
		},
		{
			name:     "A long one",
			input:    "global-Replication_Group.description",
			expected: "GlobalReplicationGroupDescription",
		},
		{
			name:     "Including a digit",
			input:    "s3bucket",
			expected: "S3Bucket",
		},
		{
			name:     "Constant case",
			input:    "CONSTANT_CASE",
			expected: "ConstantCase",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if got, want := ToCamelCase(testCase.input), testCase.expected; got != want {
				t.Errorf("got: %s, expected: %s", got, want)
			}
		})
	}
}

func TestToLowerCamelCase(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty",
			input:    "",
			expected: "",
		},
		{
			name:     "All lower",
			input:    "lower",
			expected: "lower",
		},
		{
			name:     "Initial upper",
			input:    "Lower",
			expected: "lower",
		},
		{
			name:     "Two words",
			input:    "Two words",
			expected: "twoWords",
		},
		{
			name:     "Three words",
			input:    "TheseThreeWords",
			expected: "theseThreeWords",
		},
		{
			name:     "A long one",
			input:    "global-Replication_Group.description",
			expected: "globalReplicationGroupDescription",
		},
		{
			name:     "Including a digit",
			input:    "s3bucket",
			expected: "s3Bucket",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if got, want := ToLowerCamelCase(testCase.input), testCase.expected; got != want {
				t.Errorf("got: %s, expected: %s", got, want)
			}
		})
	}
}
