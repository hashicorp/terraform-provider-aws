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

func TestIPv4CIDRNetworkAddressValidator(t *testing.T) {
	t.Parallel()

	type testCase struct {
		val         attr.Value
		expectError bool
	}
	tests := map[string]testCase{
		"not a String": {
			val:         types.Bool{Value: true},
			expectError: true,
		},
		"unknown String": {
			val: types.String{Unknown: true},
		},
		"null String": {
			val: types.String{Null: true},
		},
		"invalid String": {
			val:         types.String{Value: "test-value"},
			expectError: true,
		},
		"valid string": {
			val: types.String{Value: "10.2.2.0/24"},
		},
		"invalid IPv4 CIDR": {
			val:         types.String{Value: "10.2.2.2/24"},
			expectError: true,
		},
		"valid IPv6 CIDR": {
			val:         types.String{Value: "2000::/15"},
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
			fwvalidators.IPv4CIDRNetworkAddress().Validate(context.Background(), request, &response)

			if !response.Diagnostics.HasError() && test.expectError {
				t.Fatal("expected error, got no error")
			}

			if response.Diagnostics.HasError() && !test.expectError {
				t.Fatalf("got unexpected error: %s", response.Diagnostics)
			}
		})
	}
}
