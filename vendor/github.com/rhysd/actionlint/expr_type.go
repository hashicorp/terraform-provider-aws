package actionlint

import (
	"fmt"
	"sort"
	"strings"
)

// Types

// ExprType is interface for types of values in expression.
type ExprType interface {
	// String returns string representation of the type.
	String() string
	// Assignable returns if other type can be assignable to the type.
	Assignable(other ExprType) bool
	// Merge merges other type into this type. When other type conflicts with this type, the merged
	// result is any type as fallback.
	Merge(other ExprType) ExprType
	// DeepCopy duplicates itself. All its child types are copied recursively.
	DeepCopy() ExprType
}

// AnyType represents type which can be any type. It also indicates that a value of the type cannot
// be type-checked since it's type cannot be known statically.
type AnyType struct{}

func (ty AnyType) String() string {
	return "any"
}

// Assignable returns if other type can be assignable to the type.
func (ty AnyType) Assignable(_ ExprType) bool {
	return true
}

// Merge merges other type into this type. When other type conflicts with this type, the merged
// result is any type as fallback.
func (ty AnyType) Merge(other ExprType) ExprType {
	return ty
}

// DeepCopy duplicates itself. All its child types are copied recursively.
func (ty AnyType) DeepCopy() ExprType {
	return ty
}

// NullType is type for null value.
type NullType struct{}

func (ty NullType) String() string {
	return "null"
}

// Assignable returns if other type can be assignable to the type.
func (ty NullType) Assignable(other ExprType) bool {
	switch other.(type) {
	case NullType, AnyType:
		return true
	default:
		return false
	}
}

// Merge merges other type into this type. When other type conflicts with this type, the merged
// result is any type as fallback.
func (ty NullType) Merge(other ExprType) ExprType {
	if _, ok := other.(NullType); ok {
		return ty
	}
	return AnyType{}
}

// DeepCopy duplicates itself. All its child types are copied recursively.
func (ty NullType) DeepCopy() ExprType {
	return ty
}

// NumberType is type for number values such as integer or float.
type NumberType struct{}

func (ty NumberType) String() string {
	return "number"
}

// Assignable returns if other type can be assignable to the type.
func (ty NumberType) Assignable(other ExprType) bool {
	// TODO: Is string of numbers corced into number?
	switch other.(type) {
	case NumberType, AnyType:
		return true
	default:
		return false
	}
}

// Merge merges other type into this type. When other type conflicts with this type, the merged
// result is any type as fallback.
func (ty NumberType) Merge(other ExprType) ExprType {
	switch other.(type) {
	case NumberType:
		return ty
	case StringType:
		return other
	default:
		return AnyType{}
	}
}

// DeepCopy duplicates itself. All its child types are copied recursively.
func (ty NumberType) DeepCopy() ExprType {
	return ty
}

// BoolType is type for boolean values.
type BoolType struct{}

func (ty BoolType) String() string {
	return "bool"
}

// Assignable returns if other type can be assignable to the type.
func (ty BoolType) Assignable(other ExprType) bool {
	// Any type can be converted into bool..
	// e.g.
	//    if: ${{ steps.foo }}
	return true
}

// Merge merges other type into this type. When other type conflicts with this type, the merged
// result is any type as fallback.
func (ty BoolType) Merge(other ExprType) ExprType {
	switch other.(type) {
	case BoolType:
		return ty
	case StringType:
		return other
	default:
		return AnyType{}
	}
}

// DeepCopy duplicates itself. All its child types are copied recursively.
func (ty BoolType) DeepCopy() ExprType {
	return ty
}

// StringType is type for string values.
type StringType struct{}

func (ty StringType) String() string {
	return "string"
}

// Assignable returns if other type can be assignable to the type.
func (ty StringType) Assignable(other ExprType) bool {
	// Bool and null types also can be coerced into string. But in almost all case, those coercing
	// would be mistakes.
	switch other.(type) {
	case StringType, NumberType, AnyType:
		return true
	default:
		return false
	}
}

// Merge merges other type into this type. When other type conflicts with this type, the merged
// result is any type as fallback.
func (ty StringType) Merge(other ExprType) ExprType {
	switch other.(type) {
	case StringType, NumberType, BoolType:
		return ty
	default:
		return AnyType{}
	}
}

// DeepCopy duplicates itself. All its child types are copied recursively.
func (ty StringType) DeepCopy() ExprType {
	return ty
}

// ObjectType is type for objects, which can hold key-values.
type ObjectType struct {
	// Props is map from properties name to their type.
	Props map[string]ExprType
	// Mapped is an element type of this object. This means all props have the type. For example,
	// The element type of env context is string.
	// AnyType means its property types can be any type so it shapes a loose object. Setting nil
	// means propreties are mapped to no type so it shapes a strict object.
	//
	// Invariant: All types in Props field must be assignable to this type.
	Mapped ExprType
}

// NewEmptyObjectType creates new loose ObjectType instance which allows unknown props. When
// accessing to unknown props, their values will fall back to any.
func NewEmptyObjectType() *ObjectType {
	return &ObjectType{map[string]ExprType{}, AnyType{}}
}

// NewObjectType creates new loose ObjectType instance which allows unknown props with given props.
func NewObjectType(props map[string]ExprType) *ObjectType {
	return &ObjectType{props, AnyType{}}
}

// NewEmptyStrictObjectType creates new ObjectType instance which does not allow unknown props.
func NewEmptyStrictObjectType() *ObjectType {
	return &ObjectType{map[string]ExprType{}, nil}
}

// NewStrictObjectType creates new ObjectType instance which does not allow unknown props with
// given prop types.
func NewStrictObjectType(props map[string]ExprType) *ObjectType {
	return &ObjectType{props, nil}
}

// NewMapObjectType creates new ObjectType which maps keys to a specific type value.
func NewMapObjectType(t ExprType) *ObjectType {
	return &ObjectType{nil, t}
}

// IsStrict returns if the type is a strict object, which means no unknown prop is allowed.
func (ty *ObjectType) IsStrict() bool {
	return ty.Mapped == nil
}

// IsLoose returns if the type is a loose object, which allows any unknown props.
func (ty *ObjectType) IsLoose() bool {
	_, ok := ty.Mapped.(AnyType)
	return ok
}

// Strict sets the object is strict, which means only known properties are allowed.
func (ty *ObjectType) Strict() {
	ty.Mapped = nil
}

// Loose sets the object is loose, which means any properties can be set.
func (ty *ObjectType) Loose() {
	ty.Mapped = AnyType{}
}

func (ty *ObjectType) String() string {
	if !ty.IsStrict() {
		if ty.IsLoose() {
			return "object"
		}
		return fmt.Sprintf("{string => %s}", ty.Mapped.String())
	}

	ps := make([]string, 0, len(ty.Props))
	for n := range ty.Props {
		ps = append(ps, n)
	}
	sort.Strings(ps)

	var b strings.Builder
	b.WriteByte('{')
	first := true
	for _, p := range ps {
		if first {
			first = false
		} else {
			b.WriteString("; ")
		}
		b.WriteString(p)
		b.WriteString(": ")
		b.WriteString(ty.Props[p].String())
	}
	b.WriteByte('}')

	return b.String()
}

// Assignable returns if other type can be assignable to the type.
// In other words, rhs type is more strict than lhs (receiver) type.
func (ty *ObjectType) Assignable(other ExprType) bool {
	switch other := other.(type) {
	case AnyType:
		return true
	case *ObjectType:
		if !ty.IsStrict() {
			if !other.IsStrict() {
				return ty.Mapped.Assignable(other.Mapped)
			}
			for _, t := range other.Props {
				if !ty.Mapped.Assignable(t) {
					return false
				}
			}
			return true
		}
		// ty is strict

		if !other.IsStrict() {
			for _, t := range ty.Props {
				if !t.Assignable(other.Mapped) {
					return false
				}
			}
			return true
		}
		// ty and other are strict

		for n, r := range other.Props {
			if l, ok := ty.Props[n]; !ok || !l.Assignable(r) {
				return false
			}
		}

		return true
	default:
		return false
	}
}

// Merge merges two object types into one. When other object has unknown props, they are merged into
// current object. When both have same property, when they are assignable, it remains as-is.
// Otherwise, the property falls back to any type.
func (ty *ObjectType) Merge(other ExprType) ExprType {
	switch other := other.(type) {
	case *ObjectType:
		// Shortcuts
		if len(ty.Props) == 0 && other.IsLoose() {
			return other
		}
		if len(other.Props) == 0 && ty.IsLoose() {
			return ty
		}

		mapped := ty.Mapped
		if mapped == nil {
			mapped = other.Mapped
		} else if other.Mapped != nil {
			mapped = mapped.Merge(other.Mapped)
		}

		props := make(map[string]ExprType, len(ty.Props))
		for n, l := range ty.Props {
			props[n] = l
		}
		for n, r := range other.Props {
			if l, ok := props[n]; ok {
				props[n] = l.Merge(r)
			} else {
				props[n] = r
				if mapped != nil {
					mapped = mapped.Merge(r)
				}
			}
		}

		return &ObjectType{
			Props:  props,
			Mapped: mapped,
		}
	default:
		return AnyType{}
	}
}

// DeepCopy duplicates itself. All its child types are copied recursively.
func (ty *ObjectType) DeepCopy() ExprType {
	p := make(map[string]ExprType, len(ty.Props))
	for n, t := range ty.Props {
		p[n] = t.DeepCopy()
	}
	m := ty.Mapped
	if m != nil {
		m = m.DeepCopy()
	}
	return &ObjectType{p, m}
}

// ArrayType is type for arrays.
type ArrayType struct {
	// Elem is type of element of the array.
	Elem ExprType
	// Deref is true when this type was derived from object filtering syntax (foo.*).
	Deref bool
}

func (ty *ArrayType) String() string {
	return fmt.Sprintf("array<%s>", ty.Elem.String())
}

// Assignable returns if other type can be assignable to the type.
func (ty *ArrayType) Assignable(other ExprType) bool {
	switch other := other.(type) {
	case AnyType:
		return true
	case *ArrayType:
		return ty.Elem.Assignable(other.Elem)
	default:
		return false
	}
}

// Merge merges two object types into one. When other object has unknown props, they are merged into
// current object. When both have same property, when they are assignable, it remains as-is.
// Otherwise, the property falls back to any type.
func (ty *ArrayType) Merge(other ExprType) ExprType {
	switch other := other.(type) {
	case *ArrayType:
		if _, ok := ty.Elem.(AnyType); ok {
			return ty
		}
		if _, ok := other.Elem.(AnyType); ok {
			return other
		}
		return &ArrayType{
			Elem:  ty.Elem.Merge(other.Elem),
			Deref: false, // When fusing array deref type, it means prop deref chain breaks
		}
	default:
		return AnyType{}
	}
}

// DeepCopy duplicates itself. All its child types are copied recursively.
func (ty *ArrayType) DeepCopy() ExprType {
	return &ArrayType{ty.Elem.DeepCopy(), ty.Deref}
}

// EqualTypes returns if the two types are equal.
func EqualTypes(l, r ExprType) bool {
	return l.Assignable(r) && r.Assignable(l)
}
