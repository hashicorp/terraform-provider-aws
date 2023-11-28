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

// MappedObjectTypeOf is the attribute type of a MappedObjectValueOf.
type MappedObjectTypeOf[T any] struct {
	basetypes.MapType
}

var (
	_ basetypes.MapTypable = MappedObjectTypeOf[struct{}]{}
	_ MappedObjectType     = MappedObjectTypeOf[struct{}]{}
)

func NewMappedObjectTypeOf[T any](ctx context.Context) MappedObjectTypeOf[T] {
	return MappedObjectTypeOf[T]{basetypes.MapType{ElemType: NewObjectTypeOf[T](ctx)}}
}

func (t MappedObjectTypeOf[T]) Equal(o attr.Type) bool {
	other, ok := o.(MappedObjectTypeOf[T])

	if !ok {
		return false
	}

	return t.MapType.Equal(other.MapType)
}

func (t MappedObjectTypeOf[T]) String() string {
	var zero T
	return fmt.Sprintf("MappedObjectTypeOf[%T]", zero)
}

func (t MappedObjectTypeOf[T]) ValueFromMap(ctx context.Context, in basetypes.MapValue) (basetypes.MapValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	if in.IsNull() {
		return NewMappedObjectValueOfNull[T](ctx), diags
	}
	if in.IsUnknown() {
		return NewMappedObjectValueOfUnknown[T](ctx), diags
	}

	listValue, d := basetypes.NewMapValue(NewObjectTypeOf[T](ctx), in.Elements())
	diags.Append(d...)
	if diags.HasError() {
		return NewMappedObjectValueOfUnknown[T](ctx), diags
	}

	value := MappedObjectValueOf[T]{
		MapValue: listValue,
	}

	return value, diags
}

func (t MappedObjectTypeOf[T]) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
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

func (t MappedObjectTypeOf[T]) ValueType(ctx context.Context) attr.Value {
	return MappedObjectValueOf[T]{}
}

func (t MappedObjectTypeOf[T]) NewObjectPtr(ctx context.Context) (any, diag.Diagnostics) {
	return mapNestedObjectTypeNewObjectPtr[T](ctx)
}

func (t MappedObjectTypeOf[T]) NullValue(ctx context.Context) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	return NewMappedObjectValueOfNull[T](ctx), diags
}

func mapNestedObjectTypeNewObjectPtr[T any](_ context.Context) (*T, diag.Diagnostics) {
	var diags diag.Diagnostics

	return new(T), diags
}

func mapNestedObjectTypeNewObjectSlice[T any](_ context.Context, len, cap int) ([]*T, diag.Diagnostics) { //nolint:unparam
	var diags diag.Diagnostics

	return make([]*T, len, cap), diags
}

// MappedObjectValueOf represents a Terraform Plugin Framework Map value whose elements are of type ObjectTypeOf.
type MappedObjectValueOf[T any] struct {
	basetypes.MapValue
}

var (
	_ basetypes.MapValuable = MappedObjectValueOf[struct{}]{}
	_ MappedObjectValue     = MappedObjectValueOf[struct{}]{}
)

func (v MappedObjectValueOf[T]) Equal(o attr.Value) bool {
	other, ok := o.(MappedObjectValueOf[T])

	if !ok {
		return false
	}

	return v.MapValue.Equal(other.MapValue)
}

func (v MappedObjectValueOf[T]) Type(ctx context.Context) attr.Type {
	return NewMappedObjectTypeOf[T](ctx)
}

func (v MappedObjectValueOf[T]) ToObjectPtr(ctx context.Context) (any, diag.Diagnostics) {
	return v.ToPtrMap(ctx)
}

func (v MappedObjectValueOf[T]) ToObjectMap(ctx context.Context) (any, diag.Diagnostics) {
	return v.ToMap(ctx)
}

// ToPtr returns a pointer to the single element of a MappedObject.
func (v MappedObjectValueOf[T]) ToPtrMap(ctx context.Context) (map[string]*T, diag.Diagnostics) {
	return mapNestedObjectValueObjectPtrMap[T](ctx, v.MapValue)
}

// ToSlice returns a slice of pointers to the elements of a MappedObject.
func (v MappedObjectValueOf[T]) ToMap(ctx context.Context) (map[string]T, diag.Diagnostics) {
	return mapNestedObjectValueObjectMap[T](ctx, v.MapValue)
}

func mapNestedObjectValueObjectPtrMap[T any](ctx context.Context, val valueWithMapElements) (map[string]*T, diag.Diagnostics) {
	var diags diag.Diagnostics

	kvs := val.Elements()
	var m map[string]*T
	for k, _ := range kvs {
		ptr, d := nestedObjectValueObjectPtrFromValue[T](ctx, kvs[k])
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		m[k] = ptr
	}

	return m, diags
}

func mapNestedObjectValueObjectMap[T any](ctx context.Context, val valueWithMapElements) (map[string]T, diag.Diagnostics) {
	var diags diag.Diagnostics

	kvs := val.Elements()
	m := make(map[string]T)
	for k, _ := range kvs {
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

func NewMappedObjectValueOfNull[T any](ctx context.Context) MappedObjectValueOf[T] {
	return MappedObjectValueOf[T]{MapValue: basetypes.NewMapNull(NewObjectTypeOf[T](ctx))}
}

func NewMappedObjectValueOfUnknown[T any](ctx context.Context) MappedObjectValueOf[T] {
	return MappedObjectValueOf[T]{MapValue: basetypes.NewMapUnknown(NewObjectTypeOf[T](ctx))}
}

func NewMappedObjectValueOfMapOf[T any](ctx context.Context, t map[string]T) MappedObjectValueOf[T] {
	return newMappedObjectValueOf[T](ctx, t)
}

func NewMappedObjectValueOfMapOfPtr[T any](ctx context.Context, t map[string]*T) MappedObjectValueOf[*T] {
	return MappedObjectValueOf[*T]{MapValue: fwdiag.Must(basetypes.NewMapValueFrom(ctx, NewObjectTypeOf[*T](ctx), t))}
}

func newMappedObjectValueOf[T any](ctx context.Context, elements any) MappedObjectValueOf[T] {
	return MappedObjectValueOf[T]{MapValue: fwdiag.Must(basetypes.NewMapValueFrom(ctx, NewObjectTypeOf[T](ctx), elements))}
}
