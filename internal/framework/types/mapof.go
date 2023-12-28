// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"context"
	"fmt"
	"reflect"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
)

var _ basetypes.MapTypable = MapTypeOf[struct{}]{}

// MapTypeOf is the attribute type of a MapValueOf.
type MapTypeOf[T any] struct {
	basetypes.MapType
}

func NewMapTypeOf[T any](ctx context.Context) MapTypeOf[T] {
	return MapTypeOf[T]{basetypes.MapType{ElemType: ElementTypeMust[T](ctx)}}
}

func (t MapTypeOf[T]) Equal(o attr.Type) bool {
	other, ok := o.(MapTypeOf[T])

	if !ok {
		return false
	}

	return t.MapType.Equal(other.MapType)
}

func (t MapTypeOf[T]) String() string {
	return "types.MapTypeOf[" + t.ElementType().String() + "]"
}

func (t MapTypeOf[T]) ValueFromMap(ctx context.Context, in basetypes.MapValue) (basetypes.MapValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	if in.IsNull() {
		return NewMapValueOfNull[T](ctx), diags
	}
	if in.IsUnknown() {
		return NewMapValueOfUnknown[T](ctx), diags
	}

	// Here marks the spot where countless hours were spent all over the
	// internal organs of framework and autoflex only to discover the
	// first argument in this call should be an element type not the map
	// type.
	mapValue, d := basetypes.NewMapValue(t.ElementType(), in.Elements())
	diags.Append(d...)
	if diags.HasError() {
		return NewMapValueOfUnknown[T](ctx), diags
	}

	value := MapValueOf[T]{
		MapValue: mapValue,
	}

	return value, diags
}

func (t MapTypeOf[T]) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.MapType.ValueFromTerraform(ctx, in)

	if err != nil {
		return nil, err
	}

	mapValue, ok := attrValue.(basetypes.MapValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	mapValuable, diags := t.ValueFromMap(ctx, mapValue)
	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting MapValue to MapValuable: %v", diags)
	}

	return mapValuable, nil
}

func (t MapTypeOf[T]) ValueType(ctx context.Context) attr.Value {
	return MapValueOf[T]{}
}

var _ basetypes.MapValuable = MapValueOf[struct{}]{}

// MapValueOf represents a Terraform Plugin Framework Map value whose elements are of type MapTypeOf.
type MapValueOf[T any] struct {
	basetypes.MapValue
}

func (v MapValueOf[T]) Equal(o attr.Value) bool {
	other, ok := o.(MapValueOf[T])

	if !ok {
		return false
	}

	return v.MapValue.Equal(other.MapValue)
}

func (v MapValueOf[T]) Type(ctx context.Context) attr.Type {
	return NewMapTypeOf[T](ctx)
}

func NewMapValueOfNull[T any](ctx context.Context) MapValueOf[T] {
	return MapValueOf[T]{MapValue: basetypes.NewMapNull(NewMapTypeOf[T](ctx))}
}

func NewMapValueOfUnknown[T any](ctx context.Context) MapValueOf[T] {
	return MapValueOf[T]{MapValue: basetypes.NewMapUnknown(NewMapTypeOf[T](ctx))}
}

func NewMapValueOf[T any](ctx context.Context, elements map[string]T) MapValueOf[T] {
	return MapValueOf[T]{MapValue: fwdiag.Must(basetypes.NewMapValueFrom(ctx, ElementTypeMust[T](ctx), elements))}
}

func ElementType[T any](ctx context.Context) (attr.Type, error) {
	var t T
	val := reflect.ValueOf(t)
	typ := val.Type()

	v, ok := val.Interface().(attr.Value)

	if !ok {
		return nil, fmt.Errorf("%T has unsupported type: %s", t, typ)
	}

	return v.Type(ctx), nil
}

func ElementTypeMust[T any](ctx context.Context) attr.Type {
	return errs.Must(ElementType[T](ctx))
}
