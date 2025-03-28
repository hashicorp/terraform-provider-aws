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

func TestAwsOrganizationRootIDValidator(t *testing.T) { // nosemgrep:ci.aws-in-func-name
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
					`Attribute test value must be a valid AWS organization root ID, got: test-value`,
				),
			},
		},
		"valid AWS organization Root ID": {
			val: types.StringValue("r-xy6f"),
		},
		"too long AWS organization Root ID": {
			val: types.StringValue(`r-xy6ffog945kghe8jjg948984jf8e2ifo3e`),
			expectedDiagnostics: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid Attribute Value",
					`Attribute test value must be a valid AWS organization root ID, got: r-xy6ffog945kghe8jjg948984jf8e2ifo3e`,
				),
			},
		},
		"too short AWS organization root ID": {
			val: types.StringValue("r-xf"),
			expectedDiagnostics: diag.Diagnostics{
				diag.NewAttributeErrorDiagnostic(
					path.Root("test"),
					"Invalid Attribute Value",
					`Attribute test value must be a valid AWS organization root ID, got: r-xf`,
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
			fwvalidators.AWSOrganizationRootID().ValidateString(ctx, request, &response)

			if diff := cmp.Diff(response.Diagnostics, tc.expectedDiagnostics); diff != "" {
				t.Errorf("unexpected diagnostics difference: %s", diff)
			}
		})
	}
}
