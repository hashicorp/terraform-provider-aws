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
	_ basetypes.SetTypable        = (*setNestedObjectTypeOf[struct{}])(nil)
	_ NestedObjectCollectionType  = (*setNestedObjectTypeOf[struct{}])(nil)
	_ basetypes.SetValuable       = (*SetNestedObjectValueOf[struct{}])(nil)
	_ NestedObjectCollectionValue = (*SetNestedObjectValueOf[struct{}])(nil)
)

type SemanticEqualityFunc[T any] func(context.Context, basetypes.SetValue, basetypes.SetValue) (bool, diag.Diagnostics)

// setNestedObjectTypeOf is the attribute type of a SetNestedObjectValueOf.
type setNestedObjectTypeOf[T any] struct {
	basetypes.SetType
	semanticEqualityFunc SemanticEqualityFunc[T]
}

func NewSetNestedObjectTypeOf[T any](ctx context.Context, f ...SemanticEqualityFunc[T]) setNestedObjectTypeOf[T] {
	var sf SemanticEqualityFunc[T]
	if len(f) == 1 {
		sf = f[0]
	}
	return setNestedObjectTypeOf[T]{
		SetType:              basetypes.SetType{ElemType: NewObjectTypeOf[T](ctx)},
		semanticEqualityFunc: sf,
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
	return SetNestedObjectValueOf[T]{}
}

func (t setNestedObjectTypeOf[T]) NewObjectPtr(ctx context.Context) (any, diag.Diagnostics) {
	return objectTypeNewObjectPtr[T](ctx)
}

func (t setNestedObjectTypeOf[T]) NewObjectSlice(ctx context.Context, len, cap int) (any, diag.Diagnostics) {
	return nestedObjectTypeNewObjectSlice[T](ctx, len, cap)
}

func (t setNestedObjectTypeOf[T]) NullValue(ctx context.Context) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	return NewSetNestedObjectValueOfNull[T](ctx), diags
}

func (t setNestedObjectTypeOf[T]) ValueFromObjectPtr(ctx context.Context, ptr any) (attr.Value, diag.Diagnostics) {
	var diags diag.Diagnostics

	if v, ok := ptr.(*T); ok {
		v, d := NewSetNestedObjectValueOfPtr(ctx, v, t.semanticEqualityFunc)
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

var (
	_ basetypes.SetValuableWithSemanticEquals = (*SetNestedObjectValueOf[struct{}])(nil)
)

// SetNestedObjectValueOf represents a Terraform Plugin Framework Set value whose elements are of type `ObjectTypeOf[T]`.
type SetNestedObjectValueOf[T any] struct {
	basetypes.SetValue
	semanticEqualityFunc SemanticEqualityFunc[T]
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
	if v.semanticEqualityFunc != nil {
		return true, diags
	}

	//newValue, ok := newValuable.(SetNestedObjectValueOf[T])
	//if !ok {
	//	diags.AddError(
	//		"SetSemanticEquals",
	//		fmt.Sprintf("unexpected value type of %T", newValuable),
	//	)
	//	return false, diags
	//}
	//
	//vs, d := v.ToSetValue(ctx)
	//diags.Append(d...)
	//ns, d := newValue.SetValue.ToSetValue(ctx)
	//diags.Append(d...)
	//if diags.HasError() {
	//	return false, diags
	//}

	return true, diags
}

func (v SetNestedObjectValueOf[T]) Type(ctx context.Context) attr.Type {
	return NewSetNestedObjectTypeOf[T](ctx, v.semanticEqualityFunc)
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
	return nestedObjectValueObjectSet[T](ctx, v.SetValue, v.semanticEqualityFunc)
}

func NewSetNestedObjectValueOfNull[T any](ctx context.Context) SetNestedObjectValueOf[T] {
	return SetNestedObjectValueOf[T]{SetValue: basetypes.NewSetNull(NewObjectTypeOf[T](ctx))}
}

func NewSetNestedObjectValueOfUnknown[T any](ctx context.Context) SetNestedObjectValueOf[T] {
	return SetNestedObjectValueOf[T]{SetValue: basetypes.NewSetUnknown(NewObjectTypeOf[T](ctx))}
}

func NewSetNestedObjectValueOfPtr[T any](ctx context.Context, t *T, f SemanticEqualityFunc[T]) (SetNestedObjectValueOf[T], diag.Diagnostics) {
	return NewSetNestedObjectValueOfSlice(ctx, []*T{t}, f)
}

func NewSetNestedObjectValueOfPtrMust[T any](ctx context.Context, t *T, f SemanticEqualityFunc[T]) SetNestedObjectValueOf[T] {
	return fwdiag.Must(NewSetNestedObjectValueOfPtr(ctx, t, f))
}

func NewSetNestedObjectValueOfSlice[T any](ctx context.Context, ts []*T, f SemanticEqualityFunc[T]) (SetNestedObjectValueOf[T], diag.Diagnostics) {
	return newSetNestedObjectValueOf[T](ctx, ts, f)
}

func NewSetNestedObjectValueOfSliceMust[T any](ctx context.Context, ts []*T, f SemanticEqualityFunc[T]) SetNestedObjectValueOf[T] {
	return fwdiag.Must(NewSetNestedObjectValueOfSlice(ctx, ts, f))
}

func NewSetNestedObjectValueOfValueSlice[T any](ctx context.Context, ts []T, f SemanticEqualityFunc[T]) (SetNestedObjectValueOf[T], diag.Diagnostics) {
	return newSetNestedObjectValueOf[T](ctx, ts, f)
}

func NewSetNestedObjectValueOfValueSliceMust[T any](ctx context.Context, ts []T, f SemanticEqualityFunc[T]) SetNestedObjectValueOf[T] {
	return fwdiag.Must(NewSetNestedObjectValueOfValueSlice(ctx, ts, f))
}

func newSetNestedObjectValueOf[T any](ctx context.Context, elements any, f SemanticEqualityFunc[T]) (SetNestedObjectValueOf[T], diag.Diagnostics) {
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

func nestedObjectValueObjectSet[T any](ctx context.Context, val valueWithElements, f SemanticEqualityFunc[T]) ([]*T, diag.Diagnostics) {
	var diags diag.Diagnostics

	elements := val.Elements()
	n := len(elements)
	slice := make([]*T, n)
	for i := 0; i < n; i++ {
		ptr, d := objectValueObjectPtr[T](ctx, elements[i])
		diags.Append(d...)
		if diags.HasError() {
			return nil, diags
		}

		slice[i] = ptr
	}

	return slice, diags
}
