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

// ListNestedObjectTypeOf is the attribute type of a ListNestedObjectValueOf.
type ListNestedObjectTypeOf[T any] struct {
	basetypes.ListType
}

var (
	_ basetypes.ListTypable = ListNestedObjectTypeOf[struct{}]{}
	_ NestedObjectType      = ListNestedObjectTypeOf[struct{}]{}
)

func NewListNestedObjectTypeOf[T any](ctx context.Context) ListNestedObjectTypeOf[T] {
	return ListNestedObjectTypeOf[T]{basetypes.ListType{ElemType: NewObjectTypeOf[T](ctx)}}
}

func (t ListNestedObjectTypeOf[T]) Equal(o attr.Type) bool {
	other, ok := o.(ListNestedObjectTypeOf[T])

	if !ok {
		return false
	}

	return t.ListType.Equal(other.ListType)
}

func (t ListNestedObjectTypeOf[T]) String() string {
	var zero T
	return fmt.Sprintf("ListNestedObjectTypeOf[%T]", zero)
}

func (t ListNestedObjectTypeOf[T]) ValueFromList(ctx context.Context, in basetypes.ListValue) (basetypes.ListValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	if in.IsNull() {
		return NewListNestedObjectValueOfNull[T](ctx), diags
	}
	if in.IsUnknown() {
		return NewListNestedObjectValueOfUnknown[T](ctx), diags
	}

	listValue, d := basetypes.NewListValue(NewObjectTypeOf[T](ctx), in.Elements())
	diags.Append(d...)
	if diags.HasError() {
		return NewListNestedObjectValueOfUnknown[T](ctx), diags
	}

	value := ListNestedObjectValueOf[T]{
		ListValue: listValue,
	}

	return value, diags
}

func (t ListNestedObjectTypeOf[T]) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.ListType.ValueFromTerraform(ctx, in)

	if err != nil {
		return nil, err
	}

	listValue, ok := attrValue.(basetypes.ListValue)

	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	listValuable, diags := t.ValueFromList(ctx, listValue)

	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting ListValue to ListValuable: %v", diags)
	}

	return listValuable, nil
}

func (t ListNestedObjectTypeOf[T]) ValueType(ctx context.Context) attr.Value {
	return ListNestedObjectValueOf[T]{}
}

func (t ListNestedObjectTypeOf[T]) NewObjectPtr(ctx context.Context) (any, diag.Diagnostics) {
	return nestedObjectTypeNewObjectPtr[T](ctx)
}

func (t ListNestedObjectTypeOf[T]) NewObjectSlice(ctx context.Context, len, cap int) (any, diag.Diagnostics) {
	return nestedObjectTypeNewObjectSlice[T](ctx, len, cap)
}

func (t ListNestedObjectTypeOf[T]) NullValue(ctx context.Context) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	return NewListNestedObjectValueOfNull[T](ctx), diags
}

func (t ListNestedObjectTypeOf[T]) ValueFromObjectPtr(ctx context.Context, ptr any) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	if v, ok := ptr.(*T); ok {
		return NewListNestedObjectValueOfPtr(ctx, v), diags
	}

	diags.Append(diag.NewErrorDiagnostic("Invalid pointer value", fmt.Sprintf("incorrect type: want %T, got %T", (*T)(nil), ptr)))
	return nil, diags
}

func (t ListNestedObjectTypeOf[T]) ValueFromObjectSlice(ctx context.Context, slice any) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	if v, ok := slice.([]*T); ok {
		return NewListNestedObjectValueOfSlice(ctx, v), diags
	}

	diags.Append(diag.NewErrorDiagnostic("Invalid slice value", fmt.Sprintf("incorrect type: want %T, got %T", (*[]T)(nil), slice)))
	return nil, diags
}

func nestedObjectTypeNewObjectPtr[T any](_ context.Context) (*T, diag.Diagnostics) {
	var diags diag.Diagnostics

	return new(T), diags
}

func nestedObjectTypeNewObjectSlice[T any](_ context.Context, len, cap int) ([]*T, diag.Diagnostics) { //nolint:unparam
	var diags diag.Diagnostics

	return make([]*T, len, cap), diags
}

// ListNestedObjectValueOf represents a Terraform Plugin Framework List value whose elements are of type ObjectTypeOf.
type ListNestedObjectValueOf[T any] struct {
	basetypes.ListValue
}

var (
	_ basetypes.ListValuable = ListNestedObjectValueOf[struct{}]{}
	_ NestedObjectValue      = ListNestedObjectValueOf[struct{}]{}
)

func (v ListNestedObjectValueOf[T]) Equal(o attr.Value) bool {
	other, ok := o.(ListNestedObjectValueOf[T])

	if !ok {
		return false
	}

	return v.ListValue.Equal(other.ListValue)
}

func (v ListNestedObjectValueOf[T]) Type(ctx context.Context) attr.Type {
	return NewListNestedObjectTypeOf[T](ctx)
}

func (v ListNestedObjectValueOf[T]) ToObjectPtr(ctx context.Context) (any, diag.Diagnostics) {
	return nestedObjectValueObjectPtr[T](ctx, v.ListValue)
}

func (v ListNestedObjectValueOf[T]) ToObjectSlice(ctx context.Context) (any, diag.Diagnostics) {
	return nestedObjectValueObjectSlice[T](ctx, v.ListValue)
}

func nestedObjectValueObjectPtr[T any](ctx context.Context, val valueWithElements) (*T, diag.Diagnostics) {
	var diags diag.Diagnostics

	elements := val.Elements()
	switch n := len(elements); n {
	case 0:
		return nil, diags
	case 1:
		ptr, d := nestedObjectValueObjectPtrFromElement[T](ctx, elements[0])
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}
		return ptr, diags
	default:
		diags.Append(diag.NewErrorDiagnostic("Invalid list/set", fmt.Sprintf("too many elements: want 1, got %d", n)))
		return nil, diags
	}
}

func nestedObjectValueObjectSlice[T any](ctx context.Context, val valueWithElements) ([]*T, diag.Diagnostics) {
	var diags diag.Diagnostics

	elements := val.Elements()
	n := len(elements)
	slice := make([]*T, n)
	for i := 0; i < n; i++ {
		ptr, d := nestedObjectValueObjectPtrFromElement[T](ctx, elements[i])
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		slice[i] = ptr
	}

	return slice, diags
}

func nestedObjectValueObjectPtrFromElement[T any](ctx context.Context, val attr.Value) (*T, diag.Diagnostics) {
	var diags diag.Diagnostics

	ptr := new(T)
	diags.Append(val.(ObjectValueOf[T]).ObjectValue.As(ctx, ptr, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil, diags
	}

	return ptr, diags
}

func NewListNestedObjectValueOfNull[T any](ctx context.Context) ListNestedObjectValueOf[T] {
	return ListNestedObjectValueOf[T]{ListValue: basetypes.NewListNull(NewObjectTypeOf[T](ctx))}
}

func NewListNestedObjectValueOfUnknown[T any](ctx context.Context) ListNestedObjectValueOf[T] {
	return ListNestedObjectValueOf[T]{ListValue: basetypes.NewListUnknown(NewObjectTypeOf[T](ctx))}
}

func NewListNestedObjectValueOfPtr[T any](ctx context.Context, t *T) ListNestedObjectValueOf[T] {
	return NewListNestedObjectValueOfSlice(ctx, []*T{t})
}

func NewListNestedObjectValueOfSlice[T any](ctx context.Context, ts []*T) ListNestedObjectValueOf[T] {
	return newListNestedObjectValueOf[T](ctx, ts)
}

func NewListNestedObjectValueOfValueSlice[T any](ctx context.Context, ts []T) ListNestedObjectValueOf[T] {
	return newListNestedObjectValueOf[T](ctx, ts)
}

func newListNestedObjectValueOf[T any](ctx context.Context, elements any) ListNestedObjectValueOf[T] {
	return ListNestedObjectValueOf[T]{ListValue: fwdiag.Must(basetypes.NewListValueFrom(ctx, NewObjectTypeOf[T](ctx), elements))}
}
