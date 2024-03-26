// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestSet_Difference_strings(t *testing.T) {
	t.Parallel()

	type testCase struct {
		original Set[string]
		new      Set[string]
		expected Set[string]
	}
	tests := map[string]testCase{
		"nil": {
			original: nil,
			new:      nil,
			expected: nil,
		},
		"equal": {
			original: Set[string]{"one"},
			new:      Set[string]{"one"},
			expected: nil,
		},
		"difference": {
			original: Set[string]{"one", "two", "four"},
			new:      Set[string]{"one", "two", "three"},
			expected: Set[string]{"four"},
		},
		"difference_remove": {
			original: Set[string]{"one", "two"},
			new:      Set[string]{"one"},
			expected: Set[string]{"two"},
		},
		"difference_add": {
			original: Set[string]{"one"},
			new:      Set[string]{"one", "two"},
			expected: nil,
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := test.original.Difference(test.new)
			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}
