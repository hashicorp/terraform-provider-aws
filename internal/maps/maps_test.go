// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package maps

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestApplyToAll(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    map[int]string
		expected map[int]string
	}
	tests := map[string]testCase{
		"three elements": {
			input: map[int]string{
				1: "one",
				2: "two",
				3: "3"},
			expected: map[int]string{
				1: "ONE",
				2: "TWO",
				3: "3"},
		},
		"one element": {
			input: map[int]string{
				123: "abcdEFGH"},
			expected: map[int]string{
				123: "ABCDEFGH"},
		},
		"zero elements": {
			input:    map[int]string{},
			expected: map[int]string{},
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := ApplyToAll(test.input, strings.ToUpper)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}
