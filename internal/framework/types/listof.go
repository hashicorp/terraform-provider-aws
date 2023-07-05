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

func (t ListTypeOf[T]) ValueFromList(context.Context, basetypes.ListValue) (basetypes.ListValuable, diag.Diagnostics) {
	return nil, nil
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
	return ListValueOfT[T]{}
}

type ListValueOfT[T any] struct {
	basetypes.ListValue
}
