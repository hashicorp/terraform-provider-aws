// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iter

import (
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

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
			expected: nil,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			iter := Filter(slices.Values(test.input), func(v string) bool {
				return strings.HasPrefix(v, "a")
			})

			got := slices.Collect(iter)

			if diff := cmp.Diff(got, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestFilterValues(t *testing.T) {
	t.Parallel()

	errBoom := errors.New("boom")
	seq := func(yield func(string, error) bool) {
		for _, p := range []struct {
			v   string
			err error
		}{{"keep", nil}, {"drop", nil}, {"", errBoom}, {"keep2", nil}} {
			if !yield(p.v, p.err) {
				return
			}
		}
	}

	var values []string
	var errs []error
	// The predicate rejects everything it is shown, so anything that survives
	// got through on the error path.
	for v, err := range FilterValues(seq, func(v string) bool { return v != "drop" }) {
		values = append(values, v)
		if err != nil {
			errs = append(errs, err)
		}
	}

	if want := []string{"keep", "", "keep2"}; !slices.Equal(values, want) {
		t.Errorf("got values %v, want %v", values, want)
	}
	if len(errs) != 1 || !errors.Is(errs[0], errBoom) {
		t.Errorf("got errs %v, want one errBoom", errs)
	}
}

func TestFilterValuesStopsEarly(t *testing.T) {
	t.Parallel()

	produced := 0
	seq := func(yield func(int, error) bool) {
		for i := range 10 {
			produced++
			if !yield(i, nil) {
				return
			}
		}
	}

	for _, err := range FilterValues(seq, func(int) bool { return true }) {
		if err != nil {
			t.Fatal(err)
		}

		break
	}

	if produced != 1 {
		t.Errorf("iterator produced %d values after the consumer stopped, want 1", produced)
	}
}
