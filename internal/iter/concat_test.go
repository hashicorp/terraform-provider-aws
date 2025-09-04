// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iter

import (
	"iter"
	"slices"
	"testing"
)

func TestConcat(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    []iter.Seq[int]
		expected int
	}

	tests := map[string]testCase{
		"nil input": {},
		"1 input, 0 elements": {
			input: []iter.Seq[int]{slices.Values([]int{})},
		},
		"1 input, 2 elements": {
			input:    []iter.Seq[int]{slices.Values([]int{1, 2})},
			expected: 2,
		},
		"3 inputs": {
			input:    []iter.Seq[int]{slices.Values([]int{1, 2}), slices.Values([]int(nil)), slices.Values([]int{3, 4, 5})},
			expected: 5,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			output := Concat(test.input...)

			if got, want := len(slices.Collect(output)), test.expected; got != want {
				t.Errorf("Length of output = %v, want %v", got, want)
			}
		})
	}
}
