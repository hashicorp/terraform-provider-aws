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
		name, test := name, test
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
		expected []string
	}
	tests := map[string]testCase{
		"two occurrences": {
			input:    []string{"one", "two", "one"},
			expected: []string{"two"},
		},
		"one occurrences": {
			input:    []string{"one", "two"},
			expected: []string{"two"},
		},
		"only occurrence": {
			input:    []string{"one"},
			expected: []string{},
		},
		"no occurrences": {
			input:    []string{"two", "three", "four"},
			expected: []string{"two", "three", "four"},
		},
		"zero elements": {
			input:    []string{},
			expected: []string{},
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := RemoveAll(test.input, "one")

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

func TestChunk(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input    []string
		expected [][]string
	}
	tests := map[string]testCase{
		"three elements": {
			input:    []string{"one", "two", "3"},
			expected: [][]string{{"one", "two"}, {"3"}},
		},
		"two elements": {
			input:    []string{"aa", "bb"},
			expected: [][]string{{"aa", "bb"}},
		},
		"one element": {
			input:    []string{"1"},
			expected: [][]string{{"1"}},
		},
		"zero elements": {
			input:    []string{},
			expected: [][]string{},
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := Chunks(test.input, 2)

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
		name, test := name, test
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
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := AppendUnique(test.input, test.append...)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}
