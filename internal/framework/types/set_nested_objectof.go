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

var (
	_ basetypes.SetTypable                    = (*setNestedObjectTypeOf[struct{}])(nil)
	_ NestedObjectCollectionType              = (*setNestedObjectTypeOf[struct{}])(nil)
	_ basetypes.SetValuable                   = (*SetNestedObjectValueOf[struct{}])(nil)
	_ NestedObjectCollectionValue             = (*SetNestedObjectValueOf[struct{}])(nil)
	_ basetypes.SetValuableWithSemanticEquals = (*SetNestedObjectValueOf[struct{}])(nil)
)

// setNestedObjectTypeOf is the attribute type of a SetNestedObjectValueOf.
type setNestedObjectTypeOf[T any] struct {
	basetypes.SetType
	semanticEqualityFunc semanticEqualityFunc[T]
}

func NewSetNestedObjectTypeOf[T any](ctx context.Context, options ...NestedObjectOfOption[T]) setNestedObjectTypeOf[T] {
	opts := newNestedObjectOfOptions(options...)

	return setNestedObjectTypeOf[T]{
		SetType:              basetypes.SetType{ElemType: NewObjectTypeOf[T](ctx)},
		semanticEqualityFunc: opts.SemanticEqualityFunc,
	}
}

func (t setNestedObjectTypeOf[T]) Equal(o attr.Type) bool {
	other, ok := o.(setNestedObjectTypeOf[T])

	if !ok {
		return false
	}

	return t.SetType.Equal(other.SetType)
}

func (t setNestedObjectTypeOf[T]) String() string {
	var zero T
	return fmt.Sprintf("SetNestedObjectTypeOf[%T]", zero)
}

func (t setNestedObjectTypeOf[T]) ValueFromSet(ctx context.Context, in basetypes.SetValue) (basetypes.SetValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	if in.IsNull() {
		return NewSetNestedObjectValueOfNull[T](ctx), diags
	}
	if in.IsUnknown() {
		return NewSetNestedObjectValueOfUnknown[T](ctx), diags
	}

	typ, d := newObjectTypeOf[T](ctx)
	diags.Append(d...)
	if diags.HasError() {
		return NewSetNestedObjectValueOfUnknown[T](ctx), diags
	}

	v, d := basetypes.NewSetValue(typ, in.Elements())
	diags.Append(d...)
	if diags.HasError() {
		return NewSetNestedObjectValueOfUnknown[T](ctx), diags
	}

	return SetNestedObjectValueOf[T]{
		SetValue:             v,
		semanticEqualityFunc: t.semanticEqualityFunc,
	}, diags
}

func (t setNestedObjectTypeOf[T]) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
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

func (t setNestedObjectTypeOf[T]) ValueType(ctx context.Context) attr.Value {
	return SetNestedObjectValueOf[T]{semanticEqualityFunc: t.semanticEqualityFunc}
}

func (t setNestedObjectTypeOf[T]) NewObjectPtr(ctx context.Context) (any, diag.Diagnostics) {
	return objectTypeNewObjectPtr[T](ctx)
}

func (t setNestedObjectTypeOf[T]) NewObjectSlice(ctx context.Context, len, cap int) (any, diag.Diagnostics) {
	return nestedObjectTypeNewObjectSlice[T](ctx, len, cap)
}

func (t setNestedObjectTypeOf[T]) NullValue(ctx context.Context) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	return NewSetNestedObjectValueOfNull[T](ctx, WithSemanticEqualityFunc(t.semanticEqualityFunc)), diags
}

func (t setNestedObjectTypeOf[T]) ValueFromObjectPtr(ctx context.Context, ptr any) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	if v, ok := ptr.(*T); ok {
		v, d := NewSetNestedObjectValueOfPtr(ctx, v)
		diags.Append(d...)
		return v, d
	}

	diags.Append(diag.NewErrorDiagnostic("Invalid pointer value", fmt.Sprintf("incorrect type: want %T, got %T", (*T)(nil), ptr)))
	return nil, diags
}

func (t setNestedObjectTypeOf[T]) ValueFromObjectSlice(ctx context.Context, slice any) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	if v, ok := slice.([]*T); ok {
		v, d := NewSetNestedObjectValueOfSlice(ctx, v, t.semanticEqualityFunc)
		diags.Append(d...)
		return v, d
	}

	diags.Append(diag.NewErrorDiagnostic("Invalid slice value", fmt.Sprintf("incorrect type: want %T, got %T", (*[]T)(nil), slice)))
	return nil, diags
}

// SetNestedObjectValueOf represents a Terraform Plugin Framework Set value whose elements are of type `ObjectTypeOf[T]`.
type SetNestedObjectValueOf[T any] struct {
	basetypes.SetValue
	semanticEqualityFunc semanticEqualityFunc[T]
}

func (v SetNestedObjectValueOf[T]) Equal(o attr.Value) bool {
	other, ok := o.(SetNestedObjectValueOf[T])

	if !ok {
		return false
	}

	return v.SetValue.Equal(other.SetValue)
}

func (v SetNestedObjectValueOf[T]) SetSemanticEquals(ctx context.Context, newValuable basetypes.SetValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	// returning false will fall back to framework defined semantic equality checks
	if v.semanticEqualityFunc == nil {
		return false, diags
	}

	newValue, ok := newValuable.(SetNestedObjectValueOf[T])
	if !ok {
		diags.AddError(
			"SetSemanticEquals",
			fmt.Sprintf("unexpected value type of %T", newValuable),
		)
		return false, diags
	}

	return v.semanticEqualityFunc(ctx, v, newValue)
}

func (v SetNestedObjectValueOf[T]) Type(ctx context.Context) attr.Type {
	return NewSetNestedObjectTypeOf[T](ctx)
}

func (v SetNestedObjectValueOf[T]) ToObjectPtr(ctx context.Context) (any, diag.Diagnostics) {
	return v.ToPtr(ctx)
}

func (v SetNestedObjectValueOf[T]) ToObjectSlice(ctx context.Context) (any, diag.Diagnostics) {
	return v.ToSlice(ctx)
}

// ToPtr returns a pointer to the single element of a SetNestedObject.
func (v SetNestedObjectValueOf[T]) ToPtr(ctx context.Context) (*T, diag.Diagnostics) {
	return nestedObjectValueObjectPtr[T](ctx, v.SetValue)
}

// ToSlice returns a slice of pointers to the elements of a SetNestedObject.
func (v SetNestedObjectValueOf[T]) ToSlice(ctx context.Context) ([]*T, diag.Diagnostics) {
	return nestedObjectValueObjectSlice[T](ctx, v.SetValue)
}

func NewSetNestedObjectValueOfNull[T any](ctx context.Context, options ...NestedObjectOfOption[T]) SetNestedObjectValueOf[T] {
	opts := newNestedObjectOfOptions(options...)
	return SetNestedObjectValueOf[T]{SetValue: basetypes.NewSetNull(NewObjectTypeOf[T](ctx)), semanticEqualityFunc: opts.SemanticEqualityFunc}
}

func NewSetNestedObjectValueOfUnknown[T any](ctx context.Context) SetNestedObjectValueOf[T] {
	return SetNestedObjectValueOf[T]{SetValue: basetypes.NewSetUnknown(NewObjectTypeOf[T](ctx))}
}

func NewSetNestedObjectValueOfPtr[T any](ctx context.Context, t *T, options ...NestedObjectOfOption[T]) (SetNestedObjectValueOf[T], diag.Diagnostics) {
	opts := newNestedObjectOfOptions(options...)
	return NewSetNestedObjectValueOfSlice(ctx, []*T{t}, opts.SemanticEqualityFunc)
}

func NewSetNestedObjectValueOfPtrMust[T any](ctx context.Context, t *T, options ...NestedObjectOfOption[T]) SetNestedObjectValueOf[T] {
	return fwdiag.Must(NewSetNestedObjectValueOfPtr(ctx, t, options...))
}

func NewSetNestedObjectValueOfSlice[T any](ctx context.Context, ts []*T, f semanticEqualityFunc[T]) (SetNestedObjectValueOf[T], diag.Diagnostics) {
	return newSetNestedObjectValueOf[T](ctx, ts, f)
}

func NewSetNestedObjectValueOfSliceMust[T any](ctx context.Context, ts []*T, options ...NestedObjectOfOption[T]) SetNestedObjectValueOf[T] {
	opts := newNestedObjectOfOptions(options...)
	return fwdiag.Must(NewSetNestedObjectValueOfSlice(ctx, ts, opts.SemanticEqualityFunc))
}

func NewSetNestedObjectValueOfValueSlice[T any](ctx context.Context, ts []T, options ...NestedObjectOfOption[T]) (SetNestedObjectValueOf[T], diag.Diagnostics) {
	opts := newNestedObjectOfOptions(options...)
	return newSetNestedObjectValueOf[T](ctx, ts, opts.SemanticEqualityFunc)
}

func NewSetNestedObjectValueOfValueSliceMust[T any](ctx context.Context, ts []T, options ...NestedObjectOfOption[T]) SetNestedObjectValueOf[T] {
	return fwdiag.Must(NewSetNestedObjectValueOfValueSlice(ctx, ts, options...))
}

func newSetNestedObjectValueOf[T any](ctx context.Context, elements any, f semanticEqualityFunc[T]) (SetNestedObjectValueOf[T], diag.Diagnostics) {
	var diags diag.Diagnostics

	typ, d := newObjectTypeOf[T](ctx)
	diags.Append(d...)
	if diags.HasError() {
		return NewSetNestedObjectValueOfUnknown[T](ctx), diags
	}

	v, d := basetypes.NewSetValueFrom(ctx, typ, elements)
	diags.Append(d...)
	if diags.HasError() {
		return NewSetNestedObjectValueOfUnknown[T](ctx), diags
	}

	return SetNestedObjectValueOf[T]{SetValue: v, semanticEqualityFunc: f}, diags
}
