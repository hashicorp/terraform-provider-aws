// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework-validators/helpers/validatordiag"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestValidPolicyPathFramework(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		value         string
		expectedDiags diag.Diagnostics
	}{
		"root path": {
			value: "/",
		},

		"single element": {
			value: "/test/",
		},

		"multiple elements": {
			value: "/test1/test2/test3/",
		},

		"empty path": {
			value: "",
			expectedDiags: diag.Diagnostics{
				validatordiag.InvalidAttributeValueLengthDiagnostic(
					path.Root(names.AttrPath),
					"string length must be between 1 and 512",
					"0",
				)},
		},

		"missing leading slash": {
			value: "test/",
			expectedDiags: diag.Diagnostics{
				validatordiag.InvalidAttributeValueDiagnostic(
					path.Root(names.AttrPath),
					"value must begin and end with a slash (/)",
					"test/",
				),
			},
		},

		"missing trailing slash": {
			value: "/test",
			expectedDiags: diag.Diagnostics{
				validatordiag.InvalidAttributeValueDiagnostic(
					path.Root(names.AttrPath),
					"value must begin and end with a slash (/)",
					"/test",
				),
			},
		},

		"consecutive slashes": {
			value: "/test//",
			expectedDiags: diag.Diagnostics{
				validatordiag.InvalidAttributeValueDiagnostic(
					path.Root(names.AttrPath),
					"value must not contain consecutive slashes (//)",
					"/test//",
				),
			},
		},

		"invalid character": {
			value: "/test!/",
			expectedDiags: diag.Diagnostics{
				validatordiag.InvalidAttributeValueMatchDiagnostic(
					path.Root(names.AttrPath),
					"value must contain uppercase or lowercase alphanumeric characters or any of the following: / , . + @ = _ -",
					"/test!/",
				),
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			request := validator.StringRequest{
				Path:           path.Root(names.AttrPath),
				PathExpression: path.MatchRoot(names.AttrPath),
				ConfigValue:    types.StringValue(tc.value),
			}
			var response validator.StringResponse
			validPolicyPathFramework.ValidateString(ctx, request, &response)

			if diff := cmp.Diff(response.Diagnostics, tc.expectedDiags); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}
