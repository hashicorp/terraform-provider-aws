// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package smithy_test

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cognitoidentityprovider/document"
	"github.com/google/go-cmp/cmp"
	tfsmithy "github.com/hashicorp/terraform-provider-aws/internal/smithy"
)

func TestDocumentToFromJSONString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		testName      string
		input         string
		wantOutput    string
		wantInputErr  bool
		wantOutputErr bool
	}{
		{
			testName:   "empty string",
			input:      ``,
			wantOutput: `null`,
		},
		{
			testName:   "empty JSON",
			input:      `{}`,
			wantOutput: `{}`,
		},
		{
			testName:   "valid JSON",
			input:      `{"Field1": 42}`,
			wantOutput: `{"Field1":42}`,
		},
		{
			testName:     "invalid JSON",
			input:        `{"Field1"=42}`,
			wantInputErr: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			t.Parallel()

			json, err := tfsmithy.DocumentFromJSONString(testCase.input, document.NewLazyDocument)
			if got, want := err != nil, testCase.wantInputErr; !cmp.Equal(got, want) {
				t.Errorf("DocumentFromJSONString(%s) err %t (%v), want %t", testCase.input, got, err, want)
			}
			if err == nil {
				output, err := tfsmithy.DocumentToJSONString(json)
				if got, want := err != nil, testCase.wantOutputErr; !cmp.Equal(got, want) {
					t.Errorf("DocumentToJSONString err %t (%v), want %t", got, err, want)
				}
				if err == nil {
					if diff := cmp.Diff(output, testCase.wantOutput); diff != "" {
						t.Errorf("unexpected diff (+wanted, -got): %s", diff)
					}
				}
			}
		})
	}
}
