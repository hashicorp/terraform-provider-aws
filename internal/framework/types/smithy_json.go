// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-jsontypes/jsontypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	tfsmithy "github.com/hashicorp/terraform-provider-aws/internal/smithy"
)

var (
	_ basetypes.StringTypable = (*SmithyJSONType[tfsmithy.JSONStringer])(nil)
)

type SmithyJSONType[T tfsmithy.JSONStringer] struct {
	jsontypes.NormalizedType
	f func(any) T
}

func NewSmithyJSONType[T tfsmithy.JSONStringer](_ context.Context, f func(any) T) SmithyJSONType[T] {
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

	return t.NormalizedType.Equal(other.NormalizedType)
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
		return NewSmithyJSONNull[T](), diags
	}

	if in.IsUnknown() {
		return NewSmithyJSONUnknown[T](), diags
	}

	return NewSmithyJSONValue(in.ValueString(), t.f), diags
}

var (
	_ basetypes.StringValuable                   = (*SmithyJSON[tfsmithy.JSONStringer])(nil)
	_ basetypes.StringValuableWithSemanticEquals = (*SmithyJSON[tfsmithy.JSONStringer])(nil)
	_ xattr.ValidateableAttribute                = (*SmithyJSON[tfsmithy.JSONStringer])(nil)
	_ SmithyDocumentValue                        = (*SmithyJSON[tfsmithy.JSONStringer])(nil)
)

type SmithyJSON[T tfsmithy.JSONStringer] struct {
	jsontypes.Normalized
	f func(any) T
}

func (v SmithyJSON[T]) Equal(o attr.Value) bool {
	other, ok := o.(SmithyJSON[T])
	if !ok {
		return false
	}

	return v.Normalized.Equal(other.Normalized)
}

func (v SmithyJSON[T]) ToSmithyObjectDocument(ctx context.Context) (any, diag.Diagnostics) {
	return v.ToSmithyDocument(ctx)
}

func (v SmithyJSON[T]) ToSmithyDocument(context.Context) (T, diag.Diagnostics) {
	var diags diag.Diagnostics

	var zero T
	if v.IsNull() || v.IsUnknown() || v.f == nil {
		return zero, diags
	}

	t, err := tfsmithy.DocumentFromJSONString(v.ValueString(), v.f)
	if err != nil {
		diags.AddError(
			"JSON Unmarshal Error",
			"An unexpected error occurred while unmarshalling a JSON string. "+
				"Please report this to the provider developers.\n\n"+
				"Error: "+err.Error(),
		)
		return zero, diags
	}

	return t, diags
}

func (v SmithyJSON[T]) Type(context.Context) attr.Type {
	return SmithyJSONType[T]{
		f: v.f,
	}
}

func (v SmithyJSON[T]) StringSemanticEquals(ctx context.Context, newValuable basetypes.StringValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

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

	return v.Normalized.StringSemanticEquals(ctx, newValue.Normalized)
}

func NewSmithyJSONValue[T tfsmithy.JSONStringer](value string, f func(any) T) SmithyJSON[T] {
	return SmithyJSON[T]{
		Normalized: jsontypes.NewNormalizedValue(value),
		f:          f,
	}
}

func NewSmithyJSONNull[T tfsmithy.JSONStringer]() SmithyJSON[T] {
	return SmithyJSON[T]{
		Normalized: jsontypes.NewNormalizedNull(),
	}
}

func NewSmithyJSONUnknown[T tfsmithy.JSONStringer]() SmithyJSON[T] {
	return SmithyJSON[T]{
		Normalized: jsontypes.NewNormalizedUnknown(),
	}
}

// SmithyDocumentValue extends the Value interface for values that represent Smithy documents.
// It isn't generic on the Go interface type as it's referenced within AutoFlEx.
type SmithyDocumentValue interface {
	attr.Value

	// ToSmithyObjectDocument returns the value as a Smithy document.
	ToSmithyObjectDocument(context.Context) (any, diag.Diagnostics)
}
