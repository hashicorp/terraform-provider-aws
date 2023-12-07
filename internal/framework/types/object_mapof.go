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

// ObjectMapTypeOf is the attribute type of a ObjectMapValueOf.
type ObjectMapTypeOf[T any] struct {
	basetypes.MapType
}

var (
	_ basetypes.MapTypable = ObjectMapTypeOf[struct{}]{}
	_ ObjectMapType        = ObjectMapTypeOf[struct{}]{}
)

func NewObjectMapTypeOf[T any](ctx context.Context) ObjectMapTypeOf[T] {
	return ObjectMapTypeOf[T]{basetypes.MapType{ElemType: NewObjectTypeOf[T](ctx)}}
}

func (t ObjectMapTypeOf[T]) Equal(o attr.Type) bool {
	other, ok := o.(ObjectMapTypeOf[T])

	if !ok {
		return false
	}

	return t.MapType.Equal(other.MapType)
}

func (t ObjectMapTypeOf[T]) String() string {
	var zero T
	return fmt.Sprintf("ObjectMapTypeOf[%T]", zero)
}

func (t ObjectMapTypeOf[T]) ValueFromMap(ctx context.Context, in basetypes.MapValue) (basetypes.MapValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	if in.IsNull() {
		return NewObjectMapNullValueMapOf[T](ctx), diags
	}
	if in.IsUnknown() {
		return NewObjectMapUnknownValueMapOf[T](ctx), diags
	}

	listValue, d := basetypes.NewMapValue(NewObjectTypeOf[T](ctx), in.Elements())
	diags.Append(d...)
	if diags.HasError() {
		return NewObjectMapUnknownValueMapOf[T](ctx), diags
	}

	value := ObjectMapValueOf[T]{
		MapValue: listValue,
	}

	return value, diags
}

func (t ObjectMapTypeOf[T]) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.MapType.ValueFromTerraform(ctx, in)

	if err != nil {
		return nil, err
	}

	listValue, ok := attrValue.(basetypes.MapValue)

	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	listValuable, diags := t.ValueFromMap(ctx, listValue)

	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting MapValue to MapValuable: %v", diags)
	}

	return listValuable, nil
}

func (t ObjectMapTypeOf[T]) ValueType(ctx context.Context) attr.Value {
	return ObjectMapValueOf[T]{}
}

func (t ObjectMapTypeOf[T]) NullValue(ctx context.Context) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	return NewObjectMapNullValueMapOf[T](ctx), diags
}

// ObjectMapValueOf represents a Terraform Plugin Framework Map value whose elements are of type ObjectTypeOf.
type ObjectMapValueOf[T any] struct {
	basetypes.MapValue
}

var (
	_ basetypes.MapValuable = ObjectMapValueOf[struct{}]{}
	_ ObjectMapValue        = ObjectMapValueOf[struct{}]{}
)

func (v ObjectMapValueOf[T]) Equal(o attr.Value) bool {
	other, ok := o.(ObjectMapValueOf[T])

	if !ok {
		return false
	}

	return v.MapValue.Equal(other.MapValue)
}

func (v ObjectMapValueOf[T]) Type(ctx context.Context) attr.Type {
	return NewObjectMapTypeOf[T](ctx)
}

func (v ObjectMapValueOf[T]) ToObjectMap(ctx context.Context) (any, diag.Diagnostics) {
	return v.ToMap(ctx)
}

func (v ObjectMapValueOf[T]) ToMap(ctx context.Context) (map[string]T, diag.Diagnostics) {
	return mapNestedObjectValueObjectMap[T](ctx, v.MapValue)
}

func mapNestedObjectValueObjectMap[T any](ctx context.Context, val valueWithMapElements) (map[string]T, diag.Diagnostics) {
	var diags diag.Diagnostics

	kvs := val.Elements()
	m := make(map[string]T)
	for k := range kvs {
		ptr, d := nestedObjectValueObjectPtrFromValue[T](ctx, kvs[k])
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		m[k] = *ptr
	}

	return m, diags
}

func nestedObjectValueObjectPtrFromValue[T any](ctx context.Context, val attr.Value) (*T, diag.Diagnostics) {
	var diags diag.Diagnostics

	ptr := new(T)
	diags.Append(val.(ObjectValueOf[T]).ObjectValue.As(ctx, ptr, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil, diags
	}

	return ptr, diags
}

func NewObjectMapNullValueMapOf[T any](ctx context.Context) ObjectMapValueOf[T] {
	return ObjectMapValueOf[T]{MapValue: basetypes.NewMapNull(NewObjectTypeOf[T](ctx))}
}

func NewObjectMapUnknownValueMapOf[T any](ctx context.Context) ObjectMapValueOf[T] {
	return ObjectMapValueOf[T]{MapValue: basetypes.NewMapUnknown(NewObjectTypeOf[T](ctx))}
}

func NewObjectMapValueMapOf[T any](ctx context.Context, t map[string]T) ObjectMapValueOf[T] {
	return newObjectMapValueMapOf[T](ctx, t)
}

func NewObjectMapValuePtrMapOf[T any](ctx context.Context, t map[string]*T) ObjectMapValueOf[*T] {
	return ObjectMapValueOf[*T]{MapValue: fwdiag.Must(basetypes.NewMapValueFrom(ctx, NewObjectTypeOf[*T](ctx), t))}
}

func newObjectMapValueMapOf[T any](ctx context.Context, elements any) ObjectMapValueOf[T] {
	return ObjectMapValueOf[T]{MapValue: fwdiag.Must(basetypes.NewMapValueFrom(ctx, NewObjectTypeOf[T](ctx), elements))}
}
