// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tftypes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// Object is a Terraform type representing an unordered collection of
// attributes, potentially of differing types, each identifiable with a unique
// string name. The number of attributes, their names, and their types are part
// of the type signature for the Object, and so two Objects with different
// attribute names or types are considered to be distinct types.
type Object struct {
	// AttributeTypes is a map of attributes to their types. The key should
	// be the name of the attribute, and the value should be the type of
	// the attribute.
	AttributeTypes map[string]Type

	// OptionalAttributes is a set of attributes that are optional. This
	// allows values of this object type to include or not include those
	// attributes without changing the type of the value; other attributes
	// are considered part of the type signature, and their absence means a
	// value is no longer of that type.
	//
	// OptionalAttributes is only valid when declaring a type constraint
	// (e.g. Schema) and should not be used as part of a Type when creating
	// a Value (e.g. NewValue()). When creating a Value, all OptionalAttributes
	// must still be defined in the Object by setting each attribute to a null
	// or known value for its attribute type.
	//
	// The key of OptionalAttributes should be the name of the attribute
	// that is optional. The value should be an empty struct, used only to
	// indicate presence.
	//
	// OptionalAttributes must also be listed in the AttributeTypes
	// property, indicating their types.
	OptionalAttributes map[string]struct{}

	// used to make this type uncomparable
	// see https://golang.org/ref/spec#Comparison_operators
	// this enforces the use of Is, instead
	_ []struct{}
}

// ApplyTerraform5AttributePathStep applies an AttributePathStep to an Object,
// returning the Type found at that AttributePath within the Object. If the
// AttributePathStep cannot be applied to the Object, an ErrInvalidStep error
// will be returned.
func (o Object) ApplyTerraform5AttributePathStep(step AttributePathStep) (interface{}, error) {
	switch s := step.(type) {
	case AttributeName:
		if len(o.AttributeTypes) == 0 {
			return nil, ErrInvalidStep
		}

		attrType, ok := o.AttributeTypes[string(s)]

		if !ok {
			return nil, ErrInvalidStep
		}

		return attrType, nil
	default:
		return nil, ErrInvalidStep
	}
}

// Equal returns true if the two Objects are exactly equal. Unlike Is, passing
// in an Object with no AttributeTypes will always return false.
func (o Object) Equal(other Type) bool {
	v, ok := other.(Object)
	if !ok {
		return false
	}
	if v.AttributeTypes == nil || o.AttributeTypes == nil {
		// when doing exact comparisons, we can't compare types that
		// don't have attribute types set, so we just consider them not
		// equal
		return false
	}

	// if the don't have the exact same optional attributes, they're not
	// the same type.
	if len(v.OptionalAttributes) != len(o.OptionalAttributes) {
		return false
	}
	for attr := range o.OptionalAttributes {
		if !v.attrIsOptional(attr) {
			return false
		}
	}

	// if they don't have the same attribute types, they're not the
	// same type.
	if len(v.AttributeTypes) != len(o.AttributeTypes) {
		return false
	}
	for k, typ := range o.AttributeTypes {
		if _, ok := v.AttributeTypes[k]; !ok {
			return false
		}
		if !typ.Equal(v.AttributeTypes[k]) {
			return false
		}
	}
	return true
}

// UsableAs returns whether the two Objects are type compatible.
//
// If the other type is DynamicPseudoType, it will return true.
// If the other type is not a Object, it will return false.
// If the other Object does not have matching AttributeTypes length, it will
// return false.
// If the other Object does not have a type compatible ElementType for every
// nested attribute, it will return false.
//
// If the current type contains OptionalAttributes, it will panic.
func (o Object) UsableAs(other Type) bool {
	if other.Is(DynamicPseudoType) {
		return true
	}
	v, ok := other.(Object)
	if !ok {
		return false
	}
	if len(o.OptionalAttributes) > 0 {
		panic("Objects with OptionalAttributes cannot be used.")
	}
	if len(v.AttributeTypes) != len(o.AttributeTypes) {
		return false
	}
	for k, typ := range o.AttributeTypes {
		otherTyp, ok := v.AttributeTypes[k]
		if !ok {
			return false
		}
		if !typ.UsableAs(otherTyp) {
			return false
		}
	}
	return true
}

// Is returns whether `t` is an Object type or not. It does not perform any
// AttributeTypes checks.
func (o Object) Is(t Type) bool {
	_, ok := t.(Object)
	return ok
}

func (o Object) attrIsOptional(attr string) bool {
	if o.OptionalAttributes == nil {
		return false
	}
	_, ok := o.OptionalAttributes[attr]
	return ok
}

func (o Object) String() string {
	var res strings.Builder
	res.WriteString("tftypes.Object[")
	keys := make([]string, 0, len(o.AttributeTypes))
	for k := range o.AttributeTypes {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for pos, key := range keys {
		if pos != 0 {
			res.WriteString(", ")
		}
		res.WriteString(`"` + key + `":`)
		res.WriteString(o.AttributeTypes[key].String())
		if o.attrIsOptional(key) {
			res.WriteString(`?`)
		}
	}
	res.WriteString("]")
	return res.String()
}

func (o Object) private() {}

func (o Object) supportedGoTypes() []string {
	return []string{"map[string]tftypes.Value"}
}

func valueFromObject(types map[string]Type, optionalAttrs map[string]struct{}, in interface{}) (Value, error) {
	switch value := in.(type) {
	case map[string]Value:
		// types should only be null if the "Object" is actually a
		// DynamicPseudoType being created from a map[string]Value. In
		// which case, we don't know what types it should have, or even
		// how many there will be, so let's not validate that at all
		if types != nil {
			for k := range types {
				if _, ok := optionalAttrs[k]; ok {
					// if it's optional, we don't need to check that it has a value
					continue
				}
				if _, ok := value[k]; !ok {
					return Value{}, fmt.Errorf("can't create a tftypes.Value of type %s, required attribute %q not set", Object{AttributeTypes: types}, k)
				}
			}
			for k, v := range value {
				typ, ok := types[k]
				if !ok {
					return Value{}, fmt.Errorf("can't set a value on %q in tftypes.NewValue, key not part of the object type %s", k, Object{AttributeTypes: types})
				}
				if v.Type() == nil {
					return Value{}, NewAttributePath().WithAttributeName(k).NewErrorf("missing value type")
				}
				if !v.Type().UsableAs(typ) {
					return Value{}, NewAttributePath().WithAttributeName(k).NewErrorf("can't use %s as %s", v.Type(), typ)
				}
			}
		}
		return Value{
			typ:   Object{AttributeTypes: types, OptionalAttributes: optionalAttrs},
			value: value,
		}, nil
	default:
		return Value{}, fmt.Errorf("tftypes.NewValue can't use %T as a tftypes.Object; expected types are: %s", in, formattedSupportedGoTypes(Object{}))
	}
}

// MarshalJSON returns a JSON representation of the full type signature of `o`,
// including the AttributeTypes and, if present, OptionalAttributes.
//
// Deprecated: this is not meant to be called by third-party code.
func (o Object) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer

	buf.WriteString(`["object",{`)

	attributeTypeNames := make([]string, 0, len(o.AttributeTypes))

	for attributeTypeName := range o.AttributeTypes {
		attributeTypeNames = append(attributeTypeNames, attributeTypeName)
	}

	// Ensure consistent ordering for human readability and unit testing.
	// The slices package was introduced in Go 1.21, so it is not usable until
	// this Go module is updated to Go 1.21 minimum.
	sort.Strings(attributeTypeNames)

	for index, attributeTypeName := range attributeTypeNames {
		if index > 0 {
			buf.WriteString(`,`)
		}

		buf.Write(marshalJSONObjectAttributeName(attributeTypeName))
		buf.WriteString(`:`)

		// MarshalJSON is always error safe
		attributeTypeBytes, _ := o.AttributeTypes[attributeTypeName].MarshalJSON()

		buf.Write(attributeTypeBytes)
	}

	buf.WriteString(`}`)

	if len(o.OptionalAttributes) > 0 {
		buf.WriteString(`,[`)

		optionalAttributeNames := make([]string, 0, len(o.OptionalAttributes))

		for optionalAttributeName := range o.OptionalAttributes {
			optionalAttributeNames = append(optionalAttributeNames, optionalAttributeName)
		}

		// Ensure consistent ordering for human readability and unit testing.
		// The slices package was introduced in Go 1.21, so it is not usable
		// until this Go module is updated to Go 1.21 minimum.
		sort.Strings(optionalAttributeNames)

		for index, optionalAttributeName := range optionalAttributeNames {
			if index > 0 {
				buf.WriteString(`,`)
			}

			buf.Write(marshalJSONObjectAttributeName(optionalAttributeName))
		}

		buf.WriteString(`]`)
	}

	buf.WriteString(`]`)

	return buf.Bytes(), nil
}

// marshalJSONObjectAttributeName an object attribute name string into JSON or
// panics.
//
// JSON encoding a string has some non-trivial rules and go-cty already depends
// on the Go standard library for this, so for now this logic also offloads this
// effort the same way to handle user input. As of Go 1.21, it is not possible
// for a caller to input something that would trigger an encoding error. There
// is FuzzMarshalJSONObjectAttributeName to verify this assertion.
//
// If a panic can be induced, a Type Validate() method or requiring the use of
// Type construction functions that require validation are better solutions than
// handling validation errors at this point.
func marshalJSONObjectAttributeName(name string) []byte {
	result, err := json.Marshal(name)

	if err != nil {
		panic(fmt.Sprintf("unable to JSON encode object attribute name: %s", name))
	}

	return result
}
