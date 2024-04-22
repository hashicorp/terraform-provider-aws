// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func TestSetDifference(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name     string
		set1     Set[int]
		set2     Set[int]
		expected Set[int]
	}{
		{
			name:     "nil sets",
			expected: nil,
		},
		{
			name:     "empty sets",
			set1:     Set[int]{},
			set2:     Set[int]{},
			expected: nil,
		},
		{
			name:     "no overlap",
			set1:     Set[int]{1, 3, 5, 7},
			set2:     Set[int]{2, 4, 6, 8},
			expected: Set[int]{1, 3, 5, 7},
		},
		{
			name:     "no overlap swapped",
			set1:     Set[int]{2, 4, 6, 8},
			set2:     Set[int]{1, 3, 5, 7},
			expected: Set[int]{2, 4, 6, 8},
		},
		{
			name:     "overlap",
			set1:     Set[int]{1, 2, 3, 4, 5, 7},
			set2:     Set[int]{1, 2, 4, 6, 8},
			expected: Set[int]{3, 5, 7},
		},
	}

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			got := testCase.set1.Difference(testCase.set2)

			if diff := cmp.Diff(got, testCase.expected, cmpopts.SortSlices(func(v1, v2 int) bool { return v1 < v2 })); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}
