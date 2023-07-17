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
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

type TimestampType struct {
	basetypes.StringType
}

var (
	_ basetypes.StringTypable = TimestampType{}
	_ xattr.TypeWithValidate  = TimestampType{}
)

func (typ TimestampType) ValueFromString(_ context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	if in.IsUnknown() {
		return NewTimestampUnknown(), nil
	}

	if in.IsNull() {
		return NewTimestampNull(), nil
	}

	s := in.ValueString()

	var diags diag.Diagnostics
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		diags.AddError(
			"Timestamp Type Validation Error",
			fmt.Sprintf("Value %q cannot be parsed as a Timestamp.", s),
		)
		return nil, diags
	}

	return newTimestampValue(s, t), nil
}

func (typ TimestampType) ValueFromTerraform(_ context.Context, in tftypes.Value) (attr.Value, error) {
	if !in.IsKnown() {
		return NewTimestampUnknown(), nil
	}

	if in.IsNull() {
		return NewTimestampNull(), nil
	}

	var s string
	err := in.As(&s)
	if err != nil {
		return nil, err
	}

	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return NewTimestampUnknown(), nil //nolint: nilerr // Must not return validation errors
	}

	return newTimestampValue(s, t), nil
}

func (typ TimestampType) ValueType(context.Context) attr.Value {
	return TimestampValue{}
}

func (typ TimestampType) Equal(o attr.Type) bool {
	other, ok := o.(TimestampType)
	if !ok {
		return false
	}

	return typ.StringType.Equal(other.StringType)
}

// String returns a human-friendly description of the TimestampType.
func (typ TimestampType) String() string {
	return "types.TimestampType"
}

func (typ TimestampType) Validate(ctx context.Context, in tftypes.Value, path path.Path) diag.Diagnostics {
	var diags diag.Diagnostics

	if !in.IsKnown() || in.IsNull() {
		return diags
	}

	var s string
	err := in.As(&s)
	if err != nil {
		diags.AddAttributeError(
			path,
			"Invalid Terraform Value",
			"An unexpected error occurred while attempting to convert a Terraform value to a string. "+
				"This is generally an issue with the provider schema implementation. "+
				"Please report the following to the provider developer:\n\n"+
				"Path: "+path.String()+"\n"+
				"Error: "+err.Error(),
		)
		return diags
	}

	_, err = time.Parse(time.RFC3339, s)
	if err != nil {
		diags.AddAttributeError(
			path,
			"Invalid Timestamp Value",
			fmt.Sprintf("Value %q cannot be parsed as an RFC 3339 Timestamp.\n\n"+
				"Path: %s\n"+
				"Error: %s", s, path, err),
		)
		return diags
	}

	return diags
}
