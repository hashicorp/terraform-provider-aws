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
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ basetypes.StringTypable = (*regexpType)(nil)
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

	return RegexpValue(valueString), diags
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

var (
	_ basetypes.StringValuable    = (*Regexp)(nil)
	_ xattr.ValidateableAttribute = (*Regexp)(nil)
)

func RegexpNull() Regexp {
	return Regexp{StringValue: basetypes.NewStringNull()}
}

func RegexpUnknown() Regexp {
	return Regexp{StringValue: basetypes.NewStringUnknown()}
}

// RegexpValue initializes a new Regexp type with the provided value
//
// This function does not return diagnostics, and therefore invalid regular expression values
// are not handled during construction. Invalid values will be detected by the
// ValidateAttribute method, called by the ValidateResourceConfig RPC during
// operations like `terraform validate`, `plan`, or `apply`.
func RegexpValue(value string) Regexp {
	// swallow any regex parsing errors here and just pass along the
	// zero value regexp.Regexp. Invalid values will be handled downstream
	// by the ValidateAttribute method.
	v, _ := regexp.Compile(value)

	return Regexp{
		StringValue: basetypes.NewStringValue(value),
		value:       v,
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

func (v Regexp) ValidateAttribute(ctx context.Context, req xattr.ValidateAttributeRequest, resp *xattr.ValidateAttributeResponse) {
	if v.IsNull() || v.IsUnknown() {
		return
	}

	vs := v.ValueString()
	if _, err := regexp.Compile(vs); err != nil {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid Regexp Value",
			"The provided value cannot be parsed as a regular expression.\n\n"+
				"Path: "+req.Path.String()+"\n"+
				"Value: "+vs+"\n"+
				"Error: "+err.Error(),
		)
	}
}
