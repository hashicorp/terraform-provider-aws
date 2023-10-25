// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// ProviderErrorDetailPrefix contains instructions for reporting provider errors to provider developers
const ProviderErrorDetailPrefix = "An unexpected error was encountered trying to validate an attribute value. " +
	"This is always an error in the provider. Please report the following to the provider developer:\n\n"

type arnType struct {
	basetypes.StringType
}

var (
	ARNType = arnType{}
)

var (
	_ xattr.TypeWithValidate   = ARNType
	_ basetypes.StringTypable  = ARNType
	_ basetypes.StringValuable = ARN{}
)

func (t arnType) Equal(o attr.Type) bool {
	other, ok := o.(arnType)

	if !ok {
		return false
	}

	return t.StringType.Equal(other.StringType)
}

func (t arnType) String() string {
	return "ARNType"
}

func (t arnType) ValueFromString(_ context.Context, in types.String) (basetypes.StringValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	if in.IsNull() {
		return ARNNull(), diags
	}
	if in.IsUnknown() {
		return ARNUnknown(), diags
	}

	v, err := arn.Parse(in.ValueString())
	if err != nil {
		return ARNUnknown(), diags // Must not return validation errors.
	}

	return ARNValue(v), diags
}

func (t arnType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
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

func (t arnType) ValueType(context.Context) attr.Value {
	return ARN{}
}

func (t arnType) Validate(ctx context.Context, in tftypes.Value, path path.Path) diag.Diagnostics {
	var diags diag.Diagnostics

	if !in.IsKnown() || in.IsNull() {
		return diags
	}

	var value string
	err := in.As(&value)
	if err != nil {
		diags.AddAttributeError(
			path,
			"ARN Type Validation Error",
			ProviderErrorDetailPrefix+fmt.Sprintf("Cannot convert value to string: %s", err),
		)
		return diags
	}

	if !arn.IsARN(value) {
		diags.AddAttributeError(
			path,
			"ARN Type Validation Error",
			fmt.Sprintf("Value %q cannot be parsed as an ARN.", value),
		)
		return diags
	}

	return diags
}

func ARNNull() ARN {
	return ARN{
		state: attr.ValueStateNull,
	}
}

func ARNUnknown() ARN {
	return ARN{
		state: attr.ValueStateUnknown,
	}
}

func ARNValue(value arn.ARN) ARN {
	return ARN{
		state: attr.ValueStateKnown,
		value: value,
	}
}

type ARN struct {
	// state represents whether the value is null, unknown, or known. The
	// zero-value is null.
	state attr.ValueState

	// value contains the known value, if not null or unknown.
	value arn.ARN
}

func (a ARN) Type(_ context.Context) attr.Type {
	return ARNType
}

func (a ARN) ToStringValue(ctx context.Context) (types.String, diag.Diagnostics) {
	switch a.state {
	case attr.ValueStateKnown:
		return types.StringValue(a.value.String()), nil
	case attr.ValueStateNull:
		return types.StringNull(), nil
	case attr.ValueStateUnknown:
		return types.StringUnknown(), nil
	default:
		return types.StringUnknown(), diag.Diagnostics{
			diag.NewErrorDiagnostic(fmt.Sprintf("unhandled ARN state in ToStringValue: %s", a.state), ""),
		}
	}
}

func (a ARN) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	t := ARNType.TerraformType(ctx)

	switch a.state {
	case attr.ValueStateKnown:
		if err := tftypes.ValidateValue(t, a.value.String()); err != nil {
			return tftypes.NewValue(t, tftypes.UnknownValue), err
		}

		return tftypes.NewValue(t, a.value.String()), nil
	case attr.ValueStateNull:
		return tftypes.NewValue(t, nil), nil
	case attr.ValueStateUnknown:
		return tftypes.NewValue(t, tftypes.UnknownValue), nil
	default:
		return tftypes.NewValue(t, tftypes.UnknownValue), fmt.Errorf("unhandled ARN state in ToTerraformValue: %s", a.state)
	}
}

// Equal returns true if `other` is a *ARN and has the same value as `a`.
func (a ARN) Equal(other attr.Value) bool {
	o, ok := other.(ARN)

	if !ok {
		return false
	}

	if a.state != o.state {
		return false
	}

	if a.state != attr.ValueStateKnown {
		return true
	}

	return a.value == o.value
}

// IsNull returns true if the Value is not set, or is explicitly set to null.
func (a ARN) IsNull() bool {
	return a.state == attr.ValueStateNull
}

// IsUnknown returns true if the Value is not yet known.
func (a ARN) IsUnknown() bool {
	return a.state == attr.ValueStateUnknown
}

// String returns a summary representation of either the underlying Value,
// or UnknownValueString (`<unknown>`) when IsUnknown() returns true,
// or NullValueString (`<null>`) when IsNull() return true.
//
// This is an intentionally lossy representation, that are best suited for
// logging and error reporting, as they are not protected by
// compatibility guarantees within the framework.
func (a ARN) String() string {
	if a.IsUnknown() {
		return attr.UnknownValueString
	}

	if a.IsNull() {
		return attr.NullValueString
	}

	return a.value.String()
}

// ValueARN returns the known arn.ARN value. If ARN is null or unknown, returns {}.
func (a ARN) ValueARN() arn.ARN {
	return a.value
}
