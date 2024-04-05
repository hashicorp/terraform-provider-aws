// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	itypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

var (
	_ xattr.TypeWithValidate   = (*cidrBlockType)(nil)
	_ basetypes.StringTypable  = (*cidrBlockType)(nil)
	_ basetypes.StringValuable = (*CIDRBlock)(nil)
)

type cidrBlockType struct {
	basetypes.StringType
}

var (
	CIDRBlockType = cidrBlockType{}
)

func (t cidrBlockType) Equal(o attr.Type) bool {
	other, ok := o.(cidrBlockType)

	if !ok {
		return false
	}

	return t.StringType.Equal(other.StringType)
}

func (cidrBlockType) String() string {
	return "CIDRBlockType"
}

func (t cidrBlockType) ValueFromString(_ context.Context, in types.String) (basetypes.StringValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	if in.IsNull() {
		return CIDRBlockNull(), diags
	}
	if in.IsUnknown() {
		return CIDRBlockUnknown(), diags
	}

	valueString := in.ValueString()
	if err := itypes.ValidateCIDRBlock(valueString); err != nil {
		return CIDRBlockUnknown(), diags // Must not return validation errors
	}

	return CIDRBlockValue(valueString), diags
}

func (t cidrBlockType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.StringType.ValueFromTerraform(ctx, in)

	if err != nil {
		return nil, err
	}

	stringValue, ok := attrValue.(basetypes.StringValue)

	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	stringValuable, diags := t.ValueFromString(ctx, stringValue)

	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting StringValue to StringValuable: %v", diags)
	}

	return stringValuable, nil
}

func (cidrBlockType) ValueType(context.Context) attr.Value {
	return CIDRBlock{}
}

func (t cidrBlockType) Validate(ctx context.Context, in tftypes.Value, path path.Path) diag.Diagnostics {
	var diags diag.Diagnostics

	if !in.IsKnown() || in.IsNull() {
		return diags
	}

	var value string
	err := in.As(&value)
	if err != nil {
		diags.AddAttributeError(
			path,
			"CIDRBlock Type Validation Error",
			ProviderErrorDetailPrefix+fmt.Sprintf("Cannot convert value to string: %s", err),
		)
		return diags
	}

	if err := itypes.ValidateCIDRBlock(value); err != nil {
		diags.AddAttributeError(
			path,
			"CIDRBlock Type Validation Error",
			err.Error(),
		)
		return diags
	}

	return diags
}

func CIDRBlockNull() CIDRBlock {
	return CIDRBlock{StringValue: basetypes.NewStringNull()}
}

func CIDRBlockUnknown() CIDRBlock {
	return CIDRBlock{StringValue: basetypes.NewStringUnknown()}
}

func CIDRBlockValue(value string) CIDRBlock {
	return CIDRBlock{StringValue: basetypes.NewStringValue(value)}
}

type CIDRBlock struct {
	basetypes.StringValue
}

func (v CIDRBlock) Equal(o attr.Value) bool {
	other, ok := o.(CIDRBlock)

	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

func (CIDRBlock) Type(context.Context) attr.Type {
	return CIDRBlockType
}
