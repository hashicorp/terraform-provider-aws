// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"context"
	"fmt"
	"strings"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

const (
	// Valid time format is "ddd:hh24:mi".
	validTimeFormat             = "(sun|mon|tue|wed|thu|fri|sat):([0-1][0-9]|2[0-3]):([0-5][0-9])"
	validTimeFormatConsolidated = "^(" + validTimeFormat + "-" + validTimeFormat + "|)$"
)

var (
	validTimeFormatConsolidatedRegex = regexache.MustCompile(validTimeFormatConsolidated)
)

var (
	_ basetypes.StringTypable = (*onceAWeekWindowType)(nil)
)

type onceAWeekWindowType struct {
	basetypes.StringType
}

var (
	OnceAWeekWindowType = onceAWeekWindowType{}
)

func (t onceAWeekWindowType) Equal(o attr.Type) bool {
	other, ok := o.(onceAWeekWindowType)
	if !ok {
		return false
	}

	return t.StringType.Equal(other.StringType)
}

func (onceAWeekWindowType) String() string {
	return "OnceAWeekWindowType"
}

func (t onceAWeekWindowType) ValueFromString(_ context.Context, in types.String) (basetypes.StringValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	if in.IsNull() {
		return OnceAWeekWindowNull(), diags
	}
	if in.IsUnknown() {
		return OnceAWeekWindowUnknown(), diags
	}

	return OnceAWeekWindowValue(in.ValueString()), diags
}

func (t onceAWeekWindowType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
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

func (onceAWeekWindowType) ValueType(context.Context) attr.Value {
	return OnceAWeekWindow{}
}

var (
	_ basetypes.StringValuable                   = (*OnceAWeekWindow)(nil)
	_ basetypes.StringValuableWithSemanticEquals = (*OnceAWeekWindow)(nil)
	_ xattr.ValidateableAttribute                = (*OnceAWeekWindow)(nil)
)

type OnceAWeekWindow struct {
	basetypes.StringValue
}

func OnceAWeekWindowNull() OnceAWeekWindow {
	return OnceAWeekWindow{StringValue: basetypes.NewStringNull()}
}

func OnceAWeekWindowUnknown() OnceAWeekWindow {
	return OnceAWeekWindow{StringValue: basetypes.NewStringUnknown()}
}

func OnceAWeekWindowValue(value string) OnceAWeekWindow {
	return OnceAWeekWindow{StringValue: basetypes.NewStringValue(value)}
}

func (v OnceAWeekWindow) Equal(o attr.Value) bool {
	other, ok := o.(OnceAWeekWindow)
	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

func (OnceAWeekWindow) Type(context.Context) attr.Type {
	return OnceAWeekWindowType
}

func (v OnceAWeekWindow) StringSemanticEquals(ctx context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	newValue, ok := newValuable.(OnceAWeekWindow)
	if !ok {
		return false, diags
	}

	old, d := v.ToStringValue(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return false, diags
	}

	new, d := newValue.ToStringValue(ctx)
	diags.Append(d...)
	if diags.HasError() {
		return false, diags
	}

	// Case insensitive comparison.
	return strings.EqualFold(old.ValueString(), new.ValueString()), diags
}

func (v OnceAWeekWindow) ValidateAttribute(ctx context.Context, req xattr.ValidateAttributeRequest, resp *xattr.ValidateAttributeResponse) {
	if v.IsNull() || v.IsUnknown() {
		return
	}

	if vs := strings.ToLower(v.ValueString()); !validTimeFormatConsolidatedRegex.MatchString(vs) {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Once A Week Window Value",
			"The provided value does not satisfy the format \"ddd:hh24:mi-ddd:hh24:mi\".\n\n"+
				"Path: "+req.Path.String()+"\n"+
				"Value: "+vs,
		)
	}
}
