package validators_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
)

func TestBetweenValidator(t *testing.T) {
	t.Parallel()

	type testCase struct {
		val         attr.Value
		min         int64
		max         int64
		expectError bool
	}
	tests := map[string]testCase{
		"not a String": {
			val:         types.BoolValue(true),
			expectError: true,
		},
		"unknown String": {
			val: types.StringUnknown(),
			min: 1,
			max: 3,
		},
		"null String": {
			val: types.StringNull(),
			min: 1,
			max: 3,
		},
		"invalid String": {
			val:         types.StringValue("test-value"),
			min:         1,
			max:         3,
			expectError: true,
		},
		"valid string": {
			val: types.StringValue("2"),
			min: 1,
			max: 3,
		},
		"valid string min": {
			val: types.StringValue("1"),
			min: 1,
			max: 3,
		},
		"valid string max": {
			val: types.StringValue("3"),
			min: 1,
			max: 3,
		},
		"too small string": {
			val:         types.StringValue("-1"),
			min:         1,
			max:         3,
			expectError: true,
		},
		"too large string": {
			val:         types.StringValue("42"),
			min:         1,
			max:         3,
			expectError: true,
		},
	}

	for name, test := range tests {
		name, test := name, test
		t.Run(name, func(t *testing.T) {
			request := tfsdk.ValidateAttributeRequest{
				AttributePath:           path.Root("test"),
				AttributePathExpression: path.MatchRoot("test"),
				AttributeConfig:         test.val,
			}
			response := tfsdk.ValidateAttributeResponse{}
			fwvalidators.Int64StringBetween(test.min, test.max).Validate(context.Background(), request, &response)

			if !response.Diagnostics.HasError() && test.expectError {
				t.Fatal("expected error, got no error")
			}

			if response.Diagnostics.HasError() && !test.expectError {
				t.Fatalf("got unexpected error: %s", response.Diagnostics)
			}
		})
	}
}
