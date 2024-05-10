// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package slices

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestPredicateAnd(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input      int
		predicates []Predicate[int]
		expected   bool
	}
	tests := map[string]testCase{
		"all true": {
			input: 7,
			predicates: []Predicate[int]{
				PredicateEquals(7),
				PredicateTrue[int](),
			},
			expected: true,
		},
		"one false": {
			input: 7,
			predicates: []Predicate[int]{
				PredicateTrue[int](),
				PredicateEquals(7),
				PredicateEquals(6),
			},
			expected: false,
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := PredicateAnd(test.predicates...)(test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestPredicateOr(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input      int
		predicates []Predicate[int]
		expected   bool
	}
	tests := map[string]testCase{
		"all true": {
			input: 7,
			predicates: []Predicate[int]{
				PredicateEquals(7),
				PredicateTrue[int](),
			},
			expected: true,
		},
		"one false": {
			input: 7,
			predicates: []Predicate[int]{
				PredicateTrue[int](),
				PredicateEquals(7),
				PredicateEquals(6),
			},
			expected: true,
		},
		"all false": {
			input: 7,
			predicates: []Predicate[int]{
				PredicateEquals(6),
				PredicateEquals(5),
			},
			expected: false,
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := PredicateOr(test.predicates...)(test.input)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}
