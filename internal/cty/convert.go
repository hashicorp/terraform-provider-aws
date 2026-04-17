// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package cty

import (
	"fmt"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// Cribbed from github.com/hashicorp/terraform-plugin-sdk/internal/plugin/convert/schema.go.

func tftypeFromCtyType(in cty.Type) (tftypes.Type, error) {
	switch {
	case in.Equals(cty.String):
		return tftypes.String, nil
	case in.Equals(cty.Number):
		return tftypes.Number, nil
	case in.Equals(cty.Bool):
		return tftypes.Bool, nil
	case in.Equals(cty.DynamicPseudoType):
		return tftypes.DynamicPseudoType, nil
	case in.IsSetType():
		elemType, err := tftypeFromCtyType(in.ElementType())
		if err != nil {
			return nil, err
		}
		return tftypes.Set{
			ElementType: elemType,
		}, nil
	case in.IsListType():
		elemType, err := tftypeFromCtyType(in.ElementType())
		if err != nil {
			return nil, err
		}
		return tftypes.List{
			ElementType: elemType,
		}, nil
	case in.IsTupleType():
		elemTypes := make([]tftypes.Type, 0, in.Length())
		for _, typ := range in.TupleElementTypes() {
			elemType, err := tftypeFromCtyType(typ)
			if err != nil {
				return nil, err
			}
			elemTypes = append(elemTypes, elemType)
		}
		return tftypes.Tuple{
			ElementTypes: elemTypes,
		}, nil
	case in.IsMapType():
		elemType, err := tftypeFromCtyType(in.ElementType())
		if err != nil {
			return nil, err
		}
		return tftypes.Map{
			ElementType: elemType,
		}, nil
	case in.IsObjectType():
		attrTypes := make(map[string]tftypes.Type)
		for key, typ := range in.AttributeTypes() {
			attrType, err := tftypeFromCtyType(typ)
			if err != nil {
				return nil, err
			}
			attrTypes[key] = attrType
		}
		return tftypes.Object{
			AttributeTypes: attrTypes,
		}, nil
	}
	return nil, fmt.Errorf("unknown cty type %s", in.GoString())
}

// Cribbed from github.com/hashicorp/terraform-plugin-sdk/internal/plugin/convert/value.go.

func primitiveTfValue(in cty.Value) (*tftypes.Value, error) {
	primitiveType, err := tftypeFromCtyType(in.Type())
	if err != nil {
		return nil, err
	}

	if in.IsNull() {
		return nullTfValue(primitiveType), nil
	}

	if !in.IsKnown() {
		return unknownTfValue(primitiveType), nil
	}

	var val tftypes.Value
	switch in.Type() {
	case cty.String:
		val = tftypes.NewValue(tftypes.String, in.AsString())
	case cty.Bool:
		val = tftypes.NewValue(tftypes.Bool, in.True())
	case cty.Number:
		val = tftypes.NewValue(tftypes.Number, in.AsBigFloat())
	}

	return &val, nil
}

func listTfValue(in cty.Value) (*tftypes.Value, error) {
	listType, err := tftypeFromCtyType(in.Type())
	if err != nil {
		return nil, err
	}

	if in.IsNull() {
		return nullTfValue(listType), nil
	}

	if !in.IsKnown() {
		return unknownTfValue(listType), nil
	}

	vals := make([]tftypes.Value, 0)

	for _, v := range in.AsValueSlice() {
		tfVal, err := ToTfValue(v)
		if err != nil {
			return nil, err
		}
		vals = append(vals, *tfVal)
	}

	out := tftypes.NewValue(listType, vals)

	return &out, nil
}

func mapTfValue(in cty.Value) (*tftypes.Value, error) {
	mapType, err := tftypeFromCtyType(in.Type())
	if err != nil {
		return nil, err
	}

	if in.IsNull() {
		return nullTfValue(mapType), nil
	}

	if !in.IsKnown() {
		return unknownTfValue(mapType), nil
	}

	vals := make(map[string]tftypes.Value)

	for k, v := range in.AsValueMap() {
		tfVal, err := ToTfValue(v)
		if err != nil {
			return nil, err
		}
		vals[k] = *tfVal
	}

	out := tftypes.NewValue(mapType, vals)

	return &out, nil
}

func setTfValue(in cty.Value) (*tftypes.Value, error) {
	setType, err := tftypeFromCtyType(in.Type())
	if err != nil {
		return nil, err
	}

	if in.IsNull() {
		return nullTfValue(setType), nil
	}

	if !in.IsKnown() {
		return unknownTfValue(setType), nil
	}

	vals := make([]tftypes.Value, 0)

	for _, v := range in.AsValueSlice() {
		tfVal, err := ToTfValue(v)
		if err != nil {
			return nil, err
		}
		vals = append(vals, *tfVal)
	}

	out := tftypes.NewValue(setType, vals)

	return &out, nil
}

func objectTfValue(in cty.Value) (*tftypes.Value, error) {
	objType, err := tftypeFromCtyType(in.Type())
	if err != nil {
		return nil, err
	}

	if in.IsNull() {
		return nullTfValue(objType), nil
	}

	if !in.IsKnown() {
		return unknownTfValue(objType), nil
	}

	vals := make(map[string]tftypes.Value)

	for k, v := range in.AsValueMap() {
		tfVal, err := ToTfValue(v)
		if err != nil {
			return nil, err
		}
		vals[k] = *tfVal
	}

	out := tftypes.NewValue(objType, vals)

	return &out, nil
}

func tupleTfValue(in cty.Value) (*tftypes.Value, error) {
	tupleType, err := tftypeFromCtyType(in.Type())
	if err != nil {
		return nil, err
	}

	if in.IsNull() {
		return nullTfValue(tupleType), nil
	}

	if !in.IsKnown() {
		return unknownTfValue(tupleType), nil
	}

	vals := make([]tftypes.Value, 0)

	for _, v := range in.AsValueSlice() {
		tfVal, err := ToTfValue(v)
		if err != nil {
			return nil, err
		}
		vals = append(vals, *tfVal)
	}

	out := tftypes.NewValue(tupleType, vals)

	return &out, nil
}

func ToTfValue(in cty.Value) (*tftypes.Value, error) {
	ty := in.Type()
	switch {
	case ty.IsPrimitiveType():
		return primitiveTfValue(in)
	case ty.IsListType():
		return listTfValue(in)
	case ty.IsObjectType():
		return objectTfValue(in)
	case ty.IsMapType():
		return mapTfValue(in)
	case ty.IsSetType():
		return setTfValue(in)
	case ty.IsTupleType():
		return tupleTfValue(in)
	default:
		return nil, fmt.Errorf("unsupported type %s", ty)
	}
}

func nullTfValue(ty tftypes.Type) *tftypes.Value {
	nullValue := tftypes.NewValue(ty, nil)
	return &nullValue
}

func unknownTfValue(ty tftypes.Type) *tftypes.Value {
	unknownValue := tftypes.NewValue(ty, tftypes.UnknownValue)
	return &unknownValue
}
