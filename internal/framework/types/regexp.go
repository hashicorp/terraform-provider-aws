// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"context"
	"fmt"
	"regexp"

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
	_ xattr.TypeWithValidate   = (*regexpType)(nil)
	_ basetypes.StringTypable  = (*regexpType)(nil)
	_ basetypes.StringValuable = (*Regexp)(nil)
)

type regexpType struct {
	basetypes.StringType
}

var (
	RegexpType = regexpType{}
)

func (t regexpType) Equal(o attr.Type) bool {
	other, ok := o.(regexpType)

	if !ok {
		return false
	}

	return t.StringType.Equal(other.StringType)
}

func (regexpType) String() string {
	return "RegexpType"
}

func (t regexpType) ValueFromString(_ context.Context, in types.String) (basetypes.StringValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	if in.IsNull() {
		return RegexpNull(), diags
	}
	if in.IsUnknown() {
		return RegexpUnknown(), diags
	}

	valueString := in.ValueString()
	if _, err := regexp.Compile(valueString); err != nil {
		return RegexpUnknown(), diags // Must not return validation errors.
	}

	return RegexpValueMust(valueString), diags
}

func (t regexpType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
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

func (regexpType) ValueType(context.Context) attr.Value {
	return Regexp{}
}

func (t regexpType) Validate(_ context.Context, in tftypes.Value, path path.Path) diag.Diagnostics {
	var diags diag.Diagnostics

	if !in.IsKnown() || in.IsNull() {
		return diags
	}

	var value string
	err := in.As(&value)
	if err != nil {
		diags.AddAttributeError(
			path,
			"Regexp Type Validation Error",
			ProviderErrorDetailPrefix+fmt.Sprintf("Cannot convert value to string: %s", err),
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

func RegexpNull() Regexp {
	return Regexp{StringValue: basetypes.NewStringNull()}
}

func RegexpUnknown() Regexp {
	return Regexp{StringValue: basetypes.NewStringUnknown()}
}

func RegexpValueMust(value string) Regexp {
	return Regexp{
		StringValue: basetypes.NewStringValue(value),
		value:       regexache.MustCompile(value),
	}
}

type Regexp struct {
	basetypes.StringValue
	value *regexp.Regexp
}

func (v Regexp) Equal(o attr.Value) bool {
	other, ok := o.(Regexp)

	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

func (Regexp) Type(context.Context) attr.Type {
	return RegexpType
}

func (v Regexp) ValueRegexp() *regexp.Regexp {
	return v.value
}
