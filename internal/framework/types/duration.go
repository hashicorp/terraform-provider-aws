// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type durationType uint8

const (
	DurationType durationType = iota
)

var (
	_ xattr.TypeWithValidate = DurationType
)

func (d durationType) TerraformType(_ context.Context) tftypes.Type {
	return tftypes.String
}

func (d durationType) ValueFromString(_ context.Context, in types.String) (basetypes.StringValuable, diag.Diagnostics) {
	if in.IsUnknown() {
		return DurationUnknown(), nil
	}

	if in.IsNull() {
		return DurationNull(), nil
	}

	var diags diag.Diagnostics
	v, err := time.ParseDuration(in.ValueString())
	if err != nil {
		diags.AddError(
			"Duration Type Validation Error",
			fmt.Sprintf("Value %q cannot be parsed as a Duration.", in.ValueString()),
		)
		return nil, diags
	}

	return DurationValue(v), nil
}

func (d durationType) ValueFromTerraform(_ context.Context, in tftypes.Value) (attr.Value, error) {
	if !in.IsKnown() {
		return DurationUnknown(), nil
	}

	if in.IsNull() {
		return DurationNull(), nil
	}

	var s string
	err := in.As(&s)

	if err != nil {
		return nil, err
	}

	v, err := time.ParseDuration(s)

	if err != nil {
		return DurationUnknown(), nil //nolint: nilerr // Must not return validation errors
	}

	return DurationValue(v), nil
}

func (d durationType) ValueType(context.Context) attr.Value {
	return Duration{}
}

// Equal returns true if `o` is also a DurationType.
func (d durationType) Equal(o attr.Type) bool {
	_, ok := o.(durationType)
	return ok
}

// ApplyTerraform5AttributePathStep applies the given AttributePathStep to the
// type.
func (d durationType) ApplyTerraform5AttributePathStep(step tftypes.AttributePathStep) (interface{}, error) {
	return nil, fmt.Errorf("cannot apply AttributePathStep %T to %s", step, d.String())
}

// String returns a human-friendly description of the DurationType.
func (d durationType) String() string {
	return "types.DurationType"
}

// Validate implements type validation.
func (d durationType) Validate(ctx context.Context, in tftypes.Value, path path.Path) diag.Diagnostics {
	var diags diag.Diagnostics

	if !in.Type().Is(tftypes.String) {
		diags.AddAttributeError(
			path,
			"Duration Type Validation Error",
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
			"Duration Type Validation Error",
			"An unexpected error was encountered trying to validate an attribute value. This is always an error in the provider. Please report the following to the provider developer:\n\n"+
				fmt.Sprintf("Cannot convert value to time.Duration: %s", err),
		)
		return diags
	}

	_, err = time.ParseDuration(value)
	if err != nil {
		diags.AddAttributeError(
			path,
			"Duration Type Validation Error",
			fmt.Sprintf("Value %q cannot be parsed as a Duration.", value),
		)
		return diags
	}

	return diags
}

func (d durationType) Description() string {
	return `A sequence of numbers with a unit suffix, "h" for hour, "m" for minute, and "s" for second.`
}

func DurationNull() Duration {
	return Duration{
		state: attr.ValueStateNull,
	}
}

func DurationUnknown() Duration {
	return Duration{
		state: attr.ValueStateUnknown,
	}
}

func DurationValue(value time.Duration) Duration {
	return Duration{
		state: attr.ValueStateKnown,
		value: value,
	}
}

type Duration struct {
	// state represents whether the value is null, unknown, or known. The
	// zero-value is null.
	state attr.ValueState

	// value contains the known value, if not null or unknown.
	value time.Duration
}

// Type returns a DurationType.
func (d Duration) Type(_ context.Context) attr.Type {
	return DurationType
}

func (d Duration) ToStringValue(ctx context.Context) (types.String, diag.Diagnostics) {
	switch d.state {
	case attr.ValueStateKnown:
		return types.StringValue(d.value.String()), nil
	case attr.ValueStateNull:
		return types.StringNull(), nil
	case attr.ValueStateUnknown:
		return types.StringUnknown(), nil
	default:
		return types.StringUnknown(), diag.Diagnostics{
			diag.NewErrorDiagnostic(fmt.Sprintf("unhandled Duration state in ToStringValue: %s", d.state), ""),
		}
	}
}

// ToTerraformValue returns the data contained in the *String as a string. If
// Unknown is true, it returns a tftypes.UnknownValue. If Null is true, it
// returns nil.
func (d Duration) ToTerraformValue(ctx context.Context) (tftypes.Value, error) {
	t := DurationType.TerraformType(ctx)

	switch d.state {
	case attr.ValueStateKnown:
		if err := tftypes.ValidateValue(t, d.value); err != nil {
			return tftypes.NewValue(t, tftypes.UnknownValue), err
		}

		return tftypes.NewValue(t, d.value), nil
	case attr.ValueStateNull:
		return tftypes.NewValue(t, nil), nil
	case attr.ValueStateUnknown:
		return tftypes.NewValue(t, tftypes.UnknownValue), nil
	default:
		return tftypes.NewValue(t, tftypes.UnknownValue), fmt.Errorf("unhandled Duration state in ToTerraformValue: %s", d.state)
	}
}

// Equal returns true if `other` is a *Duration and has the same value as `d`.
func (d Duration) Equal(other attr.Value) bool {
	o, ok := other.(Duration)

	if !ok {
		return false
	}

	if d.state != o.state {
		return false
	}

	if d.state != attr.ValueStateKnown {
		return true
	}

	return d.value == o.value
}

// IsNull returns true if the Value is not set, or is explicitly set to null.
func (d Duration) IsNull() bool {
	return d.state == attr.ValueStateNull
}

// IsUnknown returns true if the Value is not yet known.
func (d Duration) IsUnknown() bool {
	return d.state == attr.ValueStateUnknown
}

// String returns a summary representation of either the underlying Value,
// or UnknownValueString (`<unknown>`) when IsUnknown() returns true,
// or NullValueString (`<null>`) when IsNull() return true.
//
// This is an intentionally lossy representation, that are best suited for
// logging and error reporting, as they are not protected by
// compatibility guarantees within the framework.
func (d Duration) String() string {
	if d.IsUnknown() {
		return attr.UnknownValueString
	}

	if d.IsNull() {
		return attr.NullValueString
	}

	return d.value.String()
}

// ValueDuration returns the known time.Duration value. If Duration is null or unknown, returns 0.
func (d Duration) ValueDuration() time.Duration {
	return d.value
}
