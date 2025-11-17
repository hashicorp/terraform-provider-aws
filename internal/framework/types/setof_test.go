// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

func TestSetOfStringFromTerraform(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tests := map[string]struct {
		val      tftypes.Value
		expected attr.Value
	}{
		"values": {
			val: tftypes.NewValue(tftypes.Set{
				ElementType: tftypes.String,
			}, []tftypes.Value{
				tftypes.NewValue(tftypes.String, "red"),
				tftypes.NewValue(tftypes.String, "blue"),
				tftypes.NewValue(tftypes.String, "green"),
			}),
			expected: fwtypes.NewSetValueOfMust[types.String](ctx, []attr.Value{
				types.StringValue("red"),
				types.StringValue("blue"),
				types.StringValue("green"),
			}),
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			val, err := fwtypes.SetOfStringType.ValueFromTerraform(ctx, test.val)

			if err != nil {
				t.Fatalf("got unexpected error: %s", err)
			}

			if diff := cmp.Diff(val, test.expected); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

func TestSetOfValidateAttribute(t *testing.T) {
	t.Parallel()

	tests := map[string]struct {
		val         []attr.Value
		expectError bool
	}{
		"null value": {
			val: []attr.Value{
				fwtypes.StringEnumNull[mockEnum](),
			},
		},
		"unknown value": {
			val: []attr.Value{
				fwtypes.StringEnumUnknown[mockEnum](),
			},
		},
		"valid values": { // lintignore:AWSAT003,AWSAT005
			val: []attr.Value{
				fwtypes.StringEnumValue[mockEnum]("red"),
				fwtypes.StringEnumValue[mockEnum]("blue"),
			},
		},
		"invalid values": {
			val: []attr.Value{
				fwtypes.StringEnumValue[mockEnum]("blue"),
				fwtypes.StringEnumValue[mockEnum]("green"),
			},
			expectError: true,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			req := xattr.ValidateAttributeRequest{}
			resp := xattr.ValidateAttributeResponse{}

			listOfEnums := fwtypes.SetOfStringEnumType[mockEnum]()
			values, _ := listOfEnums.ValueFromSet(ctx, types.SetValueMust(fwtypes.StringEnumType[mockEnum](), test.val))

			// asserting here because we know the interface is implemented
			eval := values.(fwtypes.SetValueOf[fwtypes.StringEnum[mockEnum]])
			eval.ValidateAttribute(ctx, req, &resp)
			if resp.Diagnostics.HasError() != test.expectError {
				t.Errorf("resp.Diagnostics.HasError() = %t, want = %t", resp.Diagnostics.HasError(), test.expectError)
			}
		})
	}
}
