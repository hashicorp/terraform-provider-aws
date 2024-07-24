// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package acctest_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestIsIsolatedPartition(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		input    string
		expected bool
	}{
		{
			input:    names.StandardPartitionID,
			expected: false,
		},
		{
			input:    names.ChinaPartitionID,
			expected: false,
		},
		{
			input:    names.USGovCloudPartitionID,
			expected: false,
		},
		{
			input:    names.ISOPartitionID,
			expected: true,
		},
		{
			input:    names.ISOBPartitionID,
			expected: true,
		},
		{
			input:    names.ISOEPartitionID,
			expected: true,
		},
		{
			input:    names.ISOFPartitionID,
			expected: true,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
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
			input:    names.USEast1RegionID,
			expected: false,
		},
		{
			input:    names.CNNorth1RegionID,
			expected: false,
		},
		{
			input:    names.USGovEast1RegionID,
			expected: false,
		},
		{
			input:    names.USISOEast1RegionID,
			expected: true,
		},
		{
			input:    names.USISOBEast1RegionID,
			expected: true,
		},
		{
			input:    names.EUISOEWest1RegionID,
			expected: true,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
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
			input:    names.StandardPartitionID,
			expected: true,
		},
		{
			input:    names.ChinaPartitionID,
			expected: false,
		},
		{
			input:    names.USGovCloudPartitionID,
			expected: false,
		},
		{
			input:    names.ISOPartitionID,
			expected: false,
		},
		{
			input:    names.ISOBPartitionID,
			expected: false,
		},
		{
			input:    names.ISOEPartitionID,
			expected: false,
		},
		{
			input:    names.ISOFPartitionID,
			expected: false,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
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
			input:    names.USEast1RegionID,
			expected: true,
		},
		{
			input:    names.CNNorth1RegionID,
			expected: false,
		},
		{
			input:    names.USGovEast1RegionID,
			expected: false,
		},
		{
			input:    names.USISOEast1RegionID,
			expected: false,
		},
		{
			input:    names.USISOBEast1RegionID,
			expected: false,
		},
		{
			input:    names.EUISOEWest1RegionID,
			expected: false,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.input, func(t *testing.T) {
			t.Parallel()

			if got, want := acctest.IsStandardRegion(testCase.input), testCase.expected; got != want {
				t.Errorf("got: %#v, expected: %#v", got, want)
			}
		})
	}
}
