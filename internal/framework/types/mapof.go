// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
)

var (
	_ basetypes.MapTypable  = MapTypeOf[basetypes.StringValue]{}
	_ basetypes.MapValuable = MapValueOf[basetypes.StringValue]{}
)

// MapTypeOf is the attribute type of a MapValueOf.
type MapTypeOf[T attr.Value] struct {
	basetypes.MapType
}

func NewMapTypeOf[T attr.Value](ctx context.Context) MapTypeOf[T] {
	var zero T
	return MapTypeOf[T]{basetypes.MapType{ElemType: zero.Type(ctx)}}
}

func (t MapTypeOf[T]) Equal(o attr.Type) bool {
	other, ok := o.(MapTypeOf[T])

	if !ok {
		return false
	}

	return t.MapType.Equal(other.MapType)
}

func (t MapTypeOf[T]) String() string {
	var zero T
	return fmt.Sprintf("%T", zero)
}

func (t MapTypeOf[T]) ValueFromMap(ctx context.Context, in basetypes.MapValue) (basetypes.MapValuable, diag.Diagnostics) {
	var diags diag.Diagnostics
	var zero T

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
	mapValue, d := basetypes.NewMapValue(zero.Type(ctx), in.Elements())
	diags.Append(d...)
	if diags.HasError() {
		return basetypes.NewMapUnknown(types.StringType), diags
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

// MapValueOf represents a Terraform Plugin Framework Map value whose elements are of type MapTypeOf.
type MapValueOf[T attr.Value] struct {
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

func NewMapValueOf[T attr.Value](ctx context.Context, elements map[string]attr.Value) (MapValueOf[T], diag.Diagnostics) {
	var zero T

	mapValue, diags := basetypes.NewMapValue(zero.Type(ctx), elements)
	if diags.HasError() {
		return NewMapValueOfUnknown[T](ctx), diags
	}

	return MapValueOf[T]{MapValue: mapValue}, diags
}

func NewMapValueOfNull[T attr.Value](ctx context.Context) MapValueOf[T] {
	var zero T
	return MapValueOf[T]{MapValue: basetypes.NewMapNull(zero.Type(ctx))}
}

func NewMapValueOfUnknown[T attr.Value](ctx context.Context) MapValueOf[T] {
	var zero T
	return MapValueOf[T]{MapValue: basetypes.NewMapUnknown(zero.Type(ctx))}
}

func NewMapValueOfMust[T attr.Value](ctx context.Context, elements map[string]attr.Value) MapValueOf[T] {
	return fwdiag.Must(NewMapValueOf[T](ctx, elements))
}
