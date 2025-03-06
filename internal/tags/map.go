// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tags

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

var (
	_ basetypes.MapTypable = (*mapType)(nil)
)

type mapType struct {
	basetypes.MapType
}

var (
	MapType = newMapType()
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
	return "TagsMapType"
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

	return Map{MapValue: mapValue}, diags
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
	return Map{}
}

func NewMapValueNull() Map {
	return Map{basetypes.NewMapNull(basetypes.StringType{})}
}

func NewMapValueUnknown() Map {
	return Map{basetypes.NewMapUnknown(basetypes.StringType{})}
}

func NewMapValue(elements map[string]attr.Value) (Map, diag.Diagnostics) {
	var diags diag.Diagnostics

	v, d := basetypes.NewMapValue(basetypes.StringType{}, elements)
	diags.Append(d...)
	if diags.HasError() {
		return NewMapValueUnknown(), diags
	}

	return NewMapFromMapValue(v), diags
}

func NewMapFromMapValue(m basetypes.MapValue) Map {
	return Map{MapValue: m}
}

var (
	_ basetypes.MapValuable                   = (*Map)(nil)
	_ basetypes.MapValuableWithSemanticEquals = (*Map)(nil)
)

type Map struct {
	basetypes.MapValue
}

//nolint:contextcheck // To match attr.Value interface
func (v Map) Equal(o attr.Value) bool {
	other, ok := o.(Map)
	if !ok {
		return false
	}

	ctx := context.Background()

	if !v.ElementType(ctx).Equal(other.ElementType(ctx)) {
		return false
	}

	return v.MapValue.Equal(other.MapValue)
}

func (v Map) Type(ctx context.Context) attr.Type {
	return MapType
}

func (v Map) MapSemanticEquals(ctx context.Context, oValuable basetypes.MapValuable) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	o, ok := oValuable.(Map)
	if !ok {
		return false, diags
	}

	elements := v.Elements()
	oElements := o.Elements()

	if len(elements) != len(oElements) {
		return false, diags
	}

	for k, v := range elements {
		ov := oElements[k]

		if v.IsNull() {
			if !ov.IsUnknown() && !ov.IsNull() {
				sv := ov.(types.String)
				s := sv.ValueString()
				if s != "" {
					return false, diags
				}
			}
		} else if ov.IsNull() {
			if !v.IsUnknown() && !v.IsNull() {
				sv := v.(types.String)
				s := sv.ValueString()
				if s != "" {
					return false, diags
				}
			}
		} else if !v.Equal(ov) {
			return false, diags
		}
	}

	return true, diags
}

// IsWhollyKnown returns true if the map is known and all of its elements are known.
func (v Map) IsWhollyKnown() bool {
	if v.IsUnknown() {
		return false
	}

	for _, elem := range v.Elements() {
		if elem.IsUnknown() {
			return false
		}
	}

	return true
}

// FlattenStringValueMap returns an empty, non-nil Map when the source map has no elements.
func FlattenStringValueMap(ctx context.Context, m map[string]string) Map {
	elems := make(map[string]attr.Value, len(m))

	for k, v := range m {
		elems[k] = types.StringValue(v)
	}

	return NewMapFromMapValue(types.MapValueMust(types.StringType, elems))
}
