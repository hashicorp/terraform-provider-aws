// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package validators

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestNonNullValuesValidator(t *testing.T) {
	t.Parallel()

	type testCase struct {
		val         types.Set
		expectError bool
	}
	tests := map[string]testCase{
		"Set unknown": {
			val: types.SetUnknown(
				types.StringType,
			),
			expectError: false,
		},
		"Set null": {
			val: types.SetNull(
				types.StringType,
			),
			expectError: false,
		},
		"Null values": {
			val: types.SetValueMust(
				types.StringType,
				[]attr.Value{
					types.StringValue("first"),
					types.StringNull(),
				},
			),
			expectError: true,
		},
		"Non null values": {
			val: types.SetValueMust(
				types.StringType,
				[]attr.Value{
					types.StringValue("first"),
					types.StringValue("second"),
				},
			),
			expectError: false,
		},
	}

	for name, test := range tests {
		t.Run(fmt.Sprintf("ValidateSet - %s", name), func(t *testing.T) {
			t.Parallel()
			request := validator.SetRequest{
				Path:           path.Root("test"),
				PathExpression: path.MatchRoot("test"),
				ConfigValue:    test.val,
			}
			response := validator.SetResponse{}
			NonNullValues().ValidateSet(context.TODO(), request, &response)

			if !response.Diagnostics.HasError() && test.expectError {
				t.Fatal("expected error, got no error")
			}

			if response.Diagnostics.HasError() && !test.expectError {
				t.Fatalf("got unexpected error: %s", response.Diagnostics)
			}
		})

		t.Run(fmt.Sprintf("ValidateParameterSet - %s", name), func(t *testing.T) {
			t.Parallel()
			request := function.SetParameterValidatorRequest{
				ArgumentPosition: 0,
				Value:            test.val,
			}
			response := function.SetParameterValidatorResponse{}
			NonNullValues().ValidateParameterSet(context.TODO(), request, &response)

			if response.Error == nil && test.expectError {
				t.Fatal("expected error, got no error")
			}

			if response.Error != nil && !test.expectError {
				t.Fatalf("got unexpected error: %s", response.Error)
			}
		})
	}
}
