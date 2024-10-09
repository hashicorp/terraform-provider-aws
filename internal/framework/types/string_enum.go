// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
)

type dummyValueser string

func (dummyValueser) Values() []dummyValueser {
	return nil
}

var (
	_ basetypes.StringTypable = (*stringEnumType[dummyValueser])(nil)
)

type stringEnumTypeWithAttributeDefault[T enum.Valueser[T]] interface {
	basetypes.StringTypable
	AttributeDefault(T) defaults.String
}

type stringEnumType[T enum.Valueser[T]] struct {
	basetypes.StringType
}

func StringEnumType[T enum.Valueser[T]]() stringEnumTypeWithAttributeDefault[T] {
	return stringEnumType[T]{}
}

func (t stringEnumType[T]) Equal(o attr.Type) bool {
	other, ok := o.(stringEnumType[T])

	if !ok {
		return false
	}

	return t.StringType.Equal(other.StringType)
}

func (stringEnumType[T]) String() string {
	var zero T
	// The format of this returned value is used inside AutoFlEx.
	return fmt.Sprintf("StringEnumType[%T]", zero)
}

func (t stringEnumType[T]) ValueFromString(_ context.Context, in types.String) (basetypes.StringValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	if in.IsNull() {
		return StringEnumNull[T](), diags
	}
	if in.IsUnknown() {
		return StringEnumUnknown[T](), diags
	}

	return StringEnum[T]{StringValue: in}, diags
}

func (t stringEnumType[T]) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
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

func (t stringEnumType[T]) ValueType(context.Context) attr.Value {
	return StringEnum[T]{}
}

func (t stringEnumType[T]) AttributeDefault(defaultVal T) defaults.String {
	return stringdefault.StaticString(string(defaultVal))
}

var (
	_ basetypes.StringValuable    = (*StringEnum[dummyValueser])(nil)
	_ xattr.ValidateableAttribute = (*StringEnum[dummyValueser])(nil)
)

type StringEnum[T enum.Valueser[T]] struct {
	basetypes.StringValue
}

func StringEnumNull[T enum.Valueser[T]]() StringEnum[T] {
	return StringEnum[T]{StringValue: basetypes.NewStringNull()}
}

func StringEnumUnknown[T enum.Valueser[T]]() StringEnum[T] {
	return StringEnum[T]{StringValue: basetypes.NewStringUnknown()}
}

func StringEnumValue[T enum.Valueser[T]](value T) StringEnum[T] {
	return StringEnum[T]{StringValue: basetypes.NewStringValue(string(value))}
}

func StringEnumValueToUpper[T enum.Valueser[T]](value T) StringEnum[T] {
	return StringEnumValue(T(strings.ToUpper(string(value))))
}

func (v StringEnum[T]) Equal(o attr.Value) bool {
	other, ok := o.(StringEnum[T])

	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

func (v StringEnum[T]) Type(context.Context) attr.Type {
	return StringEnumType[T]()
}

func (v StringEnum[T]) ValueEnum() T {
	return T(v.ValueString())
}

// StringEnumValue is useful if you have a zero value StringEnum but need a
// way to get a non-zero value such as when flattening.
// It's called via reflection inside AutoFlEx.
func (v StringEnum[T]) StringEnumValue(value string) StringEnum[T] {
	return StringEnum[T]{StringValue: basetypes.NewStringValue(value)}
}

func (v StringEnum[T]) ValidateAttribute(ctx context.Context, req xattr.ValidateAttributeRequest, resp *xattr.ValidateAttributeResponse) {
	if v.IsNull() || v.IsUnknown() {
		return
	}

	vs := v.ValueString()
	validValues := tfslices.AppendUnique(v.ValueEnum().Values(), "")

	for _, enumVal := range validValues {
		if vs == string(enumVal) {
			return
		}
	}

	resp.Diagnostics.AddAttributeError(
		req.Path,
		"Invalid String Enum Value",
		"The provided value does not match any valid values.\n\n"+
			"Path: "+req.Path.String()+"\n"+
			"Given Value: "+vs+"\n"+
			"Valid Values: "+fmt.Sprintf("%s", validValues),
	)
}
