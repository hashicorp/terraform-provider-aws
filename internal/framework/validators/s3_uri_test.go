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

func TestS3URIValidator(t *testing.T) {
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
					`Attribute test value must be a valid S3 URI, got: test-value`,
				),
			},
		},
		"valid S3 URI": {
			val: types.StringValue("s3://bucket/path/to/key"),
		},
		"invalid characters": {
			val: types.StringValue("s3://asbcdefg--#/key"),
			expectedDiagnostics: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid Attribute Value",
					`Attribute test value must be a valid S3 URI, got: s3://asbcdefg--#/key`,
				),
			},
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()

			request := validator.StringRequest{
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				ConfigValue:    test.val,
			}
			response := validator.StringResponse{}
			fwvalidators.S3URI().ValidateString(ctx, request, &response)

			if diff := cmp.Diff(response.Diagnostics, test.expectedDiagnostics); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}
