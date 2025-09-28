// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3control

import "testing"

func TestIsDirectoryBucketARN(t *testing.T) {
	testCases := []struct {
		name     string
		arn      string
		expected bool
	}{
		{
			name:     "Standard S3 bucket ARN",
			arn:      "arn:aws:s3:::my-bucket",
			expected: false,
		},
		{
			name:     "S3 Control Access Point ARN",
			arn:      "arn:aws:s3:us-east-1:123456789012:accesspoint/my-access-point",
			expected: false,
		},
		{
			name:     "Directory Bucket ARN (S3 Express)",
			arn:      "arn:aws:s3express:us-east-1:123456789012:bucket/my-directory-bucket--usw2-az1--x-s3",
			expected: true,
		},
		{
			name:     "Directory Bucket Access Point ARN",
			arn:      "arn:aws:s3express:us-east-1:123456789012:accesspoint/my-access-point",
			expected: true,
		},
		{
			name:     "Empty ARN",
			arn:      "",
			expected: false,
		},
		{
			name:     "Invalid ARN format",
			arn:      "not-an-arn",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isDirectoryBucketARN(tc.arn)
			if result != tc.expected {
				t.Errorf("isDirectoryBucketARN(%q) = %v, expected %v", tc.arn, result, tc.expected)
			}
		})
	}
}
