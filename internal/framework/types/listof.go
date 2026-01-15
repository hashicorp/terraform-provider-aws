// Copyright IBM Corp. 2014, 2026
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
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	inttypes "github.com/hashicorp/terraform-provider-aws/internal/types"
)

var (
	_ basetypes.ListTypable       = (*listTypeOf[basetypes.StringValue])(nil)
	_ basetypes.ListValuable      = (*ListValueOf[basetypes.StringValue])(nil)
	_ xattr.ValidateableAttribute = (*ListValueOf[basetypes.StringValue])(nil)
)

var (
	// ListOfARNType is a custom type used for defining a List of ARNs.
	ListOfARNType = listTypeOf[ARN]{basetypes.ListType{ElemType: ARNType}, nil}

	// ListOfInt64Type is a custom type used for defining a List of int64s.
	ListOfInt64Type = listTypeOf[basetypes.Int64Value]{basetypes.ListType{ElemType: basetypes.Int64Type{}}, nil}

	// ListOfStringType is a custom type used for defining a List of strings.
	ListOfStringType = listTypeOf[basetypes.StringValue]{basetypes.ListType{ElemType: basetypes.StringType{}}, nil}
)

type validateAttributeFunc[T attr.Value] func(context.Context, path.Path, []attr.Value) diag.Diagnostics

// ListOfStringEnumType is a custom type used for defining a List of string enums.
func ListOfStringEnumType[T enum.Valueser[T]]() listTypeOf[StringEnum[T]] {
	return listTypeOf[StringEnum[T]]{basetypes.ListType{ElemType: StringEnumType[T]()}, validateStringEnumSlice[T]}
}

type listTypeOf[T attr.Value] struct {
	basetypes.ListType
	validateAttributeFunc validateAttributeFunc[T]
}

func newListTypeOf[T attr.Value](ctx context.Context) listTypeOf[T] {
	return listTypeOf[T]{basetypes.ListType{ElemType: newAttrTypeOf[T](ctx)}, nil}
}

func (t listTypeOf[T]) Equal(o attr.Type) bool {
	other, ok := o.(listTypeOf[T])

	if !ok {
		return false
	}

	return t.ListType.Equal(other.ListType)
}

func (t listTypeOf[T]) String() string {
	return fmt.Sprintf("ListTypeOf[%T]", inttypes.Zero[T]())
}

func (t listTypeOf[T]) ValueFromList(ctx context.Context, in basetypes.ListValue) (basetypes.ListValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	if in.IsNull() {
		return NewListValueOfNull[T](ctx), diags
	}

	if in.IsUnknown() {
		return NewListValueOfUnknown[T](ctx), diags
	}

	v, d := basetypes.NewListValue(newAttrTypeOf[T](ctx), in.Elements())
	diags.Append(d...)
	if diags.HasError() {
		return NewListValueOfUnknown[T](ctx), diags
	}

	return ListValueOf[T]{ListValue: v, validateAttributeFunc: t.validateAttributeFunc}, diags
}

func (t listTypeOf[T]) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
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

func (t listTypeOf[T]) ValueType(ctx context.Context) attr.Value {
	return ListValueOf[T]{}
}

type ListValueOf[T attr.Value] struct {
	basetypes.ListValue
	validateAttributeFunc validateAttributeFunc[T]
}

type (
	ListOfARN                            = ListValueOf[ARN]
	ListOfInt64                          = ListValueOf[basetypes.Int64Value]
	ListOfString                         = ListValueOf[basetypes.StringValue]
	ListOfStringEnum[T enum.Valueser[T]] = ListValueOf[StringEnum[T]]
)

func (v ListValueOf[T]) Equal(o attr.Value) bool {
	other, ok := o.(ListValueOf[T])

	if !ok {
		return false
	}

	return v.ListValue.Equal(other.ListValue)
}

func (v ListValueOf[T]) Type(ctx context.Context) attr.Type {
	return newListTypeOf[T](ctx)
}

func (v ListValueOf[T]) ValidateAttribute(ctx context.Context, req xattr.ValidateAttributeRequest, resp *xattr.ValidateAttributeResponse) {
	if v.IsNull() || v.IsUnknown() || v.validateAttributeFunc == nil {
		return
	}

	resp.Diagnostics.Append(v.validateAttributeFunc(ctx, req.Path, v.Elements())...)
}

// IsFullyKnown returns true if `v` all its elements are known.
func (v ListValueOf[T]) IsFullyKnown() bool {
	switch {
	case v.IsUnknown():
		return false
	case v.IsNull():
		return true
	default:
		return tfslices.All(v.Elements(), func(v attr.Value) bool {
			return !v.IsUnknown()
		})
	}
}

func NewListValueOfNull[T attr.Value](ctx context.Context) ListValueOf[T] {
	return ListValueOf[T]{ListValue: basetypes.NewListNull(newAttrTypeOf[T](ctx))}
}

func NewListValueOfUnknown[T attr.Value](ctx context.Context) ListValueOf[T] {
	return ListValueOf[T]{ListValue: basetypes.NewListUnknown(newAttrTypeOf[T](ctx))}
}

func NewListValueOf[T attr.Value](ctx context.Context, elements []attr.Value) (ListValueOf[T], diag.Diagnostics) {
	var diags diag.Diagnostics

	v, d := basetypes.NewListValue(newAttrTypeOf[T](ctx), elements)
	diags.Append(d...)
	if diags.HasError() {
		return NewListValueOfUnknown[T](ctx), diags
	}

	return ListValueOf[T]{ListValue: v}, diags
}

func NewListValueOfMust[T attr.Value](ctx context.Context, elements []attr.Value) ListValueOf[T] {
	return fwdiag.Must(NewListValueOf[T](ctx, elements))
}

func validateStringEnumSlice[T enum.Valueser[T]](ctx context.Context, path path.Path, values []attr.Value) diag.Diagnostics {
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
