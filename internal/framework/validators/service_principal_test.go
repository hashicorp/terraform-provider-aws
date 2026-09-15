// Copyright IBM Corp. 2014, 2026
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

func TestServicePrincipalValidator(t *testing.T) {
	t.Parallel()

	type testCase struct {
		val                 types.String
		expectedDiagnostics diag.Diagnostics
	}
	tests := map[string]testCase{
		"unknown String": {
			val: types.StringUnknown(),
		},
		"null String": {
			val: types.StringNull(),
		},
		"invalid String": {
			val: types.StringValue("test-value"),
			expectedDiagnostics: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid Attribute Value",
					`Attribute test value must be a valid AWS service principal (e.g., ec2.amazonaws.com), got: test-value`,
				),
			},
		},
		"valid service principal ec2": {
			val: types.StringValue("ec2.amazonaws.com"),
		},
		"valid service principal s3": {
			val: types.StringValue("s3.amazonaws.com"),
		},
		"valid service principal lambda": {
			val: types.StringValue("lambda.amazonaws.com"),
		},
		"valid service principal with hyphen": {
			val: types.StringValue("ecs-tasks.amazonaws.com"),
		},
		"valid service principal with multiple parts": {
			val: types.StringValue("delivery.logs.amazonaws.com"),
		},
		"valid service principal amazon.com": {
			val: types.StringValue("ec2.amazon.com"),
		},
		"invalid ARN instead of principal": {
			val: types.StringValue("arn:aws:iam::123456789012:role/test"),
			expectedDiagnostics: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid Attribute Value",
					`Attribute test value must be a valid AWS service principal (e.g., ec2.amazonaws.com), got: arn:aws:iam::123456789012:role/test`,
				),
			},
		},
		"invalid account ID instead of principal": {
			val: types.StringValue("123456789012"),
			expectedDiagnostics: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid Attribute Value",
					`Attribute test value must be a valid AWS service principal (e.g., ec2.amazonaws.com), got: 123456789012`,
				),
			},
		},
		"invalid domain": {
			val: types.StringValue("ec2.example.com"),
			expectedDiagnostics: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid Attribute Value",
					`Attribute test value must be a valid AWS service principal (e.g., ec2.amazonaws.com), got: ec2.example.com`,
				),
			},
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			request := validator.StringRequest{
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				ConfigValue:    test.val,
			}
			response := validator.StringResponse{}
			fwvalidators.ServicePrincipal().ValidateString(ctx, request, &response)

			if diff := cmp.Diff(response.Diagnostics, test.expectedDiagnostics); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}
