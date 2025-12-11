// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package iter

import (
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestAppliedToEach(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    []string
		expected []string
	}
	tests := map[string]testCase{
		"three elements": {
			input:    []string{"one", "two", "3"},
			expected: []string{"ONE", "TWO", "3"},
		},
		"one element": {
			input:    []string{"abcdEFGH"},
			expected: []string{"ABCDEFGH"},
		},
		"zero elements": {
			input:    []string{},
			expected: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			iter := AppliedToEach(slices.Values(test.input), strings.ToUpper)

			got := slices.Collect(iter)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}
