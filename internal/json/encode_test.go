// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package json_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest/jsoncmp"
	"github.com/hashicorp/terraform-provider-aws/internal/json"
)

func TestEncodeToString(t *testing.T) {
	t.Parallel()

	type from struct {
		A string
		B int
		C struct {
			A bool
		}
		D []string
	}
	var from0 from
	from1 := from{
		A: "test1",
		B: 42,
		C: struct {
			A bool
		}{A: true},
		D: []string{"test2", "test3"},
	}

	testCases := []struct {
		testName   string
		input      any
		wantOutput string
		wantErr    bool
	}{
		{
			testName:   "empty JSON",
			input:      &from0,
			wantOutput: `{"A": "", "B": 0, "C": {"A": false}, "D": null}`,
		},
		{
			testName:   "empty JSON",
			input:      &from1,
			wantOutput: `{"A": "test1", "D": ["test2", "test3"], "C": {"A": true}, "B": 42}`,
		},
	}

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.testName, func(t *testing.T) {
			t.Parallel()

			output, err := json.EncodeToString(testCase.input)
			if got, want := err != nil, testCase.wantErr; !cmp.Equal(got, want) {
				t.Errorf("EncodeToString(%v) err %t, want %t", testCase.input, got, want)
			}
			if err == nil {
				if diff := jsoncmp.Diff(output, testCase.wantOutput); diff != "" {
					t.Errorf("unexpected diff (+wanted, -got): %s", diff)
				}
			}
		})
	}
}
