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
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ basetypes.StringTypable = (*durationType)(nil)
)

type durationType struct {
	basetypes.StringType
}

var (
	DurationType = durationType{}
)

func (t durationType) Equal(o attr.Type) bool {
	other, ok := o.(durationType)

	if !ok {
		return false
	}

	return t.StringType.Equal(other.StringType)
}

func (durationType) String() string {
	return "DurationType"
}

func (t durationType) ValueFromString(_ context.Context, in types.String) (basetypes.StringValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	if in.IsNull() {
		return DurationNull(), diags
	}
	if in.IsUnknown() {
		return DurationUnknown(), diags
	}

	valueString := in.ValueString()
	if _, err := time.ParseDuration(valueString); err != nil {
		return DurationUnknown(), diags // Must not return validation errors
	}

	return DurationValue(valueString), diags
}

func (t durationType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
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

func (durationType) ValueType(context.Context) attr.Value {
	return Duration{}
}

var (
	_ basetypes.StringValuable    = (*Duration)(nil)
	_ xattr.ValidateableAttribute = (*Duration)(nil)
)

func DurationNull() Duration {
	return Duration{StringValue: basetypes.NewStringNull()}
}

func DurationUnknown() Duration {
	return Duration{StringValue: basetypes.NewStringUnknown()}
}

// DurationValue initializes a new Duration type with the provided value
//
// This function does not return diagnostics, and therefore invalid duration values
// are not handled during construction. Invalid values will be detected by the
// ValidateAttribute method, called by the ValidateResourceConfig RPC during
// operations like `terraform validate`, `plan`, or `apply`.
func DurationValue(value string) Duration {
	// swallow any Duration parsing errors here and just pass along the
	// zero value time.Duration. Invalid values will be handled downstream
	// by the ValidateAttribute method.
	v, _ := time.ParseDuration(value)

	return Duration{
		StringValue: basetypes.NewStringValue(value),
		value:       v,
	}
}

type Duration struct {
	basetypes.StringValue
	value time.Duration
}

func (v Duration) Equal(o attr.Value) bool {
	other, ok := o.(Duration)

	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

func (Duration) Type(context.Context) attr.Type {
	return DurationType
}

// ValueDuration returns the known time.Duration value. If Duration is null or unknown, returns 0.
func (v Duration) ValueDuration() time.Duration {
	return v.value
}

func (v Duration) ValidateAttribute(ctx context.Context, req xattr.ValidateAttributeRequest, resp *xattr.ValidateAttributeResponse) {
	if v.IsNull() || v.IsUnknown() {
		return
	}

	if _, err := time.ParseDuration(v.ValueString()); err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Duration Value",
			"The provided value cannot be parsed as a Duration.\n\n"+
				"Path: "+req.Path.String()+"\n"+
				"Error: "+err.Error(),
		)
	}
}
