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

type arnType uint8

const (
	ARNType arnType = iota
)

var (
	_ xattr.TypeWithValidate = ARNType
)

func (t arnType) TerraformType(_ context.Context) tftypes.Type {
	return tftypes.String
}

func (t arnType) ValueFromString(_ context.Context, st types.String) (basetypes.StringValuable, diag.Diagnostics) {
	if st.IsNull() {
		return ARNNull(), nil
	}
	if st.IsUnknown() {
		return ARNUnknown(), nil
	}

	var diags diag.Diagnostics
	v, err := arn.Parse(st.ValueString())
	if err != nil {
		diags.AddError(
			"ARN ValueFromString Error",
			fmt.Sprintf("String %s cannot be parsed as an ARN.", st),
		)
		return nil, diags
	}

	return ARNValue(v), diags
}

func (t arnType) ValueFromTerraform(_ context.Context, in tftypes.Value) (attr.Value, error) {
	if !in.IsKnown() {
		return ARNUnknown(), nil
	}

	if in.IsNull() {
		return ARNNull(), nil
	}

	var s string
	err := in.As(&s)

	if err != nil {
		return nil, err
	}

	v, err := arn.Parse(s)

	if err != nil {
		return ARNUnknown(), nil //nolint: nilerr // Must not return validation errors
	}

	return ARNValue(v), nil
}

func (t arnType) ValueType(context.Context) attr.Value {
	return ARN{}
}

// Equal returns true if `o` is also an ARNType.
func (t arnType) Equal(o attr.Type) bool {
	_, ok := o.(arnType)
	return ok
}

// ApplyTerraform5AttributePathStep applies the given AttributePathStep to the
// type.
func (t arnType) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (interface{}, error) {
	return nil, fmt.Errorf("cannot apply AttributePathStep %T to %s", step, t.String())
}

// String returns a human-friendly description of the ARNType.
func (t arnType) String() string {
	return "types.ARNType"
}

// Validate implements type validation.
func (t arnType) Validate(ctx context.Context, in tftypes.Value, path path.Path) diag.Diagnostics {
	var diags diag.Diagnostics

	if !in.Type().Is(tftypes.String) {
		diags.AddAttributeError(
			path,
			"ARN Type Validation Error",
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
			"ARN Type Validation Error",
			"An unexpected error was encountered trying to validate an attribute value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
				fmt.Sprintf("Cannot convert value to arn.ARN: %s", err),
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

func (t arnType) Description() string {
	return `An Amazon Resource Name.`
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
