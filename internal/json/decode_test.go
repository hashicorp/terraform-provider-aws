// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package json_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-provider-aws/internal/json"
)

func TestDecodeFromString(t *testing.T) {
	t.Parallel()

	type to struct {
		A string
		B int
		C struct {
			A bool
		}
		D []string
	}
	var to0, to1, to2, to3 to
	to4 := to{
		A: "test1",
		B: 42,
		C: struct {
			A bool
		}{A: true},
		D: []string{"test2", "test3"},
	}

	testCases := []struct {
		testName   string
		input      string
		output     any
		wantOutput any
		wantErr    bool
	}{
		{
			testName:   "empty JSON",
			input:      `{}`,
			output:     &to1,
			wantOutput: &to0,
		},
		{
			testName: "bad JSON",
			input:    `{test`,
			output:   &to2,
			wantErr:  true,
		},
		{
			testName:   "full JSON",
			input:      `{"A": "test1", "D": ["test2", "test3"], "C": {"A": true}, "B": 42}`,
			output:     &to3,
			wantOutput: &to4,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.testName, func(t *testing.T) {
			t.Parallel()

			err := json.DecodeFromString(testCase.input, testCase.output)
			if got, want := err != nil, testCase.wantErr; !cmp.Equal(got, want) {
				t.Errorf("DecodeFromString(%s) err %t, want %t", testCase.input, got, want)
			}
			if err == nil {
				if diff := cmp.Diff(testCase.output, testCase.wantOutput); diff != "" {
					t.Errorf("unexpected diff (+wanted, -got): %s", diff)
				}
			}
		})
	}
}
