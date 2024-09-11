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
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
)

var (
	_ basetypes.ObjectTypable  = (*objectTypeOf[struct{}])(nil)
	_ NestedObjectType         = (*objectTypeOf[struct{}])(nil)
	_ basetypes.ObjectValuable = (*ObjectValueOf[struct{}])(nil)
	_ NestedObjectValue        = (*ObjectValueOf[struct{}])(nil)
)

// objectTypeOf is the attribute type of an ObjectValueOf.
type objectTypeOf[T any] struct {
	basetypes.ObjectType
}

func newObjectTypeOf[T any](ctx context.Context) (objectTypeOf[T], diag.Diagnostics) {
	var diags diag.Diagnostics

	m, d := AttributeTypes[T](ctx)
	diags.Append(d...)
	if diags.HasError() {
		return objectTypeOf[T]{}, diags
	}

	return objectTypeOf[T]{basetypes.ObjectType{AttrTypes: m}}, diags
}

func NewObjectTypeOf[T any](ctx context.Context) objectTypeOf[T] {
	return fwdiag.Must(newObjectTypeOf[T](ctx))
}

func (t objectTypeOf[T]) Equal(o attr.Type) bool {
	other, ok := o.(objectTypeOf[T])
	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t objectTypeOf[T]) String() string {
	var zero T
	return fmt.Sprintf("ObjectTypeOf[%T]", zero)
}

func (t objectTypeOf[T]) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	if in.IsNull() {
		return NewObjectValueOfNull[T](ctx), diags
	}
	if in.IsUnknown() {
		return NewObjectValueOfUnknown[T](ctx), diags
	}

	m, d := AttributeTypes[T](ctx)
	diags.Append(d...)
	if diags.HasError() {
		return NewObjectValueOfUnknown[T](ctx), diags
	}

	v, d := basetypes.NewObjectValue(m, in.Attributes())
	diags.Append(d...)
	if diags.HasError() {
		return NewObjectValueOfUnknown[T](ctx), diags
	}

	value := ObjectValueOf[T]{
		ObjectValue: v,
	}

	return value, diags
}

func (t objectTypeOf[T]) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
	attrValue, err := t.ObjectType.ValueFromTerraform(ctx, in)

	if err != nil {
		return nil, err
	}

	objectValue, ok := attrValue.(basetypes.ObjectValue)
	if !ok {
		return nil, fmt.Errorf("unexpected value type of %T", attrValue)
	}

	objectValuable, diags := t.ValueFromObject(ctx, objectValue)
	if diags.HasError() {
		return nil, fmt.Errorf("unexpected error converting ObjectValue to ObjectValuable: %v", diags)
	}

	return objectValuable, nil
}

func (t objectTypeOf[T]) ValueType(ctx context.Context) attr.Value {
	return ObjectValueOf[T]{}
}

func (t objectTypeOf[T]) NewObjectPtr(ctx context.Context) (any, diag.Diagnostics) {
	return objectTypeNewObjectPtr[T](ctx)
}

func (t objectTypeOf[T]) NullValue(ctx context.Context) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics
	return NewObjectValueOfNull[T](ctx), diags
}

func (t objectTypeOf[T]) ValueFromObjectPtr(ctx context.Context, ptr any) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	if v, ok := ptr.(*T); ok {
		v, d := NewObjectValueOf(ctx, v)
		diags.Append(d...)
		return v, diags
	}

	diags.Append(diag.NewErrorDiagnostic("Invalid pointer value", fmt.Sprintf("incorrect type: want %T, got %T", (*T)(nil), ptr)))
	return nil, diags
}

func objectTypeNewObjectPtr[T any](context.Context) (*T, diag.Diagnostics) {
	var diags diag.Diagnostics
	return new(T), diags
}

// NullOutObjectPtrFields sets all applicable fields of the specified object pointer to their null values.
func NullOutObjectPtrFields[T any](ctx context.Context, t *T) diag.Diagnostics {
	var diags diag.Diagnostics
	val := reflect.ValueOf(t)
	typ := val.Type().Elem()

	if typ.Kind() != reflect.Struct {
		return diags
	}

	val = val.Elem()

	for i := 0; i < typ.NumField(); i++ {
		val := val.Field(i)
		if !val.CanInterface() {
			continue
		}

		attrValue, err := NullValueOf(ctx, val.Interface())

		if err != nil {
			diags.Append(diag.NewErrorDiagnostic("attr.Type.ValueFromTerraform", err.Error()))
			return diags
		}

		if attrValue == nil {
			continue
		}

		val.Set(reflect.ValueOf(attrValue))
	}

	return diags
}

// ObjectValueOf represents a Terraform Plugin Framework Object value whose corresponding Go type is the structure T.
type ObjectValueOf[T any] struct {
	basetypes.ObjectValue
}

func (v ObjectValueOf[T]) Equal(o attr.Value) bool {
	other, ok := o.(ObjectValueOf[T])
	if !ok {
		return false
	}

	return v.ObjectValue.Equal(other.ObjectValue)
}

func (v ObjectValueOf[T]) Type(ctx context.Context) attr.Type {
	return NewObjectTypeOf[T](ctx)
}

func (v ObjectValueOf[T]) ToObjectPtr(ctx context.Context) (any, diag.Diagnostics) {
	return v.ToPtr(ctx)
}

func (v ObjectValueOf[T]) ToPtr(ctx context.Context) (*T, diag.Diagnostics) {
	return objectValueObjectPtr[T](ctx, v)
}

func objectValueObjectPtr[T any](ctx context.Context, val attr.Value) (*T, diag.Diagnostics) {
	var diags diag.Diagnostics

	ptr, d := objectTypeNewObjectPtr[T](ctx)
	diags.Append(d...)
	if diags.HasError() {
		return nil, diags
	}

	diags.Append(val.(ObjectValueOf[T]).ObjectValue.As(ctx, ptr, basetypes.ObjectAsOptions{})...)
	if diags.HasError() {
		return nil, diags
	}

	return ptr, diags
}

func NewObjectValueOfNull[T any](ctx context.Context) ObjectValueOf[T] {
	return ObjectValueOf[T]{ObjectValue: basetypes.NewObjectNull(AttributeTypesMust[T](ctx))}
}

func NewObjectValueOfUnknown[T any](ctx context.Context) ObjectValueOf[T] {
	return ObjectValueOf[T]{ObjectValue: basetypes.NewObjectUnknown(AttributeTypesMust[T](ctx))}
}

func NewObjectValueOf[T any](ctx context.Context, t *T) (ObjectValueOf[T], diag.Diagnostics) {
	var diags diag.Diagnostics

	m, d := AttributeTypes[T](ctx)
	diags.Append(d...)
	if diags.HasError() {
		return NewObjectValueOfUnknown[T](ctx), diags
	}

	v, d := basetypes.NewObjectValueFrom(ctx, m, t)
	diags.Append(d...)
	if diags.HasError() {
		return NewObjectValueOfUnknown[T](ctx), diags
	}

	return ObjectValueOf[T]{ObjectValue: v}, diags
}

func NewObjectValueOfMust[T any](ctx context.Context, t *T) ObjectValueOf[T] {
	return fwdiag.Must(NewObjectValueOf[T](ctx, t))
}
