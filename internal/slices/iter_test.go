// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package slices

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

			iter := AppliedToEach(test.input, strings.ToUpper)

			got := slices.Collect(iter)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

// Copied and adapted from stdlib slices package
func TestBackwardValues(t *testing.T) {
	t.Parallel()

	for size := range 10 {
		var s []int
		for i := range size {
			s = append(s, i)
		}
		ev := size - 1
		cnt := 0
		for v := range BackwardValues(s) {
			if v != ev {
				t.Errorf("at iteration %d got  %d want %d", cnt, v, ev)
			}
			ev--
			cnt++
		}
		if cnt != size {
			t.Errorf("read %d values expected %d", cnt, size)
		}
	}
}
