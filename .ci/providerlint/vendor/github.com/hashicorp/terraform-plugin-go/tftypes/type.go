// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tftypes

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

// Type is an interface representing a Terraform type. It is only meant to be
// implemented by the tftypes package. Types define the shape and
// characteristics of data coming from or being sent to Terraform.
type Type interface {
	// AttributePathStepper requires each Type to implement the
	// ApplyTerraform5AttributePathStep method, so Type is compatible with
	// WalkAttributePath. The method should return the Type found at that
	// AttributePath within the Type or ErrInvalidStep.
	AttributePathStepper

	// Is is used to determine what type a Type implementation is. It is
	// the recommended method for determining whether two types are
	// equivalent or not.

	// Is performs shallow type equality checks, in that the root type is
	// compared, but underlying attribute/element types are not.
	Is(Type) bool

	// Equal performs deep type equality checks, including attribute/element
	// types and whether attributes are optional or not.
	Equal(Type) bool

	// UsableAs performs type conformance checks. This primarily checks if the
	// target implements DynamicPsuedoType in a compatible manner.
	UsableAs(Type) bool

	// String returns a string representation of the Type's name.
	String() string

	// MarshalJSON returns a JSON representation of the Type's signature.
	// It is modeled based on Terraform's requirements for type signature
	// JSON representations, and may change over time to match Terraform's
	// formatting.
	//
	// Deprecated: this is not meant to be called by third-party code.
	MarshalJSON() ([]byte, error)

	// private is meant to keep this interface from being implemented by
	// types from other packages.
	private()

	// supportedGoTypes returns a list of string representations of the Go
	// types that the Type supports for its values.
	supportedGoTypes() []string
}

// TypeFromElements returns the common type that the passed elements all have
// in common. An error will be returned if the passed elements are not of the
// same type.
func TypeFromElements(elements []Value) (Type, error) {
	var typ Type
	for _, el := range elements {
		if typ == nil {
			typ = el.Type()
			continue
		}
		if !typ.Equal(el.Type()) {
			return nil, errors.New("elements do not all have the same types")
		}
	}
	if typ == nil {
		return DynamicPseudoType, nil
	}
	return typ, nil
}

type jsonType struct {
	t Type
}

// ParseJSONType returns a Type from its JSON representation. The JSON
// representation should come from Terraform or from MarshalJSON as the format
// is not part of this package's API guarantees.
//
// Deprecated: this is not meant to be called by third-party code.
func ParseJSONType(buf []byte) (Type, error) {
	var t jsonType
	err := json.Unmarshal(buf, &t)
	return t.t, err
}

func (t *jsonType) UnmarshalJSON(buf []byte) error {
	r := bytes.NewReader(buf)
	dec := json.NewDecoder(r)

	tok, err := dec.Token()
	if err != nil {
		return err
	}

	switch v := tok.(type) {
	case string:
		switch v {
		case "bool":
			t.t = Bool
		case "number":
			t.t = Number
		case "string":
			t.t = String
		case "dynamic":
			t.t = DynamicPseudoType
		default:
			return fmt.Errorf("invalid primitive type name %q", v)
		}

		if dec.More() {
			return fmt.Errorf("extraneous data after type description")
		}
		return nil
	case json.Delim:
		if rune(v) != '[' {
			return fmt.Errorf("invalid complex type description")
		}

		tok, err = dec.Token()
		if err != nil {
			return err
		}

		kind, ok := tok.(string)
		if !ok {
			return fmt.Errorf("invalid complex type kind name")
		}

		switch kind {
		case "list":
			var ety jsonType
			err = dec.Decode(&ety)
			if err != nil {
				return err
			}
			t.t = List{
				ElementType: ety.t,
			}
		case "map":
			var ety jsonType
			err = dec.Decode(&ety)
			if err != nil {
				return err
			}
			t.t = Map{
				ElementType: ety.t,
			}
		case "set":
			var ety jsonType
			err = dec.Decode(&ety)
			if err != nil {
				return err
			}
			t.t = Set{
				ElementType: ety.t,
			}
		case "object":
			var atys map[string]jsonType
			err = dec.Decode(&atys)
			if err != nil {
				return err
			}
			types := make(map[string]Type, len(atys))
			for k, v := range atys {
				types[k] = v.t
			}
			o := Object{
				AttributeTypes:     types,
				OptionalAttributes: map[string]struct{}{},
			}
			if dec.More() {
				var optionals []string
				err = dec.Decode(&optionals)
				if err != nil {
					return err
				}
				for _, attr := range optionals {
					o.OptionalAttributes[attr] = struct{}{}
				}
			}
			t.t = o
		case "tuple":
			var etys []jsonType
			err = dec.Decode(&etys)
			if err != nil {
				return err
			}
			types := make([]Type, 0, len(etys))
			for _, ty := range etys {
				types = append(types, ty.t)
			}
			t.t = Tuple{
				ElementTypes: types,
			}
		default:
			return fmt.Errorf("invalid complex type kind name")
		}

		tok, err = dec.Token()
		if err != nil {
			return err
		}
		if delim, ok := tok.(json.Delim); !ok || rune(delim) != ']' || dec.More() {
			return fmt.Errorf("unexpected extra data in type description")
		}

		return nil

	default:
		return fmt.Errorf("invalid type description")
	}
}

func formattedSupportedGoTypes(t Type) string {
	sgt := t.supportedGoTypes()
	switch len(sgt) {
	case 0:
		return "no supported Go types"
	case 1:
		return sgt[0]
	case 2:
		return sgt[0] + " or " + sgt[1]
	default:
		sgt[len(sgt)-1] = "or " + sgt[len(sgt)-1]
		return strings.Join(sgt, ", ")
	}
}
