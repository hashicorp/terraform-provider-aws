// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package json_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	tfjson "github.com/hashicorp/terraform-provider-aws/internal/json"
)

func TestCompactString(t *testing.T) {
	testCases := []struct {
		testName   string
		input      string
		wantOutput string
		wantErr    bool
	}{
		{
			testName: "empty string",
			input:    ` `,
			wantErr:  true,
		},
		{
			testName:   "empty JSON",
			input:      ` {       } `,
			wantOutput: `{}`,
		},
		{
			testName: "some JSON",
			input: `
			{
			"name": "example",
			"value": 123
			}
			`,
			wantOutput: `{"name":"example","value":123}`,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			t.Parallel()

			output, err := tfjson.CompactString(testCase.input)
			if got, want := err != nil, testCase.wantErr; !cmp.Equal(got, want) {
				t.Errorf("CompactString(%q) err %t, want %t", testCase.input, got, want)
			}
			if err == nil {
				if diff := cmp.Diff(output, testCase.wantOutput); diff != "" {
					t.Errorf("unexpected diff (+wanted, -got): %s", diff)
				}
			}
		})
	}
}
