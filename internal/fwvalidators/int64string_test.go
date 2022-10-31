package fwvalidators_test

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/fwvalidators"
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
			val:         types.Bool{Value: true},
			expectError: true,
		},
		"unknown String": {
			val: types.String{Unknown: true},
			min: 1,
			max: 3,
		},
		"null String": {
			val: types.String{Null: true},
			min: 1,
			max: 3,
		},
		"invalid String": {
			val:         types.String{Value: "test-value"},
			min:         1,
			max:         3,
			expectError: true,
		},
		"valid string": {
			val: types.String{Value: "2"},
			min: 1,
			max: 3,
		},
		"valid string min": {
			val: types.String{Value: "1"},
			min: 1,
			max: 3,
		},
		"valid string max": {
			val: types.String{Value: "3"},
			min: 1,
			max: 3,
		},
		"too small string": {
			val:         types.String{Value: "-1"},
			min:         1,
			max:         3,
			expectError: true,
		},
		"too large string": {
			val:         types.String{Value: "42"},
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
