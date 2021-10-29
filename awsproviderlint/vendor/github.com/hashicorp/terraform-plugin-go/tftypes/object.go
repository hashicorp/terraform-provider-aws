package tftypes

import (
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
				if v.Type().Is(DynamicPseudoType) && v.IsKnown() {
					return Value{}, NewAttributePath().WithAttributeName(k).NewErrorf("invalid value %s for %s", v, v.Type())
				} else if !v.Type().Is(DynamicPseudoType) && !v.Type().UsableAs(typ) {
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
// including the AttributeTypes.
//
// Deprecated: this is not meant to be called by third-party code.
func (o Object) MarshalJSON() ([]byte, error) {
	attrs, err := json.Marshal(o.AttributeTypes)
	if err != nil {
		return nil, err
	}
	var optionalAttrs []byte
	if len(o.OptionalAttributes) > 0 {
		optionalAttrs = append(optionalAttrs, []byte(",")...)
		names := make([]string, 0, len(o.OptionalAttributes))
		for k := range o.OptionalAttributes {
			names = append(names, k)
		}
		sort.Strings(names)
		optionalsJSON, err := json.Marshal(names)
		if err != nil {
			return nil, err
		}
		optionalAttrs = append(optionalAttrs, optionalsJSON...)
	}
	return []byte(`["object",` + string(attrs) + string(optionalAttrs) + `]`), nil
}
