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

	return NewMapFromMapValue(v), diags
}

func NewMapFromMapValue(m basetypes.MapValue) MapValue {
	return MapValue{MapValue: m}
}

// func NewMapEmpty() MapValue {
// 	return fwdiag.Must(NewMapValue(map[string]attr.Value{}))
// }

var (
	_ basetypes.MapValuable                   = (*MapValue)(nil)
	_ basetypes.MapValuableWithSemanticEquals = (*MapValue)(nil)
)

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
	// maps.DeleteFunc(elements, func(_ string, v attr.Value) bool {
	// 	return v.IsNull()
	// })
	oElements := o.Elements()
	// maps.DeleteFunc(oElements, func(_ string, v attr.Value) bool {
	// 	return v.IsNull()
	// })

	if len(elements) != len(oElements) {
		return false, diags
	}

	for k, v := range elements {
		ov, _ := oElements[k]

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
