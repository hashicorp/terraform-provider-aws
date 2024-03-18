// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
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
	_ xattr.TypeWithValidate   = (*stringEnumType[dummyValueser])(nil)
	_ basetypes.StringTypable  = (*stringEnumType[dummyValueser])(nil)
	_ basetypes.StringValuable = (*StringEnum[dummyValueser])(nil)
)

type customStringTypeWithValidator struct {
	basetypes.StringType
	validator validator.String
}

func (t customStringTypeWithValidator) Validate(ctx context.Context, in tftypes.Value, path path.Path) diag.Diagnostics {
	var diags diag.Diagnostics

	if in.IsNull() || !in.IsKnown() {
		return diags
	}

	var value string
	err := in.As(&value)
	if err != nil {
		diags.AddAttributeError(
			path,
			"Invalid Terraform Value",
			"An unexpected error occurred while attempting to convert a Terraform value to a string. "+
				"This generally is an issue with the provider schema implementation. "+
				"Please contact the provider developers.\n\n"+
				"Path: "+path.String()+"\n"+
				"Error: "+err.Error(),
		)
		return diags
	}

	request := validator.StringRequest{
		ConfigValue: types.StringValue(value),
		Path:        path,
	}
	response := validator.StringResponse{}
	t.validator.ValidateString(ctx, request, &response)
	diags.Append(response.Diagnostics...)

	return diags
}

type stringEnumTypeWithAttributeDefault[T enum.Valueser[T]] interface {
	basetypes.StringTypable
	AttributeDefault(T) defaults.String
}

type stringEnumType[T enum.Valueser[T]] struct {
	customStringTypeWithValidator
}

func StringEnumType[T enum.Valueser[T]]() stringEnumTypeWithAttributeDefault[T] {
	return stringEnumType[T]{customStringTypeWithValidator: customStringTypeWithValidator{validator: stringvalidator.OneOf(tfslices.AppendUnique(enum.Values[T](), "")...)}}
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

func StringEnumNull[T enum.Valueser[T]]() StringEnum[T] {
	return StringEnum[T]{StringValue: basetypes.NewStringNull()}
}

func StringEnumUnknown[T enum.Valueser[T]]() StringEnum[T] {
	return StringEnum[T]{StringValue: basetypes.NewStringUnknown()}
}

func StringEnumValue[T enum.Valueser[T]](value T) StringEnum[T] {
	return StringEnum[T]{StringValue: basetypes.NewStringValue(string(value))}
}

type StringEnum[T enum.Valueser[T]] struct {
	basetypes.StringValue
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
