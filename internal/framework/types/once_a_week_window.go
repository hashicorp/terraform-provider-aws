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
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ xattr.TypeWithValidate                     = (*onceAWeekWindowType)(nil)
	_ basetypes.StringTypable                    = (*onceAWeekWindowType)(nil)
	_ basetypes.StringValuable                   = (*OnceAWeekWindow)(nil)
	_ basetypes.StringValuableWithSemanticEquals = (*OnceAWeekWindow)(nil)
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

func (t onceAWeekWindowType) Validate(ctx context.Context, in tftypes.Value, path path.Path) diag.Diagnostics {
	var diags diag.Diagnostics

	if !in.IsKnown() || in.IsNull() {
		return diags
	}

	var value string
	err := in.As(&value)
	if err != nil {
		diags.AddAttributeError(
			path,
			"OnceAWeekWindowType Validation Error",
			ProviderErrorDetailPrefix+fmt.Sprintf("Cannot convert value to string: %s", err),
		)
		return diags
	}

	// Valid time format is "ddd:hh24:mi".
	validTimeFormat := "(sun|mon|tue|wed|thu|fri|sat):([0-1][0-9]|2[0-3]):([0-5][0-9])"
	validTimeFormatConsolidated := "^(" + validTimeFormat + "-" + validTimeFormat + "|)$"

	if v := strings.ToLower(value); !regexache.MustCompile(validTimeFormatConsolidated).MatchString(v) {
		diags.AddAttributeError(
			path,
			"OnceAWeekWindowType Validation Error",
			fmt.Sprintf("Value %q must satisfy the format of \"ddd:hh24:mi-ddd:hh24:mi\".", value),
		)
		return diags
	}

	return diags
}

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
