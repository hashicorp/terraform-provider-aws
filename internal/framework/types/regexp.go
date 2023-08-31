// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type regexpType struct{}

var (
	RegexpType = regexpType{}
)

var (
	_ xattr.TypeWithValidate = RegexpType
)

func (t regexpType) TerraformType(_ context.Context) tftypes.Type {
	return tftypes.String
}

func (t regexpType) ValueFromString(_ context.Context, st types.String) (basetypes.StringValuable, diag.Diagnostics) {
	if st.IsNull() {
		return RegexpNull(), nil
	}
	if st.IsUnknown() {
		return RegexpUnknown(), nil
	}

	var diags diag.Diagnostics
	v, err := regexp.Compile(st.ValueString())
	if err != nil {
		diags.AddError(
			"Regexp ValueFromString Error",
			fmt.Sprintf("String %s cannot be parsed as a regular expression.", st),
		)
		return nil, diags
	}

	return RegexpValue(v), diags
}

func (t regexpType) ValueFromTerraform(_ context.Context, in tftypes.Value) (attr.Value, error) {
	if !in.IsKnown() {
		return RegexpUnknown(), nil
	}

	if in.IsNull() {
		return RegexpNull(), nil
	}

	var s string
	err := in.As(&s)

	if err != nil {
		return nil, err
	}

	v, err := regexp.Compile(s)

	if err != nil {
		return RegexpUnknown(), nil //nolint: nilerr // Must not return validation errors
	}

	return RegexpValue(v), nil
}

func (t regexpType) ValueType(context.Context) attr.Value {
	return Regexp{}
}

// Equal returns true if `o` is also a RegexpType.
func (t regexpType) Equal(o attr.Type) bool {
	_, ok := o.(regexpType)
	return ok
}

// ApplyTerraform5AttributePathStep applies the given AttributePathStep to the
// type.
func (t regexpType) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (interface{}, error) {
	return nil, fmt.Errorf("cannot apply AttributePathStep %T to %s", step, t.String())
}

// String returns a human-friendly description of the RegexpType.
func (t regexpType) String() string {
	return "types.RegexpType"
}

// Validate implements type validation.
func (t regexpType) Validate(ctx context.Context, in tftypes.Value, path path.Path) diag.Diagnostics {
	var diags diag.Diagnostics

	if !in.Type().Is(tftypes.String) {
		diags.AddAttributeError(
			path,
			"Regexp Type Validation Error",
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
			"Regexp Type Validation Error",
			"An unexpected error was encountered trying to validate an attribute value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
				fmt.Sprintf("Cannot convert value to string: %s", err),
		)
		return diags
	}

	if _, err := regexp.Compile(value); err != nil {
		diags.AddAttributeError(
			path,
			"Regexp Type Validation Error",
			fmt.Sprintf("Value %q cannot be parsed as a regular expression: %s", value, err),
		)
		return diags
	}

	return diags
}

func (t regexpType) Description() string {
	return `A regular expression.`
}

func RegexpNull() Regexp {
	return Regexp{
		state: attr.ValueStateNull,
	}
}

func RegexpUnknown() Regexp {
	return Regexp{
		state: attr.ValueStateUnknown,
	}
}

func RegexpValue(value *regexp.Regexp) Regexp {
	return Regexp{
		state: attr.ValueStateKnown,
		value: value,
	}
}

type Regexp struct {
	// state represents whether the value is null, unknown, or known. The
	// zero-value is null.
	state attr.ValueState

	// value contains the known value, if not null or unknown.
	value *regexp.Regexp
}

func (a Regexp) Type(_ context.Context) attr.Type {
	return RegexpType
}

func (a Regexp) ToStringValue(ctx context.Context) (types.String, diag.Diagnostics) {
	switch a.state {
	case attr.ValueStateKnown:
		return types.StringValue(a.value.String()), nil
	case attr.ValueStateNull:
		return types.StringNull(), nil
	case attr.ValueStateUnknown:
		return types.StringUnknown(), nil
	default:
		return types.StringUnknown(), diag.Diagnostics{
			diag.NewErrorDiagnostic(fmt.Sprintf("unhandled Regexp state in ToStringValue: %s", a.state), ""),
		}
	}
}

func (a Regexp) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	t := RegexpType.TerraformType(ctx)

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
		return tftypes.NewValue(t, tftypes.UnknownValue), fmt.Errorf("unhandled Regexp state in ToTerraformValue: %s", a.state)
	}
}

// Equal returns true if `other` is a Regexp and has the same value as `a`.
func (a Regexp) Equal(other attr.Value) bool {
	o, ok := other.(Regexp)

	if !ok {
		return false
	}

	if a.state != o.state {
		return false
	}

	if a.state != attr.ValueStateKnown {
		return true
	}

	return a.value.String() == o.value.String()
}

// IsNull returns true if the Value is not set, or is explicitly set to null.
func (a Regexp) IsNull() bool {
	return a.state == attr.ValueStateNull
}

// IsUnknown returns true if the Value is not yet known.
func (a Regexp) IsUnknown() bool {
	return a.state == attr.ValueStateUnknown
}

// String returns a summary representation of either the underlying Value,
// or UnknownValueString (`<unknown>`) when IsUnknown() returns true,
// or NullValueString (`<null>`) when IsNull() return true.
//
// This is an intentionally lossy representation, that are best suited for
// logging and error reporting, as they are not protected by
// compatibility guarantees within the framework.
func (a Regexp) String() string {
	if a.IsUnknown() {
		return attr.UnknownValueString
	}

	if a.IsNull() {
		return attr.NullValueString
	}

	return a.value.String()
}

// ValueRegexp returns the known *regexp.Regexp value. If Regexp is null or unknown, returns nil.
func (a Regexp) ValueRegexp() *regexp.Regexp {
	return a.value
}
