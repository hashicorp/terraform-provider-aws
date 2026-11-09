// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iter

import (
	"errors"
	"iter"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
)

func TestConcat(t *testing.T) {
	t.Parallel()

	type testCase struct {
		input   []iter.Seq[int]
		wantLen int
	}
	tests := map[string]testCase{
		"nil input": {},
		"1 input, 0 elements": {
			input: []iter.Seq[int]{slices.Values([]int{})},
		},
		"1 input, 2 elements": {
			input:   []iter.Seq[int]{slices.Values([]int{1, 2})},
			wantLen: 2,
		},
		"3 inputs": {
			input:   []iter.Seq[int]{slices.Values([]int{1, 2}), slices.Values([]int(nil)), slices.Values([]int{3, 4, 5})},
			wantLen: 5,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			output := Concat(test.input...)

			if got, want := len(slices.Collect(output)), test.wantLen; !cmp.Equal(got, want) {
				t.Errorf("Length of output = %d, want %d", got, want)
			}
		})
	}
}

func TestConcatValuesWithError(t *testing.T) {
	t.Parallel()

	noError := func(yield func([]int, error) bool) {
		if !yield([]int{1, 2, 3}, nil) {
			return
		}
		if !yield([]int{4, 5}, nil) {
			return
		}
	}
	hasError := func(yield func([]int, error) bool) {
		if !yield([]int{1, 2, 3}, nil) {
			return
		}
		if !yield(nil, errors.New("test error")) {
			return
		}
		if !yield([]int{4, 5}, nil) {
			return
		}
	}

	type testCase struct {
		input   iter.Seq2[[]int, error]
		wantErr bool
		wantLen int
	}
	tests := map[string]testCase{
		"nul input": {
			input: Null2[[]int, error](),
		},
		"no error": {
			input:   noError,
			wantLen: 5,
		},
		"has error": {
			input:   hasError,
			wantErr: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := tfslices.CollectWithError(ConcatValuesWithError(test.input))

			if got, want := err != nil, test.wantErr; !cmp.Equal(got, want) {
				t.Errorf("ConcatValuesWithError() err %t, want %t", got, want)
			}
			if err == nil {
				if got, want := len(got), test.wantLen; !cmp.Equal(got, want) {
					t.Errorf("ConcatValuesWithError() len %d, want %d", got, want)
				}
			}
		})
	}
}
