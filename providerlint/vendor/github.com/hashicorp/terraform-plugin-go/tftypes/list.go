package tftypes

import (
	"fmt"
)

// List is a Terraform type representing an ordered collection of elements, all
// of the same type.
type List struct {
	ElementType Type

	// used to make this type uncomparable
	// see https://golang.org/ref/spec#Comparison_operators
	// this enforces the use of Is, instead
	_ []struct{}
}

// Equal returns true if the two Lists are exactly equal. Unlike Is, passing in
// a List with no ElementType will always return false.
func (l List) Equal(o Type) bool {
	v, ok := o.(List)
	if !ok {
		return false
	}
	if l.ElementType == nil || v.ElementType == nil {
		// when doing exact comparisons, we can't compare types that
		// don't have element types set, so we just consider them not
		// equal
		return false
	}
	return l.ElementType.Equal(v.ElementType)
}

// UsableAs returns whether the two Lists are type compatible.
//
// If the other type is DynamicPseudoType, it will return true.
// If the other type is not a List, it will return false.
// If the other List does not have a type compatible ElementType, it will
// return false.
func (l List) UsableAs(o Type) bool {
	if o.Is(DynamicPseudoType) {
		return true
	}
	v, ok := o.(List)
	if !ok {
		return false
	}
	return l.ElementType.UsableAs(v.ElementType)
}

// Is returns whether `t` is a List type or not. It does not perform any
// ElementType checks.
func (l List) Is(t Type) bool {
	_, ok := t.(List)
	return ok
}

func (l List) String() string {
	return "tftypes.List[" + l.ElementType.String() + "]"
}

func (l List) private() {}

func (l List) supportedGoTypes() []string {
	return []string{"[]tftypes.Value"}
}

func valueFromList(typ Type, in interface{}) (Value, error) {
	switch value := in.(type) {
	case []Value:
		var valType Type
		for pos, v := range value {
			if !v.Type().UsableAs(typ) {
				return Value{}, NewAttributePath().WithElementKeyInt(pos).NewErrorf("can't use %s as %s", v.Type(), typ)
			}
			if valType == nil {
				valType = v.Type()
			}
			if !v.Type().Equal(valType) {
				return Value{}, fmt.Errorf("lists must only contain one type of element, saw %s and %s", valType, v.Type())
			}
		}
		return Value{
			typ:   List{ElementType: typ},
			value: value,
		}, nil
	default:
		return Value{}, fmt.Errorf("tftypes.NewValue can't use %T as a tftypes.List; expected types are: %s", in, formattedSupportedGoTypes(List{}))
	}
}

// MarshalJSON returns a JSON representation of the full type signature of `l`,
// including its ElementType.
//
// Deprecated: this is not meant to be called by third-party code.
func (l List) MarshalJSON() ([]byte, error) {
	elementType, err := l.ElementType.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("error marshaling tftypes.List's element type %T to JSON: %w", l.ElementType, err)
	}
	return []byte(`["list",` + string(elementType) + `]`), nil
}
