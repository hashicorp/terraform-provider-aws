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

// ObjectTypeOf is the attribute type of an ObjectValueOf.
type ObjectTypeOf[T any] struct {
	basetypes.ObjectType
}

var _ basetypes.ObjectTypable = ObjectTypeOf[struct{}]{}

func NewObjectTypeOf[T any](ctx context.Context) ObjectTypeOf[T] {
	return ObjectTypeOf[T]{basetypes.ObjectType{AttrTypes: AttributeTypesMust[T](ctx)}}
}

func (t ObjectTypeOf[T]) Equal(o attr.Type) bool {
	other, ok := o.(ObjectTypeOf[T])

	if !ok {
		return false
	}

	return t.ObjectType.Equal(other.ObjectType)
}

func (t ObjectTypeOf[T]) String() string {
	var zero T
	return fmt.Sprintf("ObjectTypeOf[%T]", zero)
}

func (t ObjectTypeOf[T]) ValueFromObject(ctx context.Context, in basetypes.ObjectValue) (basetypes.ObjectValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	if in.IsNull() {
		return NewObjectValueOfNull[T](ctx), diags
	}
	if in.IsUnknown() {
		return NewObjectValueOfUnknown[T](ctx), diags
	}

	objectValue, d := basetypes.NewObjectValue(AttributeTypesMust[T](ctx), in.Attributes())
	diags.Append(d...)
	if diags.HasError() {
		return NewObjectValueOfUnknown[T](ctx), diags
	}

	value := ObjectValueOf[T]{
		ObjectValue: objectValue,
	}

	return value, diags
}

func (t ObjectTypeOf[T]) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
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

func (t ObjectTypeOf[T]) ValueType(ctx context.Context) attr.Value {
	return ObjectValueOf[T]{}
}

// ObjectValueOf represents a Terraform Plugin Framework Object value whose corresponding Go type is the structure T.
type ObjectValueOf[T any] struct {
	basetypes.ObjectValue
}

var _ basetypes.ObjectValuable = ObjectValueOf[struct{}]{}

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

func NewObjectValueOfNull[T any](ctx context.Context) ObjectValueOf[T] {
	return ObjectValueOf[T]{ObjectValue: basetypes.NewObjectNull(AttributeTypesMust[T](ctx))}
}

func NewObjectValueOfUnknown[T any](ctx context.Context) ObjectValueOf[T] {
	return ObjectValueOf[T]{ObjectValue: basetypes.NewObjectUnknown(AttributeTypesMust[T](ctx))}
}

func NewObjectValueOf[T any](ctx context.Context, t *T) ObjectValueOf[T] {
	return ObjectValueOf[T]{ObjectValue: fwdiag.Must(basetypes.NewObjectValueFrom(ctx, AttributeTypesMust[T](ctx), t))}
}
