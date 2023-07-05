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

type cidrBlockType uint8

const (
	CIDRBlockType cidrBlockType = iota
)

var (
	_ xattr.TypeWithValidate = CIDRBlockType
)

func (t cidrBlockType) TerraformType(_ context.Context) tftypes.Type {
	return tftypes.String
}

func (t cidrBlockType) ValueFromString(_ context.Context, st types.String) (basetypes.StringValuable, diag.Diagnostics) {
	if st.IsNull() {
		return CIDRBlockNull(), nil
	}
	if st.IsUnknown() {
		return CIDRBlockUnknown(), nil
	}

	return CIDRBlockValue(st.ValueString()), nil
}

func (t cidrBlockType) ValueFromTerraform(_ context.Context, in tftypes.Value) (attr.Value, error) {
	if in.IsNull() {
		return CIDRBlockNull(), nil
	}
	if !in.IsKnown() {
		return CIDRBlockUnknown(), nil
	}

	var s string
	err := in.As(&s)

	if err != nil {
		return nil, err
	}

	if err := itypes.ValidateCIDRBlock(s); err != nil {
		return CIDRBlockUnknown(), nil //nolint: nilerr // Must not return validation errors
	}

	return CIDRBlockValue(s), nil
}

func (t cidrBlockType) ValueType(context.Context) attr.Value {
	return CIDRBlock{}
}

func (t cidrBlockType) Equal(o attr.Type) bool {
	_, ok := o.(cidrBlockType)
	return ok
}

func (t cidrBlockType) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (interface{}, error) {
	return nil, fmt.Errorf("cannot apply AttributePathStep %T to %s", step, t.String())
}

func (t cidrBlockType) String() string {
	return "types.CIDRBlockType"
}

func (t cidrBlockType) Validate(ctx context.Context, in tftypes.Value, path path.Path) diag.Diagnostics {
	var diags diag.Diagnostics

	if !in.Type().Is(tftypes.String) {
		diags.AddAttributeError(
			path,
			"CIDRBlock Type Validation Error",
			"An unexpected error was encountered trying to validate an attribute value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
				fmt.Sprintf("Expected String value, received %T with value: %v", in, in),
		)
		return diags
	}

	if !in.IsKnown() || in.IsNull() {
		return diags
	}

	var value string
	err := in.As(&value)
	if err != nil {
		diags.AddAttributeError(
			path,
			"CIDRBlock Type Validation Error",
			"An unexpected error was encountered trying to validate an attribute value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
				fmt.Sprintf("Cannot convert value to string: %s", err),
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

func (t cidrBlockType) Description() string {
	return `A CIDR block.`
}

func CIDRBlockNull() CIDRBlock {
	return CIDRBlock{
		state: attr.ValueStateNull,
	}
}

func CIDRBlockUnknown() CIDRBlock {
	return CIDRBlock{
		state: attr.ValueStateUnknown,
	}
}

func CIDRBlockValue(value string) CIDRBlock {
	return CIDRBlock{
		state: attr.ValueStateKnown,
		value: value,
	}
}

type CIDRBlock struct {
	state attr.ValueState
	value string
}

func (c CIDRBlock) Type(_ context.Context) attr.Type {
	return CIDRBlockType
}

func (c CIDRBlock) ToStringValue(ctx context.Context) (types.String, diag.Diagnostics) {
	switch c.state {
	case attr.ValueStateKnown:
		return types.StringValue(c.value), nil
	case attr.ValueStateNull:
		return types.StringNull(), nil
	case attr.ValueStateUnknown:
		return types.StringUnknown(), nil
	default:
		return types.StringUnknown(), diag.Diagnostics{
			diag.NewErrorDiagnostic(fmt.Sprintf("unhandled CIDRBlock state in ToStringValue: %s", c.state), ""),
		}
	}
}

func (c CIDRBlock) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	t := CIDRBlockType.TerraformType(ctx)

	switch c.state {
	case attr.ValueStateKnown:
		if err := tftypes.ValidateValue(t, c.value); err != nil {
			return tftypes.NewValue(t, tftypes.UnknownValue), err
		}

		return tftypes.NewValue(t, c.value), nil
	case attr.ValueStateNull:
		return tftypes.NewValue(t, nil), nil
	case attr.ValueStateUnknown:
		return tftypes.NewValue(t, tftypes.UnknownValue), nil
	default:
		return tftypes.NewValue(t, tftypes.UnknownValue), fmt.Errorf("unhandled CIDRBlock state in ToTerraformValue: %s", c.state)
	}
}

func (c CIDRBlock) Equal(other attr.Value) bool {
	o, ok := other.(CIDRBlock)

	if !ok {
		return false
	}

	if c.state != o.state {
		return false
	}

	if c.state != attr.ValueStateKnown {
		return true
	}

	return c.value == o.value
}

func (c CIDRBlock) IsNull() bool {
	return c.state == attr.ValueStateNull
}

func (c CIDRBlock) IsUnknown() bool {
	return c.state == attr.ValueStateUnknown
}

func (c CIDRBlock) String() string {
	if c.IsNull() {
		return attr.NullValueString
	}
	if c.IsUnknown() {
		return attr.UnknownValueString
	}

	return c.value
}

func (c CIDRBlock) ValueCIDRBlock() string {
	return c.value
}
