// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	smithyjson "github.com/hashicorp/terraform-provider-aws/internal/json"
)

var (
	_ basetypes.StringTypable = (*SmithyJSONType[smithyjson.JSONStringer])(nil)
)

type SmithyJSONType[T smithyjson.JSONStringer] struct {
	basetypes.StringType
	f func(any) T
}

func NewSmithyJSONType[T smithyjson.JSONStringer](_ context.Context, f func(any) T) SmithyJSONType[T] {
	return SmithyJSONType[T]{
		f: f,
	}
}

// String returns a human readable string of the type name.
func (t SmithyJSONType[T]) String() string {
	return "fwtypes.SmithyJSONType"
}

// ValueType returns the Value type.
func (t SmithyJSONType[T]) ValueType(context.Context) attr.Value {
	return SmithyJSON[T]{}
}

// Equal returns true if the given type is equivalent.
func (t SmithyJSONType[T]) Equal(o attr.Type) bool {
	other, ok := o.(SmithyJSONType[T])

	if !ok {
		return false
	}

	return t.StringType.Equal(other.StringType)
}

func (t SmithyJSONType[T]) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
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

func (t SmithyJSONType[T]) ValueFromString(ctx context.Context, in basetypes.StringValue) (basetypes.StringValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	if in.IsNull() {
		return SmithyJSONNull[T](), diags
	}

	if in.IsUnknown() {
		return SmithyJSONUnknown[T](), diags
	}

	var data map[string]any
	if err := json.Unmarshal([]byte(in.ValueString()), &data); err != nil {
		return SmithyJSONUnknown[T](), diags
	}

	return SmithyJSONValue[T](in.ValueString(), t.f), diags
}

var (
	_ basetypes.StringValuable                   = (*SmithyJSON[smithyjson.JSONStringer])(nil)
	_ basetypes.StringValuableWithSemanticEquals = (*SmithyJSON[smithyjson.JSONStringer])(nil)
	_ xattr.ValidateableAttribute                = (*SmithyJSON[smithyjson.JSONStringer])(nil)
)

type SmithyJSON[T smithyjson.JSONStringer] struct {
	basetypes.StringValue
	validate func(context.Context, tftypes.Value, path.Path) diag.Diagnostics
	f        func(any) T
}

func (v SmithyJSON[T]) Equal(o attr.Value) bool {
	other, ok := o.(SmithyJSON[T])

	if !ok {
		return false
	}

	return v.StringValue.Equal(other.StringValue)
}

func (v SmithyJSON[T]) ValueInterface() (T, diag.Diagnostics) {
	var diags diag.Diagnostics

	var zero T
	if v.IsNull() || v.IsUnknown() {
		return zero, diags
	}

	var data map[string]any
	err := json.Unmarshal([]byte(v.ValueString()), &data)

	if err != nil {
		diags.AddError(
			"JSON Unmarshal Error",
			"An unexpected error occurred while unmarshalling a JSON string. "+
				"Please report this to the provider developers.\n\n"+
				"Error: "+err.Error(),
		)
		return zero, diags
	}

	return v.f(data), diags
}

func (v SmithyJSON[T]) Type(context.Context) attr.Type {
	return SmithyJSONType[T]{}
}

func (v SmithyJSON[T]) ValidateAttribute(ctx context.Context, req xattr.ValidateAttributeRequest, resp *xattr.ValidateAttributeResponse) {
	if v.IsNull() || v.IsUnknown() {
		return
	}

	jt, err := jsontypes.NewNormalizedValue(v.ValueString()).ToTerraformValue(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"JSON Normalized Type Validation Error",
			"An unexpected error was encountered trying to validate an attribute value. This is always an error in the provider. "+
				"Please report the following to the provider developer:\n\n"+err.Error(),
		)
		return
	}

	resp.Diagnostics.Append(v.validate(ctx, jt, req.Path)...)
}

func (v SmithyJSON[T]) StringSemanticEquals(ctx context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	oldString := jsontypes.NewNormalizedValue(v.ValueString())

	newValue, ok := newValuable.(SmithyJSON[T])
	if !ok {
		diags.AddError(
			"Semantic Equality Check Error",
			"An unexpected value type was received while performing semantic equality checks. "+
				"Please report this to the provider developers.\n\n"+
				"Expected Value Type: "+fmt.Sprintf("%T", v)+"\n"+
				"Got Value Type: "+fmt.Sprintf("%T", newValuable),
		)

		return false, diags
	}
	newString := jsontypes.NewNormalizedValue(newValue.ValueString())

	result, err := oldString.StringSemanticEquals(ctx, newString)
	diags.Append(err...)

	if diags.HasError() {
		return false, diags
	}

	return result, diags
}

func SmithyJSONValue[T smithyjson.JSONStringer](value string, f func(any) T) SmithyJSON[T] {
	return SmithyJSON[T]{
		StringValue: basetypes.NewStringValue(value),
		validate:    jsontypes.NormalizedType{}.Validate,
		f:           f,
	}
}
func SmithyJSONNull[T smithyjson.JSONStringer]() SmithyJSON[T] {
	return SmithyJSON[T]{
		StringValue: basetypes.NewStringNull(),
	}
}

func SmithyJSONUnknown[T smithyjson.JSONStringer]() SmithyJSON[T] {
	return SmithyJSON[T]{
		StringValue: basetypes.NewStringUnknown(),
	}
}
