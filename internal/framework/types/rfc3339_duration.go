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
	"github.com/hashicorp/terraform-provider-aws/internal/types/duration"
)

var (
	_ basetypes.StringTypable = (*rfc3339DurationType)(nil)
)

type rfc3339DurationType struct {
	basetypes.StringType
}

var (
	RFC3339DurationType = rfc3339DurationType{}
)

func (t rfc3339DurationType) Equal(o attr.Type) bool {
	other, ok := o.(rfc3339DurationType)

	if !ok {
		return false
	}

	return t.StringType.Equal(other.StringType)
}

func (rfc3339DurationType) String() string {
	return "RFC3339DurationType"
}

func (t rfc3339DurationType) ValueFromString(_ context.Context, in types.String) (basetypes.StringValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	if in.IsNull() {
		return RFC3339DurationNull(), diags
	}
	if in.IsUnknown() {
		return RFC3339DurationUnknown(), diags
	}

	valueString := in.ValueString()
	if _, err := duration.Parse(valueString); err != nil {
		return RFC3339DurationUnknown(), diags // Must not return validation errors
	}

	return RFC3339DurationValue(valueString), diags
}

func (t rfc3339DurationType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
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

func (rfc3339DurationType) ValueType(context.Context) attr.Value {
	return RFC3339Duration{}
}

var (
	_ basetypes.StringValuable    = (*RFC3339Duration)(nil)
	_ xattr.ValidateableAttribute = (*RFC3339Duration)(nil)
)

func RFC3339DurationNull() RFC3339Duration {
	return RFC3339Duration{StringValue: basetypes.NewStringNull()}
}

func RFC3339DurationUnknown() RFC3339Duration {
	return RFC3339Duration{StringValue: basetypes.NewStringUnknown()}
}

// DurationValue initializes a new RFC3339Duration type with the provided value
//
// This function does not return diagnostics, and therefore invalid duration values
// are not handled during construction. Invalid values will be detected by the
// ValidateAttribute method, called by the ValidateResourceConfig RPC during
// operations like `terraform validate`, `plan`, or `apply`.
func RFC3339DurationValue(value string) RFC3339Duration {
	// swallow any RFC3339Duration parsing errors here and just pass along the
	// zero value duration.Duration. Invalid values will be handled downstream
	// by the ValidateAttribute method.
	v, _ := duration.Parse(value)

	return RFC3339Duration{
		StringValue: basetypes.NewStringValue(value),
		value:       v,
	}
}

func RFC3339DurationTimeDurationValue(value time.Duration) RFC3339Duration {
	v := duration.NewFromTimeDuration(value)

	return RFC3339Duration{
		StringValue: basetypes.NewStringValue(v.String()),
		value:       v,
	}
}

type RFC3339Duration struct {
	basetypes.StringValue
	value duration.Duration
}

func (v RFC3339Duration) Equal(o attr.Value) bool {
	other, ok := o.(RFC3339Duration)

	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

func (RFC3339Duration) Type(context.Context) attr.Type {
	return RFC3339DurationType
}

// ValueDuration returns the known duration.Duration value. If RFC3339Duration is null or unknown, returns 0.
func (v RFC3339Duration) ValueDuration() duration.Duration {
	return v.value
}

func (v RFC3339Duration) ValidateAttribute(ctx context.Context, req xattr.ValidateAttributeRequest, resp *xattr.ValidateAttributeResponse) {
	if v.IsNull() || v.IsUnknown() {
		return
	}

	if _, err := duration.Parse(v.ValueString()); err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Duration Value",
			"The provided value cannot be parsed as a Duration.\n\n"+
				"Path: "+req.Path.String()+"\n"+
				"Error: "+err.Error(),
		)
	}
}
