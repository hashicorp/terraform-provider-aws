// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iter

import (
	"maps"
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestApplyToAll(t *testing.T) {
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

			iter := ApplyToAll(slices.Values(test.input), strings.ToUpper)

			got := slices.Collect(iter)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestApplyToAll2(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    map[string]int
		expected map[string]int64
	}
	tests := map[string]testCase{
		"three elements": {
			input:    map[string]int{"one": 1, "two": 2, "three": 3},
			expected: map[string]int64{"ONE": -1, "TWO": -2, "THREE": -3},
		},
		"one element": {
			input:    map[string]int{"Four": 4},
			expected: map[string]int64{"FOUR": -4},
		},
		"zero elements": {
			input:    map[string]int{},
			expected: map[string]int64{},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			iter := ApplyToAll2(maps.All(test.input), func(k string, v int) (string, int64) {
				return strings.ToUpper(k), -int64(v)
			})

			got := maps.Collect(iter)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}
