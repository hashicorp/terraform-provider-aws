// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package names

import (
	"testing"
)

func TestToSnakeCase(t *testing.T) {
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
			input:    "TwoWords",
			expected: "two_words",
		},
		{
			name:     "A long one",
			input:    "GlobalReplicationGroupDescription",
			expected: "global_replication_group_description",
		},
		{
			name:     "Including a digit",
			input:    "S3Bucket",
			expected: "s3_bucket",
		},
		{
			name:     "ARN",
			input:    "ARN",
			expected: "arn",
		},
		{
			name:     "ResourceArn",
			input:    "ResourceArn",
			expected: "resource_arn",
		},
		{
			name:     "ResourceARN",
			input:    "ResourceARN",
			expected: "resource_arn",
		},
		{
			name:     "Resource-ARN",
			input:    "Resource-ARN",
			expected: "resource_arn",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			if got, want := ToSnakeCase(testCase.input), testCase.expected; got != want {
				t.Errorf("got: %s, expected: %s", got, want)
			}
		})
	}
}
