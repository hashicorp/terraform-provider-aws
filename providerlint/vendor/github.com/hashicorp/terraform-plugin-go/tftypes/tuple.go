package tftypes

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Tuple is a Terraform type representing an ordered collection of elements,
// potentially of differing types. The number of elements and their types are
// part of the type signature for the Tuple, and so two Tuples with different
// numbers or types of elements are considered to be distinct types.
type Tuple struct {
	ElementTypes []Type

	// used to make this type uncomparable
	// see https://golang.org/ref/spec#Comparison_operators
	// this enforces the use of Is, instead
	_ []struct{}
}

// Equal returns true if the two Tuples are exactly equal. Unlike Is, passing
// in a Tuple with no ElementTypes will always return false.
func (tu Tuple) Equal(o Type) bool {
	v, ok := o.(Tuple)
	if !ok {
		return false
	}
	if v.ElementTypes == nil || tu.ElementTypes == nil {
		// when doing exact comparisons, we can't compare types that
		// don't have element types set, so we just consider them not
		// equal
		return false
	}
	if len(v.ElementTypes) != len(tu.ElementTypes) {
		return false
	}
	for pos, typ := range tu.ElementTypes {
		if !typ.Equal(v.ElementTypes[pos]) {
			return false
		}
	}
	return true
}

// UsableAs returns whether the two Tuples are type compatible.
//
// If the other type is DynamicPseudoType, it will return true.
// If the other type is not a Tuple, it will return false.
// If the other Tuple does not have matching ElementTypes length, it will
// return false.
// If the other Tuple does not have type compatible ElementTypes in each
// position, it will return false.
func (tu Tuple) UsableAs(o Type) bool {
	if o.Is(DynamicPseudoType) {
		return true
	}
	v, ok := o.(Tuple)
	if !ok {
		return false
	}
	if len(v.ElementTypes) != len(tu.ElementTypes) {
		return false
	}
	for pos, typ := range tu.ElementTypes {
		if !typ.UsableAs(v.ElementTypes[pos]) {
			return false
		}
	}
	return true
}

// Is returns whether `t` is a Tuple type or not. It does not perform any
// ElementTypes checks.
func (tu Tuple) Is(t Type) bool {
	_, ok := t.(Tuple)
	return ok
}

func (tu Tuple) String() string {
	var res strings.Builder
	res.WriteString("tftypes.Tuple[")
	for pos, t := range tu.ElementTypes {
		if pos != 0 {
			res.WriteString(", ")
		}
		res.WriteString(t.String())
	}
	res.WriteString("]")
	return res.String()
}

func (tu Tuple) private() {}

func (tu Tuple) supportedGoTypes() []string {
	return []string{"[]tftypes.Value"}
}

func valueFromTuple(types []Type, in interface{}) (Value, error) {
	switch value := in.(type) {
	case []Value:
		// types should only be null if the "Tuple" is actually a
		// DynamicPseudoType being created from a []Value. In which
		// case, we don't know what types it should have, or even how
		// many there will be, so let's not validate that at all
		if types != nil {
			if len(types) != len(value) {
				return Value{}, fmt.Errorf("can't create a tftypes.Value with %d elements, type %s requires %d elements", len(value), Tuple{ElementTypes: types}, len(types))
			}
			for pos, v := range value {
				typ := types[pos]
				if !v.Type().UsableAs(typ) {
					return Value{}, NewAttributePath().WithElementKeyInt(pos).NewErrorf("can't use %s as %s", v.Type(), typ)
				}
			}
		}
		return Value{
			typ:   Tuple{ElementTypes: types},
			value: value,
		}, nil
	default:
		return Value{}, fmt.Errorf("tftypes.NewValue can't use %T as a tftypes.Tuple; expected types are: %s", in, formattedSupportedGoTypes(Tuple{}))
	}
}

// MarshalJSON returns a JSON representation of the full type signature of
// `tu`, including the ElementTypes.
//
// Deprecated: this is not meant to be called by third-party code.
func (tu Tuple) MarshalJSON() ([]byte, error) {
	elements, err := json.Marshal(tu.ElementTypes)
	if err != nil {
		return nil, err
	}
	return []byte(`["tuple",` + string(elements) + `]`), nil
}
