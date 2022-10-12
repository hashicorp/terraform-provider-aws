package slices

import (
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
			got := Reverse(test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestReversed(t *testing.T) {
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
			got := Reversed(test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}
