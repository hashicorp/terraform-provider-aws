// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tftypes

import (
	"fmt"
	"sort"
)

// Map is a Terraform type representing an unordered collection of elements,
// all of the same type, each identifiable with a unique string key.
type Map struct {
	ElementType Type

	// used to make this type uncomparable
	// see https://golang.org/ref/spec#Comparison_operators
	// this enforces the use of Is, instead
	_ []struct{}
}

// ApplyTerraform5AttributePathStep applies an AttributePathStep to a Map,
// returning the Type found at that AttributePath within the Map. If the
// AttributePathStep cannot be applied to the Map, an ErrInvalidStep error
// will be returned.
func (m Map) ApplyTerraform5AttributePathStep(step AttributePathStep) (interface{}, error) {
	switch step.(type) {
	case ElementKeyString:
		return m.ElementType, nil
	default:
		return nil, ErrInvalidStep
	}
}

// Equal returns true if the two Maps are exactly equal. Unlike Is, passing in
// a Map with no ElementType will always return false.
func (m Map) Equal(o Type) bool {
	v, ok := o.(Map)
	if !ok {
		return false
	}
	if v.ElementType == nil || m.ElementType == nil {
		// when doing exact comparisons, we can't compare types that
		// don't have element types set, so we just consider them not
		// equal
		return false
	}
	return m.ElementType.Equal(v.ElementType)
}

// UsableAs returns whether the two Maps are type compatible.
//
// If the other type is DynamicPseudoType, it will return true.
// If the other type is not a Map, it will return false.
// If the other Map does not have a type compatible ElementType, it will
// return false.
func (m Map) UsableAs(o Type) bool {
	if o.Is(DynamicPseudoType) {
		return true
	}
	v, ok := o.(Map)
	if !ok {
		return false
	}
	return m.ElementType.UsableAs(v.ElementType)
}

// Is returns whether `t` is a Map type or not. It does not perform any
// ElementType checks.
func (m Map) Is(t Type) bool {
	_, ok := t.(Map)
	return ok
}

func (m Map) String() string {
	return "tftypes.Map[" + m.ElementType.String() + "]"
}

func (m Map) private() {}

func (m Map) supportedGoTypes() []string {
	return []string{"map[string]tftypes.Value"}
}

// MarshalJSON returns a JSON representation of the full type signature of `m`,
// including its ElementType.
//
// Deprecated: this is not meant to be called by third-party code.
func (m Map) MarshalJSON() ([]byte, error) {
	attributeType, err := m.ElementType.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("error marshaling tftypes.Map's attribute type %T to JSON: %w", m.ElementType, err)
	}
	return []byte(`["map",` + string(attributeType) + `]`), nil
}

func valueFromMap(typ Type, in interface{}) (Value, error) {
	switch value := in.(type) {
	case map[string]Value:
		keys := make([]string, 0, len(value))
		for k := range value {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		var elType Type
		for _, k := range keys {
			v := value[k]
			if !v.Type().UsableAs(typ) {
				return Value{}, NewAttributePath().WithElementKeyString(k).NewErrorf("can't use %s as %s", v.Type(), typ)
			}
			if elType == nil {
				elType = v.Type()
			}
			if !elType.Equal(v.Type()) {
				return Value{}, fmt.Errorf("maps must only contain one type of element, saw %s and %s", elType, v.Type())
			}
		}
		return Value{
			typ:   Map{ElementType: typ},
			value: value,
		}, nil
	default:
		return Value{}, fmt.Errorf("tftypes.NewValue can't use %T as a tftypes.Map; expected types are: %s", in, formattedSupportedGoTypes(Map{}))
	}
}
