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

func TestNoneOfCaseInsensitive(t *testing.T) {
	t.Parallel()

	type testCase struct {
		in          types.String
		noneOf      []string
		expectError bool
	}

	testCases := map[string]testCase{
		"valid-not-in-list": {
			in:          types.StringValue("x-custom-header"),
			noneOf:      []string{"Authorization", "Host"},
			expectError: false,
		},
		"invalid-exact-match": {
			in:          types.StringValue("Authorization"),
			noneOf:      []string{"Authorization", "Host"},
			expectError: true,
		},
		"invalid-case-insensitive": {
			in:          types.StringValue("authorization"),
			noneOf:      []string{"Authorization", "Host"},
			expectError: true,
		},
		"invalid-mixed-case": {
			in:          types.StringValue("AUTHORIZATION"),
			noneOf:      []string{"Authorization", "Host"},
			expectError: true,
		},
		"null": {
			in:          types.StringNull(),
			noneOf:      []string{"Authorization"},
			expectError: false,
		},
		"unknown": {
			in:          types.StringUnknown(),
			noneOf:      []string{"Authorization"},
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
			tfstringvalidator.NoneOfCaseInsensitive(test.noneOf...).ValidateString(t.Context(), req, &res)

			if !res.Diagnostics.HasError() && test.expectError {
				t.Fatal("expected error, got no error")
			}

			if res.Diagnostics.HasError() && !test.expectError {
				t.Fatalf("got unexpected error: %s", res.Diagnostics)
			}
		})
	}
}
