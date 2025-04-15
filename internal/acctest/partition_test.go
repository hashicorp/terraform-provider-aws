// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package acctest_test

import (
	"testing"

	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestIsIsolatedPartition(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		input    string
		expected bool
	}{
		{
			input:    endpoints.AwsPartitionID,
			expected: false,
		},
		{
			input:    endpoints.AwsCnPartitionID,
			expected: false,
		},
		{
			input:    endpoints.AwsUsGovPartitionID,
			expected: false,
		},
		{
			input:    endpoints.AwsIsoPartitionID,
			expected: true,
		},
		{
			input:    endpoints.AwsIsoBPartitionID,
			expected: true,
		},
		{
			input:    endpoints.AwsIsoEPartitionID,
			expected: true,
		},
		{
			input:    endpoints.AwsIsoFPartitionID,
			expected: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.input, func(t *testing.T) {
			t.Parallel()

			if got, want := acctest.IsIsolatedPartition(testCase.input), testCase.expected; got != want {
				t.Errorf("got: %#v, expected: %#v", got, want)
			}
		})
	}
}

func TestIsIsolatedRegion(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		input    string
		expected bool
	}{
		{
			input:    endpoints.UsEast1RegionID,
			expected: false,
		},
		{
			input:    endpoints.CnNorth1RegionID,
			expected: false,
		},
		{
			input:    endpoints.UsGovEast1RegionID,
			expected: false,
		},
		{
			input:    endpoints.UsIsoEast1RegionID,
			expected: true,
		},
		{
			input:    endpoints.UsIsobEast1RegionID,
			expected: true,
		},
		{
			input:    endpoints.EuIsoeWest1RegionID,
			expected: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.input, func(t *testing.T) {
			t.Parallel()

			if got, want := acctest.IsIsolatedRegion(testCase.input), testCase.expected; got != want {
				t.Errorf("got: %#v, expected: %#v", got, want)
			}
		})
	}
}

func TestIsStandardPartition(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		input    string
		expected bool
	}{
		{
			input:    endpoints.AwsPartitionID,
			expected: true,
		},
		{
			input:    endpoints.AwsCnPartitionID,
			expected: false,
		},
		{
			input:    endpoints.AwsUsGovPartitionID,
			expected: false,
		},
		{
			input:    endpoints.AwsIsoPartitionID,
			expected: false,
		},
		{
			input:    endpoints.AwsIsoBPartitionID,
			expected: false,
		},
		{
			input:    endpoints.AwsIsoEPartitionID,
			expected: false,
		},
		{
			input:    endpoints.AwsIsoFPartitionID,
			expected: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.input, func(t *testing.T) {
			t.Parallel()

			if got, want := acctest.IsStandardPartition(testCase.input), testCase.expected; got != want {
				t.Errorf("got: %#v, expected: %#v", got, want)
			}
		})
	}
}

func TestIsStandardRegion(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		input    string
		expected bool
	}{
		{
			input:    endpoints.UsEast1RegionID,
			expected: true,
		},
		{
			input:    endpoints.CnNorth1RegionID,
			expected: false,
		},
		{
			input:    endpoints.UsGovEast1RegionID,
			expected: false,
		},
		{
			input:    endpoints.UsIsoEast1RegionID,
			expected: false,
		},
		{
			input:    endpoints.UsIsobEast1RegionID,
			expected: false,
		},
		{
			input:    endpoints.EuIsoeWest1RegionID,
			expected: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.input, func(t *testing.T) {
			t.Parallel()

			if got, want := acctest.IsStandardRegion(testCase.input), testCase.expected; got != want {
				t.Errorf("got: %#v, expected: %#v", got, want)
			}
		})
	}
}
