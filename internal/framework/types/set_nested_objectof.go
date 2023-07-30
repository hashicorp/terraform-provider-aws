// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
)

// SetNestedObjectTypeOf is the attribute type of a SetNestedObjectValueOf.
type SetNestedObjectTypeOf[T any] struct {
	basetypes.SetType
}

var (
	_ basetypes.SetTypable = SetNestedObjectTypeOf[struct{}]{}
	_ NestedObjectType     = SetNestedObjectTypeOf[struct{}]{}
)

func NewSetNestedObjectTypeOf[T any](ctx context.Context) SetNestedObjectTypeOf[T] {
	return SetNestedObjectTypeOf[T]{basetypes.SetType{ElemType: NewObjectTypeOf[T](ctx)}}
}

func (t SetNestedObjectTypeOf[T]) Equal(o attr.Type) bool {
	other, ok := o.(SetNestedObjectTypeOf[T])

	if !ok {
		return false
	}

	return t.SetType.Equal(other.SetType)
}

func (t SetNestedObjectTypeOf[T]) String() string {
	var zero T
	return fmt.Sprintf("SetNestedObjectTypeOf[%T]", zero)
}

func (t SetNestedObjectTypeOf[T]) ValueFromSet(ctx context.Context, in basetypes.SetValue) (basetypes.SetValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	if in.IsNull() {
		return NewSetNestedObjectValueOfNull[T](ctx), diags
	}
	if in.IsUnknown() {
		return NewSetNestedObjectValueOfUnknown[T](ctx), diags
	}

	setValue, d := basetypes.NewSetValue(NewObjectTypeOf[T](ctx), in.Elements())
	diags.Append(d...)
	if diags.HasError() {
		return NewSetNestedObjectValueOfUnknown[T](ctx), diags
	}

	value := SetNestedObjectValueOf[T]{
		SetValue: setValue,
	}

	return value, diags
}

func (t SetNestedObjectTypeOf[T]) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.SetType.ValueFromTerraform(ctx, in)

	if err != nil {
		return nil, err
	}

	setValue, ok := attrValue.(basetypes.SetValue)

	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	setValuable, diags := t.ValueFromSet(ctx, setValue)

	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting SetValue to SetValuable: %v", diags)
	}

	return setValuable, nil
}

func (t SetNestedObjectTypeOf[T]) ValueType(ctx context.Context) attr.Value {
	return SetNestedObjectValueOf[T]{}
}

func (t SetNestedObjectTypeOf[T]) NewObjectPtr(ctx context.Context) (any, diag.Diagnostics) {
	return nestedObjectTypeNewObjectPtr[T](ctx)
}

func (t SetNestedObjectTypeOf[T]) NewObjectSlice(ctx context.Context, len, cap int) (any, diag.Diagnostics) {
	return nestedObjectTypeNewObjectSlice[T](ctx, len, cap)
}

func (t SetNestedObjectTypeOf[T]) NullValue(ctx context.Context) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	return NewSetNestedObjectValueOfNull[T](ctx), diags
}

func (t SetNestedObjectTypeOf[T]) ValueFromObjectPtr(ctx context.Context, ptr any) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	if v, ok := ptr.(*T); ok {
		return NewSetNestedObjectValueOfPtr(ctx, v), diags
	}

	diags.Append(diag.NewErrorDiagnostic("Invalid pointer value", fmt.Sprintf("incorrect type: want %T, got %T", (*T)(nil), ptr)))
	return nil, diags
}

func (t SetNestedObjectTypeOf[T]) ValueFromObjectSlice(ctx context.Context, slice any) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	if v, ok := slice.([]*T); ok {
		return NewSetNestedObjectValueOfSlice(ctx, v), diags
	}

	diags.Append(diag.NewErrorDiagnostic("Invalid slice value", fmt.Sprintf("incorrect type: want %T, got %T", (*[]T)(nil), slice)))
	return nil, diags
}

// SetNestedObjectValueOf represents a Terraform Plugin Framework Set value whose elements are of type ObjectTypeOf.
type SetNestedObjectValueOf[T any] struct {
	basetypes.SetValue
}

var (
	_ basetypes.SetValuable = SetNestedObjectValueOf[struct{}]{}
	_ NestedObjectValue     = SetNestedObjectValueOf[struct{}]{}
)

func (v SetNestedObjectValueOf[T]) Equal(o attr.Value) bool {
	other, ok := o.(SetNestedObjectValueOf[T])

	if !ok {
		return false
	}

	return v.SetValue.Equal(other.SetValue)
}

func (v SetNestedObjectValueOf[T]) Type(ctx context.Context) attr.Type {
	return NewSetNestedObjectTypeOf[T](ctx)
}

func (v SetNestedObjectValueOf[T]) ToObjectPtr(ctx context.Context) (any, diag.Diagnostics) {
	return nestedObjectValueObjectPtr[T](ctx, v.SetValue)
}

func (v SetNestedObjectValueOf[T]) ToObjectSlice(ctx context.Context) (any, diag.Diagnostics) {
	return nestedObjectValueObjectSlice[T](ctx, v.SetValue)
}

func NewSetNestedObjectValueOfNull[T any](ctx context.Context) SetNestedObjectValueOf[T] {
	return SetNestedObjectValueOf[T]{SetValue: basetypes.NewSetNull(NewObjectTypeOf[T](ctx))}
}

func NewSetNestedObjectValueOfUnknown[T any](ctx context.Context) SetNestedObjectValueOf[T] {
	return SetNestedObjectValueOf[T]{SetValue: basetypes.NewSetUnknown(NewObjectTypeOf[T](ctx))}
}

func NewSetNestedObjectValueOfPtr[T any](ctx context.Context, t *T) SetNestedObjectValueOf[T] {
	return NewSetNestedObjectValueOfSlice(ctx, []*T{t})
}

func NewSetNestedObjectValueOfSlice[T any](ctx context.Context, ts []*T) SetNestedObjectValueOf[T] {
	return newSetNestedObjectValueOf[T](ctx, ts)
}

func NewSetNestedObjectValueOfValueSlice[T any](ctx context.Context, ts []T) SetNestedObjectValueOf[T] {
	return newSetNestedObjectValueOf[T](ctx, ts)
}

func newSetNestedObjectValueOf[T any](ctx context.Context, elements any) SetNestedObjectValueOf[T] {
	return SetNestedObjectValueOf[T]{SetValue: fwdiag.Must(basetypes.NewSetValueFrom(ctx, NewObjectTypeOf[T](ctx), elements))}
}
