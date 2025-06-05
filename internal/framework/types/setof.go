// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/attr/xattr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
)

var (
	_ basetypes.SetTypable        = (*setTypeOf[basetypes.StringValue])(nil)
	_ basetypes.SetValuable       = (*SetValueOf[basetypes.StringValue])(nil)
	_ xattr.ValidateableAttribute = (*SetValueOf[basetypes.StringValue])(nil)
)

// setTypeOf is the attribute type of a SetValueOf.
type setTypeOf[T attr.Value] struct {
	basetypes.SetType
	validateAttributeFunc validateAttributeFunc
}

var (
	// SetOfStringType is a custom type used for defining a Set of strings.
	SetOfStringType = setTypeOf[basetypes.StringValue]{basetypes.SetType{ElemType: basetypes.StringType{}}, nil}

	// SetOfARNType is a custom type used for defining a Set of ARNs.
	SetOfARNType = setTypeOf[ARN]{basetypes.SetType{ElemType: ARNType}, nil}
)

// TODO Replace with Go 1.24 generic type alias when available.
func SetOfStringEnumType[T enum.Valueser[T]]() setTypeOf[StringEnum[T]] {
	validateFunc := func(ctx context.Context, path path.Path, values []attr.Value) diag.Diagnostics {
		var diags diag.Diagnostics
		for index, enumVal := range values {
			val, ok := enumVal.(StringEnum[T])
			if !ok {
				diags.AddAttributeError(
					path,
					"Invalid String Enum Type",
					fmt.Sprintf("Expected type: %v, got: %v", StringEnum[T]{}.Type(ctx), enumVal.Type(ctx)),
				)

				return diags
			}

			if val.IsNull() || val.IsUnknown() {
				continue
			}

			if !slices.Contains(val.ValueEnum().Values(), val.ValueEnum()) {
				parentPath := fmt.Sprintf("%v[%d]", path, index)
				diags.AddAttributeError(
					path,
					"Invalid String Enum Value",
					fmt.Sprintf("Value [%s] at attribute %v is not a valid enum value. Valid values are: %s",
						val.ValueString(), parentPath, val.ValueEnum().Values()),
				)
			}
		}
		return diags
	}

	return setTypeOf[StringEnum[T]]{basetypes.SetType{ElemType: StringEnumType[T]()}, validateFunc}
}

func NewSetTypeOf[T attr.Value](ctx context.Context) setTypeOf[T] {
	return setTypeOf[T]{basetypes.SetType{ElemType: newAttrTypeOf[T](ctx)}, nil}
}

func (t setTypeOf[T]) Equal(o attr.Type) bool {
	other, ok := o.(setTypeOf[T])

	if !ok {
		return false
	}

	return t.SetType.Equal(other.SetType)
}

func (t setTypeOf[T]) String() string {
	var zero T
	return fmt.Sprintf("SetTypeOf[%T]", zero)
}

func (t setTypeOf[T]) ValueFromSet(ctx context.Context, in basetypes.SetValue) (basetypes.SetValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	if in.IsNull() {
		return NewSetValueOfNull[T](ctx), diags
	}
	if in.IsUnknown() {
		return NewSetValueOfUnknown[T](ctx), diags
	}

	v, d := basetypes.NewSetValue(newAttrTypeOf[T](ctx), in.Elements())
	diags.Append(d...)
	if diags.HasError() {
		return NewSetValueOfUnknown[T](ctx), diags
	}

	return SetValueOf[T]{SetValue: v, validateAttributeFunc: t.validateAttributeFunc}, diags
}

func (t setTypeOf[T]) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
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

func (t setTypeOf[T]) ValueType(ctx context.Context) attr.Value {
	return SetValueOf[T]{}
}

// SetValueOf represents a Terraform Plugin Framework Set value whose elements are of type `T`.
type SetValueOf[T attr.Value] struct {
	basetypes.SetValue
	validateAttributeFunc validateAttributeFunc
}

type (
	SetOfString = SetValueOf[basetypes.StringValue]
	SetOfARN    = SetValueOf[ARN]
)

func (v SetValueOf[T]) Equal(o attr.Value) bool {
	other, ok := o.(SetValueOf[T])

	if !ok {
		return false
	}

	return v.SetValue.Equal(other.SetValue)
}

func (v SetValueOf[T]) Type(ctx context.Context) attr.Type {
	return NewSetTypeOf[T](ctx)
}

func (v SetValueOf[T]) ValidateAttribute(ctx context.Context, req xattr.ValidateAttributeRequest, resp *xattr.ValidateAttributeResponse) {
	if v.IsNull() || v.IsUnknown() || v.validateAttributeFunc == nil {
		return
	}

	resp.Diagnostics.Append(v.validateAttributeFunc(ctx, req.Path, v.Elements())...)
}

func NewSetValueOfNull[T attr.Value](ctx context.Context) SetValueOf[T] {
	return SetValueOf[T]{SetValue: basetypes.NewSetNull(newAttrTypeOf[T](ctx))}
}

func NewSetValueOfUnknown[T attr.Value](ctx context.Context) SetValueOf[T] {
	return SetValueOf[T]{SetValue: basetypes.NewSetUnknown(newAttrTypeOf[T](ctx))}
}

func NewSetValueOf[T attr.Value](ctx context.Context, elements []attr.Value) (SetValueOf[T], diag.Diagnostics) {
	var diags diag.Diagnostics

	v, d := basetypes.NewSetValue(newAttrTypeOf[T](ctx), elements)
	diags.Append(d...)
	if diags.HasError() {
		return NewSetValueOfUnknown[T](ctx), diags
	}

	return SetValueOf[T]{SetValue: v}, diags
}

func NewSetValueOfMust[T attr.Value](ctx context.Context, elements []attr.Value) SetValueOf[T] {
	return fwdiag.Must(NewSetValueOf[T](ctx, elements))
}
