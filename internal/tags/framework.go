// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tags

import (
	"context"
	"fmt"
	"maps"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/hashicorp/terraform-provider-aws/internal/errs/fwdiag"
)

// Terraform Plugin Framework variants of tags schemas.

func TagsAttribute() schema.Attribute {
	return schema.MapAttribute{
		ElementType: types.StringType,
		CustomType:  MapType,
		Optional:    true,
		Computed:    true,
		// PlanModifiers: []planmodifier.Map{
		// 	x{},
		// },
		Default: mapdefault.StaticValue(Empty.MapValue),
	}
}

func TagsAttributeComputedOnly() schema.Attribute {
	return schema.MapAttribute{
		ElementType: types.StringType,
		Computed:    true,
	}
}

var (
	Unknown = types.MapUnknown(types.StringType)
)

var (
	Empty = NewMapEmpty()
)

// var (
// 	_ planmodifier.Map = x{}
// )

// type x struct{}

// // Description implements planmodifier.Map.
// func (x x) Description(context.Context) string {
// 	return ""
// }

// // MarkdownDescription implements planmodifier.Map.
// func (x x) MarkdownDescription(context.Context) string {
// 	return ""
// }

// // PlanModifyMap implements planmodifier.Map.
// func (x x) PlanModifyMap(ctx context.Context, req planmodifier.MapRequest, resp *planmodifier.MapResponse) {
// 	// Do not replace on resource destroy.
// 	if req.Plan.Raw.IsNull() {
// 		return
// 	}

// 	// Do nothing if there is an unknown configuration value
// 	if req.ConfigValue.IsUnknown() {
// 		return
// 	}

// 	e := req.ConfigValue.Elements()
// 	l := tfmaps.Values(e)
// 	if slices.All(l, func(v attr.Value) bool {
// 		return v.IsNull()
// 	}) {
// 		resp.PlanValue = Empty.MapValue
// 		return
// 	}
// }

var (
	_ basetypes.MapTypable                    = (*mapType)(nil)
	_ basetypes.MapValuable                   = (*MapValue)(nil)
	_ basetypes.MapValuableWithSemanticEquals = (*MapValue)(nil)
)

type mapType struct {
	basetypes.MapType
}

var (
	MapType mapType = newMapType()
)

func newMapType() mapType {
	return mapType{basetypes.MapType{ElemType: basetypes.StringType{}}}
}

func (t mapType) Equal(o attr.Type) bool {
	other, ok := o.(mapType)
	if !ok {
		return false
	}

	return t.MapType.Equal(other.MapType)
}

func (t mapType) String() string {
	return "fooMapType"
}

func (t mapType) ValueFromMap(ctx context.Context, in basetypes.MapValue) (basetypes.MapValuable, diag.Diagnostics) {
	var diags diag.Diagnostics

	if in.IsNull() {
		return NewMapValueNull(), diags
	}

	if in.IsUnknown() {
		return NewMapValueUnknown(), diags
	}

	mapValue, d := basetypes.NewMapValue(basetypes.StringType{}, in.Elements())
	diags.Append(d...)
	if diags.HasError() {
		return NewMapValueUnknown(), diags
	}

	return MapValue{MapValue: mapValue}, diags
}

func (t mapType) ValueFromTerraform(ctx context.Context, in tftypes.Value) (attr.Value, error) {
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

func (t mapType) ValueType(ctx context.Context) attr.Value {
	return MapValue{}
}

func NewMapValueNull() MapValue {
	return MapValue{basetypes.NewMapNull(basetypes.StringType{})}
}

func NewMapValueUnknown() MapValue {
	return MapValue{basetypes.NewMapUnknown(basetypes.StringType{})}
}

func NewMapValue(elements map[string]attr.Value) (MapValue, diag.Diagnostics) {
	var diags diag.Diagnostics

	v, d := basetypes.NewMapValue(basetypes.StringType{}, elements)
	diags.Append(d...)
	if diags.HasError() {
		return NewMapValueUnknown(), diags
	}

	return MapValue{MapValue: v}, diags
}

func NewMapEmpty() MapValue {
	return fwdiag.Must(NewMapValue(map[string]attr.Value{}))
}

type MapValue struct {
	basetypes.MapValue
}

func (v MapValue) Equal(o attr.Value) bool {
	other, ok := o.(MapValue)
	if !ok {
		return false
	}

	ctx := context.Background()

	if !v.ElementType(ctx).Equal(other.ElementType(ctx)) {
		return false
	}

	return v.MapValue.Equal(other.MapValue)
}

func (v MapValue) Type(ctx context.Context) attr.Type {
	return MapType
}

func (v MapValue) MapSemanticEquals(ctx context.Context, oValuable basetypes.MapValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	o, ok := oValuable.(MapValue)
	if !ok {
		return false, diags
	}

	elements := v.Elements()
	maps.DeleteFunc(elements, func(_ string, v attr.Value) bool {
		return v.IsNull()
	})
	oElements := o.Elements()
	maps.DeleteFunc(oElements, func(_ string, v attr.Value) bool {
		return v.IsNull()
	})

	if len(elements) != len(oElements) {
		return false, diags
	}

	for k, v := range elements {
		ov, _ := oElements[k]

		if !v.Equal(ov) {
			return false, diags
		}
	}

	return true, diags
}
