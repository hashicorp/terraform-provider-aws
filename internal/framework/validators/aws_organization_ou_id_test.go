// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validators_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
)

func TestAWSOrganizationOUIDValidator(t *testing.T) { // nosemgrep:ci.aws-in-func-name
	t.Parallel()

	type testCase struct {
		val                 types.String
		expectedDiagnostics diag.Diagnostics
	}
	tests := map[string]testCase{
		"unknown string": {
			val: types.StringUnknown(),
		},
		"null string": {
			val: types.StringNull(),
		},
		"invalid string": {
			val: types.StringValue("test-value"),
			expectedDiagnostics: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid Attribute Value",
					`Attribute test value must be a valid AWS organizational unit ID, got: test-value`,
				),
			},
		},
		"valid AWS organizational unit ID": {
			val: types.StringValue("ou-z7jt-19mqs9sp"),
		},
		"too long AWS organizational unit ID": {
			val: types.StringValue(`ou-z7jt-19mqs9sp42iu99e322rt46hf9er237hf9xc`),
			expectedDiagnostics: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid Attribute Value",
					`Attribute test value must be a valid AWS organizational unit ID, got: ou-z7jt-19mqs9sp42iu99e322rt46hf9er237hf9xc`,
				),
			},
		},
		"too short AWS organizational unit ID": {
			val: types.StringValue("r-xf"),
			expectedDiagnostics: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid Attribute Value",
					`Attribute test value must be a valid AWS organizational unit ID, got: r-xf`,
				),
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			request := validator.StringRequest{
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				ConfigValue:    tc.val,
			}
			response := validator.StringResponse{}
			fwvalidators.AWSOrganizationOUID().ValidateString(ctx, request, &response)

			if diff := cmp.Diff(response.Diagnostics, tc.expectedDiagnostics); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}
