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

type ListTypeOf[T any] struct {
	basetypes.ListType
}

var _ basetypes.ListTypable = ListTypeOf[struct{}]{}

func NewListTypeOf[T any](ctx context.Context) ListTypeOf[T] {
	return ListTypeOf[T]{basetypes.ListType{ElemType: NewObjectTypeOf[T](ctx)}}
}

func (t ListTypeOf[T]) Equal(o attr.Type) bool {
	other, ok := o.(ListTypeOf[T])

	if !ok {
		return false
	}

	return t.ListType.Equal(other.ListType)
}

func (t ListTypeOf[T]) String() string {
	var zero T
	return fmt.Sprintf("ListTypeOf[%T]", zero)
}

func (t ListTypeOf[T]) ValueFromList(ctx context.Context, in basetypes.ListValue) (basetypes.ListValuable, diag.Diagnostics) {
	if in.IsNull() {
		return NewListValueOfNull[T](ctx), nil
	}
	if in.IsUnknown() {
		return NewListValueOfUnknown[T](ctx), nil
	}

	listValue, diags := basetypes.NewListValue(NewObjectTypeOf[T](ctx), in.Elements())

	if diags.HasError() {
		return NewListValueOfUnknown[T](ctx), diags
	}

	value := ListValueOf[T]{
		ListValue: listValue,
	}

	return value, nil
}

func (t ListTypeOf[T]) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
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

func (t ListTypeOf[T]) ValueType(ctx context.Context) attr.Value {
	return ListValueOf[T]{}
}

type ListValueOf[T any] struct {
	basetypes.ListValue
}

func (v ListValueOf[T]) Equal(o attr.Value) bool {
	other, ok := o.(ListValueOf[T])

	if !ok {
		return false
	}

	return v.ListValue.Equal(other.ListValue)
}

func (v ListValueOf[T]) Type(ctx context.Context) attr.Type {
	return NewListTypeOf[T](ctx)
}

func NewListValueOfNull[T any](ctx context.Context) ListValueOf[T] {
	return ListValueOf[T]{ListValue: basetypes.NewListNull(NewObjectTypeOf[T](ctx))}
}

func NewListValueOfUnknown[T any](ctx context.Context) ListValueOf[T] {
	return ListValueOf[T]{ListValue: basetypes.NewListUnknown(NewObjectTypeOf[T](ctx))}
}

func NewListValueOfPtr[T any](ctx context.Context, t *T) ListValueOf[T] {
	return NewListValueOfPtrSlice(ctx, []*T{t})
}

func NewListValueOfPtrSlice[T any](ctx context.Context, ts []*T) ListValueOf[T] {
	return newListValueOf[T](ctx, ts)
}

func NewListValueOfValueSlice[T any](ctx context.Context, ts []T) ListValueOf[T] {
	return newListValueOf[T](ctx, ts)
}

func newListValueOf[T any](ctx context.Context, elements any) ListValueOf[T] {
	return ListValueOf[T]{ListValue: fwdiag.Must(basetypes.NewListValueFrom(ctx, NewObjectTypeOf[T](ctx), elements))}
}
