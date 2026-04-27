// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package stringvalidator_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	tfstringvalidator "github.com/hashicorp/terraform-provider-aws/internal/framework/validators/stringvalidator"
)

func TestPrefixNoneOfCaseInsensitive(t *testing.T) {
	t.Parallel()

	type testCase struct {
		in          types.String
		prefixes    []string
		exceptions  []string
		expectError bool
	}

	testCases := map[string]testCase{
		"valid-no-match": {
			in:          types.StringValue("x-custom-header"),
			prefixes:    []string{"X-Amzn-"},
			expectError: false,
		},
		"invalid-prefix-match": {
			in:          types.StringValue("X-Amzn-Custom"),
			prefixes:    []string{"X-Amzn-"},
			expectError: true,
		},
		"invalid-case-insensitive": {
			in:          types.StringValue("x-amzn-custom"),
			prefixes:    []string{"X-Amzn-"},
			expectError: true,
		},
		"valid-exception": {
			in:          types.StringValue("X-Amzn-Bedrock-AgentCore-Runtime-Custom-MyHeader"),
			prefixes:    []string{"X-Amzn-"},
			exceptions:  []string{"X-Amzn-Bedrock-AgentCore-Runtime-Custom-"},
			expectError: false,
		},
		"valid-exception-case-insensitive": {
			in:          types.StringValue("x-amzn-bedrock-agentcore-runtime-custom-myheader"),
			prefixes:    []string{"X-Amzn-"},
			exceptions:  []string{"X-Amzn-Bedrock-AgentCore-Runtime-Custom-"},
			expectError: false,
		},
		"null": {
			in:          types.StringNull(),
			prefixes:    []string{"X-Amzn-"},
			expectError: false,
		},
		"unknown": {
			in:          types.StringUnknown(),
			prefixes:    []string{"X-Amzn-"},
			expectError: false,
		},
	}

	for name, test := range testCases {
		t.Run(fmt.Sprintf("ValidateString - %s", name), func(t *testing.T) {
			t.Parallel()
			req := validator.StringRequest{
				ConfigValue: test.in,
			}
			res := validator.StringResponse{}
			tfstringvalidator.PrefixNoneOfCaseInsensitive(test.prefixes, test.exceptions).ValidateString(t.Context(), req, &res)

			if !res.Diagnostics.HasError() && test.expectError {
				t.Fatal("expected error, got no error")
			}

			if res.Diagnostics.HasError() && !test.expectError {
				t.Fatalf("got unexpected error: %s", res.Diagnostics)
			}
		})
	}
}
