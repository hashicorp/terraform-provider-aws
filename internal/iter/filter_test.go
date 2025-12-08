// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iter

import (
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestFiltered(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    []string
		expected []string
	}
	tests := map[string]testCase{
		"three elements": {
			input:    []string{"one", "two", "3", "a0"},
			expected: []string{"a0"},
		},
		"one element": {
			input:    []string{"abcdEFGH"},
			expected: []string{"abcdEFGH"},
		},
		"zero elements": {
			input:    []string{},
			expected: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			iter := Filtered(slices.Values(test.input), func(v string) bool {
				return strings.HasPrefix(v, "a")
			})

			got := slices.Collect(iter)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}
