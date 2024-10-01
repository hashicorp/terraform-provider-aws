// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package slices

import (
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestReverse(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    []string
		expected []string
	}
	tests := map[string]testCase{
		"three elements": {
			input:    []string{"one", "two", "3"},
			expected: []string{"3", "two", "one"},
		},
		"two elements": {
			input:    []string{"aa", "bb"},
			expected: []string{"bb", "aa"},
		},
		"one element": {
			input:    []string{"1"},
			expected: []string{"1"},
		},
		"zero elements": {
			input:    []string{},
			expected: []string{},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := Reverse(test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestRemoveAll(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    []string
		remove   []string
		expected []string
	}
	tests := map[string]testCase{
		"two occurrences": {
			input:    []string{"one", "two", "one"},
			remove:   []string{"one"},
			expected: []string{"two"},
		},
		"one occurrences": {
			input:    []string{"one", "two"},
			remove:   []string{"one"},
			expected: []string{"two"},
		},
		"only occurrence": {
			input:    []string{"one"},
			remove:   []string{"one"},
			expected: []string{},
		},
		"no occurrences": {
			input:    []string{"two", "three", "four"},
			remove:   []string{"one"},
			expected: []string{"two", "three", "four"},
		},
		"zero elements": {
			input:    []string{},
			remove:   []string{"one"},
			expected: []string{},
		},
		"duplicate remove": {
			input:    []string{"one", "two", "one"},
			remove:   []string{"one", "one"},
			expected: []string{"two"},
		},
		"remove all": {
			input:    []string{"one", "two", "three", "two", "one"},
			remove:   []string{"two", "one", "one", "three"},
			expected: []string{},
		},
		"remove none": {
			input:    []string{"two", "three", "four"},
			remove:   []string{"six", "one"},
			expected: []string{"two", "three", "four"},
		},
		"remove two": {
			input:    []string{"one", "two", "three", "four", "five", "six"},
			remove:   []string{"six", "one"},
			expected: []string{"two", "three", "four", "five"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := RemoveAll(test.input, test.remove...)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

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
			expected: []string{},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := ApplyToAll(test.input, strings.ToUpper)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestFilter(t *testing.T) {
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
			expected: []string{},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := Filter(test.input, func(v string) bool {
				return strings.HasPrefix(v, "a")
			})

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestAppendUnique(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    []string
		append   []string
		expected []string
	}
	tests := map[string]testCase{
		"all nil": {},
		"all empty": {
			input:    []string{},
			append:   []string{},
			expected: []string{},
		},
		"append to nil": {
			append:   []string{"alpha", "beta", "alpha"},
			expected: []string{"alpha", "beta"},
		},
		"append nothing": {
			input:    []string{"alpha", "beta"},
			append:   []string{},
			expected: []string{"alpha", "beta"},
		},
		"append no unique": {
			input:    []string{"alpha", "beta", "gamma"},
			append:   []string{"beta", "gamma", "alpha"},
			expected: []string{"alpha", "beta", "gamma"},
		},
		"append one unique": {
			input:    []string{"alpha", "beta", "gamma"},
			append:   []string{"beta", "delta"},
			expected: []string{"alpha", "beta", "gamma", "delta"},
		},
		"append three unique": {
			input:    []string{"alpha", "beta", "gamma"},
			append:   []string{"delta", "gamma", "epsilon", "alpha", "epsilon", "zeta"},
			expected: []string{"alpha", "beta", "gamma", "delta", "epsilon", "zeta"},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := AppendUnique(test.input, test.append...)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestIndexOf(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    []any
		element  string
		expected int
	}
	tests := map[string]testCase{
		"index 0": {
			input:    []any{"one", 2.0, 3, "four"},
			element:  "one",
			expected: 0,
		},
		"index 3": {
			input:    []any{"one", 2.0, 3, "four"},
			element:  "four",
			expected: 3,
		},
		"index -1": {
			input:    []any{"one", 2.0, 3, "four"},
			element:  "3",
			expected: -1,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := IndexOf(test.input, test.element)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestRange(t *testing.T) {
	t.Parallel()

	type testCase struct {
		start, stop, step int
		expected          []int
	}
	tests := map[string]testCase{
		"0 step": {
			start:    0,
			stop:     10,
			step:     0,
			expected: nil,
		},
		"start == stop": {
			start:    0,
			stop:     0,
			step:     1,
			expected: nil,
		},
		"start == 0, step == 1": {
			start:    0,
			stop:     10,
			step:     1,
			expected: []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9},
		},
		"start == 1, step == 1": {
			start:    1,
			stop:     11,
			step:     1,
			expected: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
		},
		"start == 0, step == 5": {
			start:    0,
			stop:     30,
			step:     5,
			expected: []int{0, 5, 10, 15, 20, 25},
		},
		"start == 0, step == 11": {
			start:    0,
			stop:     30,
			step:     11,
			expected: []int{0, 11, 22},
		},
		"start == 0, step == -1": {
			start:    0,
			stop:     -10,
			step:     -1,
			expected: []int{0, -1, -2, -3, -4, -5, -6, -7, -8, -9},
		},
		"start == 0, stop = 5, step == -1": {
			start:    0,
			stop:     5,
			step:     -1,
			expected: nil,
		},
		"start == 0, stop = -5, step == 1": {
			start:    0,
			stop:     -5,
			step:     1,
			expected: nil,
		},
		"start == 1, step == -5": {
			start:    1,
			stop:     -30,
			step:     -5,
			expected: []int{1, -4, -9, -14, -19, -24, -29},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := Range(test.start, test.stop, test.step)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}
