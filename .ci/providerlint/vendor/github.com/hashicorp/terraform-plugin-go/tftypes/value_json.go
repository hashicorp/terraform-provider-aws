// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package tftypes

import (
	"bytes"
	"encoding/json"
	"math/big"
	"strings"
)

// ValueFromJSON returns a Value from the JSON-encoded bytes, using the
// provided Type to determine what shape the Value should be.
// DynamicPseudoTypes will be transparently parsed into the types they
// represent.
//
// Deprecated: this function is exported for internal use in
// terraform-plugin-go.  Third parties should not use it, and its behavior is
// not covered under the API compatibility guarantees. Don't use this.
func ValueFromJSON(data []byte, typ Type) (Value, error) {
	return jsonUnmarshal(data, typ, NewAttributePath(), ValueFromJSONOpts{})
}

// ValueFromJSONOpts contains options that can be used to modify the behaviour when
// unmarshalling JSON.
type ValueFromJSONOpts struct {
	// IgnoreUndefinedAttributes is used to ignore any attributes which appear in the
	// JSON but do not have a corresponding entry in the schema. For example, raw state
	// where an attribute has been removed from the schema.
	IgnoreUndefinedAttributes bool
}

// ValueFromJSONWithOpts is identical to ValueFromJSON with the exception that it
// accepts ValueFromJSONOpts which can be used to modify the unmarshalling behaviour, such
// as ignoring undefined attributes, for instance. This can occur when the JSON
// being unmarshalled does not have a corresponding attribute in the schema.
func ValueFromJSONWithOpts(data []byte, typ Type, opts ValueFromJSONOpts) (Value, error) {
	return jsonUnmarshal(data, typ, NewAttributePath(), opts)
}

func jsonByteDecoder(buf []byte) *json.Decoder {
	r := bytes.NewReader(buf)
	dec := json.NewDecoder(r)
	dec.UseNumber()
	return dec
}

func jsonUnmarshal(buf []byte, typ Type, p *AttributePath, opts ValueFromJSONOpts) (Value, error) {
	dec := jsonByteDecoder(buf)

	tok, err := dec.Token()
	if err != nil {
		return Value{}, p.NewErrorf("error reading token: %w", err)
	}

	if tok == nil {
		return NewValue(typ, nil), nil
	}

	switch {
	case typ.Is(String):
		return jsonUnmarshalString(buf, typ, p)
	case typ.Is(Number):
		return jsonUnmarshalNumber(buf, typ, p)
	case typ.Is(Bool):
		return jsonUnmarshalBool(buf, typ, p)
	case typ.Is(DynamicPseudoType):
		return jsonUnmarshalDynamicPseudoType(buf, typ, p, opts)
	case typ.Is(List{}):
		//nolint:forcetypeassert // Is func above guarantees this type assertion
		return jsonUnmarshalList(buf, typ.(List).ElementType, p, opts)
	case typ.Is(Set{}):
		//nolint:forcetypeassert // Is func above guarantees this type assertion
		return jsonUnmarshalSet(buf, typ.(Set).ElementType, p, opts)
	case typ.Is(Map{}):
		//nolint:forcetypeassert // Is func above guarantees this type assertion
		return jsonUnmarshalMap(buf, typ.(Map).ElementType, p, opts)
	case typ.Is(Tuple{}):
		//nolint:forcetypeassert // Is func above guarantees this type assertion
		return jsonUnmarshalTuple(buf, typ.(Tuple).ElementTypes, p, opts)
	case typ.Is(Object{}):
		//nolint:forcetypeassert // Is func above guarantees this type assertion
		return jsonUnmarshalObject(buf, typ.(Object).AttributeTypes, p, opts)
	}
	return Value{}, p.NewErrorf("unknown type %s", typ)
}

func jsonUnmarshalString(buf []byte, _ Type, p *AttributePath) (Value, error) {
	dec := jsonByteDecoder(buf)

	tok, err := dec.Token()
	if err != nil {
		return Value{}, p.NewErrorf("error reading token: %w", err)
	}
	switch v := tok.(type) {
	case string:
		return NewValue(String, v), nil
	case json.Number:
		return NewValue(String, string(v)), nil
	case bool:
		if v {
			return NewValue(String, "true"), nil
		}
		return NewValue(String, "false"), nil
	}
	return Value{}, p.NewErrorf("unsupported type %T sent as %s", tok, String)
}

func jsonUnmarshalNumber(buf []byte, typ Type, p *AttributePath) (Value, error) {
	dec := jsonByteDecoder(buf)

	tok, err := dec.Token()
	if err != nil {
		return Value{}, p.NewErrorf("error reading token: %w", err)
	}
	switch numTok := tok.(type) {
	case json.Number:
		f, _, err := big.ParseFloat(string(numTok), 10, 512, big.ToNearestEven)
		if err != nil {
			return Value{}, p.NewErrorf("error parsing number: %w", err)
		}
		return NewValue(typ, f), nil
	case string:
		f, _, err := big.ParseFloat(numTok, 10, 512, big.ToNearestEven)
		if err != nil {
			return Value{}, p.NewErrorf("error parsing number: %w", err)
		}
		return NewValue(typ, f), nil
	}
	return Value{}, p.NewErrorf("unsupported type %T sent as %s", tok, Number)
}

func jsonUnmarshalBool(buf []byte, _ Type, p *AttributePath) (Value, error) {
	dec := jsonByteDecoder(buf)
	tok, err := dec.Token()
	if err != nil {
		return Value{}, p.NewErrorf("error reading token: %w", err)
	}
	switch v := tok.(type) {
	case bool:
		return NewValue(Bool, v), nil
	case string:
		switch v {
		case "true", "1":
			return NewValue(Bool, true), nil
		case "false", "0":
			return NewValue(Bool, false), nil
		}
		switch strings.ToLower(v) {
		case "true":
			return Value{}, p.NewErrorf("to convert from string, use lowercase \"true\"")
		case "false":
			return Value{}, p.NewErrorf("to convert from string, use lowercase \"false\"")
		}
	case json.Number:
		switch v {
		case "1":
			return NewValue(Bool, true), nil
		case "0":
			return NewValue(Bool, false), nil
		}
	}
	return Value{}, p.NewErrorf("unsupported type %T sent as %s", tok, Bool)
}

func jsonUnmarshalDynamicPseudoType(buf []byte, _ Type, p *AttributePath, opts ValueFromJSONOpts) (Value, error) {
	dec := jsonByteDecoder(buf)
	tok, err := dec.Token()
	if err != nil {
		return Value{}, p.NewErrorf("error reading token: %w", err)
	}
	if tok != json.Delim('{') {
		return Value{}, p.NewErrorf("invalid JSON, expected %q, got %q", json.Delim('{'), tok)
	}
	var t Type
	var valBody []byte
	for dec.More() {
		tok, err = dec.Token()
		if err != nil {
			return Value{}, p.NewErrorf("error reading token: %w", err)
		}
		key, ok := tok.(string)
		if !ok {
			return Value{}, p.NewErrorf("expected key to be a string, got %T", tok)
		}
		var rawVal json.RawMessage
		err = dec.Decode(&rawVal)
		if err != nil {
			return Value{}, p.NewErrorf("error decoding value: %w", err)
		}
		switch key {
		case "type":
			t, err = ParseJSONType(rawVal) //nolint:staticcheck
			if err != nil {
				return Value{}, p.NewErrorf("error decoding type information: %w", err)
			}
		case "value":
			valBody = rawVal
		default:
			return Value{}, p.NewErrorf("invalid key %q in dynamically-typed value", key)
		}
	}
	tok, err = dec.Token()
	if err != nil {
		return Value{}, p.NewErrorf("error reading token: %w", err)
	}
	if tok != json.Delim('}') {
		return Value{}, p.NewErrorf("invalid JSON, expected %q, got %q", json.Delim('}'), tok)
	}
	if t == nil {
		return Value{}, p.NewErrorf("missing type in dynamically-typed value")
	}
	if valBody == nil {
		return Value{}, p.NewErrorf("missing value in dynamically-typed value")
	}
	return jsonUnmarshal(valBody, t, p, opts)
}

func jsonUnmarshalList(buf []byte, elementType Type, p *AttributePath, opts ValueFromJSONOpts) (Value, error) {
	dec := jsonByteDecoder(buf)

	tok, err := dec.Token()
	if err != nil {
		return Value{}, p.NewErrorf("error reading token: %w", err)
	}
	if tok != json.Delim('[') {
		return Value{}, p.NewErrorf("invalid JSON, expected %q, got %q", json.Delim('['), tok)
	}

	// we want to have a value for this always, even if there are no
	// elements, because no elements is *technically* different than empty,
	// and we want to preserve that distinction
	//
	// var vals []Value
	// would evaluate as nil if the list is empty
	//
	// while generally in Go it's undesirable to treat empty and nil slices
	// separately, in this case we're surfacing a non-Go-in-origin
	// distinction, so we'll allow it.
	vals := []Value{}

	var idx int
	for dec.More() {
		innerPath := p.WithElementKeyInt(idx)
		// update the index
		idx++

		var rawVal json.RawMessage
		err = dec.Decode(&rawVal)
		if err != nil {
			return Value{}, innerPath.NewErrorf("error decoding value: %w", err)
		}
		val, err := jsonUnmarshal(rawVal, elementType, innerPath, opts)
		if err != nil {
			return Value{}, err
		}
		vals = append(vals, val)
	}

	tok, err = dec.Token()
	if err != nil {
		return Value{}, p.NewErrorf("error reading token: %w", err)
	}
	if tok != json.Delim(']') {
		return Value{}, p.NewErrorf("invalid JSON, expected %q, got %q", json.Delim(']'), tok)
	}

	elTyp := elementType
	if elTyp.Is(DynamicPseudoType) {
		elTyp, err = TypeFromElements(vals)
		if err != nil {
			return Value{}, p.NewErrorf("invalid elements for list: %w", err)
		}
	}
	return NewValue(List{
		ElementType: elTyp,
	}, vals), nil
}

func jsonUnmarshalSet(buf []byte, elementType Type, p *AttributePath, opts ValueFromJSONOpts) (Value, error) {
	dec := jsonByteDecoder(buf)

	tok, err := dec.Token()
	if err != nil {
		return Value{}, p.NewErrorf("error reading token: %w", err)
	}
	if tok != json.Delim('[') {
		return Value{}, p.NewErrorf("invalid JSON, expected %q, got %q", json.Delim('['), tok)
	}

	// we want to have a value for this always, even if there are no
	// elements, because no elements is *technically* different than empty,
	// and we want to preserve that distinction
	//
	// var vals []Value
	// would evaluate as nil if the set is empty
	//
	// while generally in Go it's undesirable to treat empty and nil slices
	// separately, in this case we're surfacing a non-Go-in-origin
	// distinction, so we'll allow it.
	vals := []Value{}

	for dec.More() {
		innerPath := p.WithElementKeyValue(NewValue(elementType, UnknownValue))
		var rawVal json.RawMessage
		err = dec.Decode(&rawVal)
		if err != nil {
			return Value{}, innerPath.NewErrorf("error decoding value: %w", err)
		}
		val, err := jsonUnmarshal(rawVal, elementType, innerPath, opts)
		if err != nil {
			return Value{}, err
		}
		vals = append(vals, val)
	}
	tok, err = dec.Token()
	if err != nil {
		return Value{}, p.NewErrorf("error reading token: %w", err)
	}
	if tok != json.Delim(']') {
		return Value{}, p.NewErrorf("invalid JSON, expected %q, got %q", json.Delim(']'), tok)
	}

	elTyp := elementType
	if elTyp.Is(DynamicPseudoType) {
		elTyp, err = TypeFromElements(vals)
		if err != nil {
			return Value{}, p.NewErrorf("invalid elements for list: %w", err)
		}
	}
	return NewValue(Set{
		ElementType: elTyp,
	}, vals), nil
}

func jsonUnmarshalMap(buf []byte, attrType Type, p *AttributePath, opts ValueFromJSONOpts) (Value, error) {
	dec := jsonByteDecoder(buf)

	tok, err := dec.Token()
	if err != nil {
		return Value{}, p.NewErrorf("error reading token: %w", err)
	}
	if tok != json.Delim('{') {
		return Value{}, p.NewErrorf("invalid JSON, expected %q, got %q", json.Delim('{'), tok)
	}

	vals := map[string]Value{}
	for dec.More() {
		innerPath := p.WithElementKeyValue(NewValue(attrType, UnknownValue))
		tok, err := dec.Token()
		if err != nil {
			return Value{}, innerPath.NewErrorf("error reading token: %w", err)
		}
		key, ok := tok.(string)
		if !ok {
			return Value{}, innerPath.NewErrorf("expected map key to be a string, got %T", tok)
		}

		//fix the path value, we have an actual key now
		innerPath = p.WithElementKeyString(key)

		var rawVal json.RawMessage
		err = dec.Decode(&rawVal)
		if err != nil {
			return Value{}, innerPath.NewErrorf("error decoding value: %w", err)
		}
		val, err := jsonUnmarshal(rawVal, attrType, innerPath, opts)
		if err != nil {
			return Value{}, err
		}
		vals[key] = val
	}
	tok, err = dec.Token()
	if err != nil {
		return Value{}, p.NewErrorf("error reading token: %w", err)
	}
	if tok != json.Delim('}') {
		return Value{}, p.NewErrorf("invalid JSON, expected %q, got %q", json.Delim('}'), tok)
	}

	return NewValue(Map{
		ElementType: attrType,
	}, vals), nil
}

func jsonUnmarshalTuple(buf []byte, elementTypes []Type, p *AttributePath, opts ValueFromJSONOpts) (Value, error) {
	dec := jsonByteDecoder(buf)

	tok, err := dec.Token()
	if err != nil {
		return Value{}, p.NewErrorf("error reading token: %w", err)
	}
	if tok != json.Delim('[') {
		return Value{}, p.NewErrorf("invalid JSON, expected %q, got %q", json.Delim('['), tok)
	}

	// we want to have a value for this always, even if there are no
	// elements, because no elements is *technically* different than empty,
	// and we want to preserve that distinction
	//
	// var vals []Value
	// would evaluate as nil if the tuple is empty
	//
	// while generally in Go it's undesirable to treat empty and nil slices
	// separately, in this case we're surfacing a non-Go-in-origin
	// distinction, so we'll allow it.
	vals := []Value{}

	var idx int
	for dec.More() {
		if idx >= len(elementTypes) {
			return Value{}, p.NewErrorf("too many tuple elements (only have types for %d)", len(elementTypes))
		}

		innerPath := p.WithElementKeyInt(idx)
		elementType := elementTypes[idx]
		idx++

		var rawVal json.RawMessage
		err = dec.Decode(&rawVal)
		if err != nil {
			return Value{}, innerPath.NewErrorf("error decoding value: %w", err)
		}
		val, err := jsonUnmarshal(rawVal, elementType, innerPath, opts)
		if err != nil {
			return Value{}, err
		}
		vals = append(vals, val)
	}

	tok, err = dec.Token()
	if err != nil {
		return Value{}, p.NewErrorf("error reading token: %w", err)
	}
	if tok != json.Delim(']') {
		return Value{}, p.NewErrorf("invalid JSON, expected %q, got %q", json.Delim(']'), tok)
	}

	if len(vals) != len(elementTypes) {
		return Value{}, p.NewErrorf("not enough tuple elements (only have %d, have types for %d)", len(vals), len(elementTypes))
	}

	return NewValue(Tuple{
		ElementTypes: elementTypes,
	}, vals), nil
}

// jsonUnmarshalObject attempts to decode JSON object structure to tftypes.Value object.
// opts contains fields that can be used to modify the behaviour of JSON unmarshalling.
func jsonUnmarshalObject(buf []byte, attrTypes map[string]Type, p *AttributePath, opts ValueFromJSONOpts) (Value, error) {
	dec := jsonByteDecoder(buf)

	tok, err := dec.Token()
	if err != nil {
		return Value{}, p.NewErrorf("error reading token: %w", err)
	}
	if tok != json.Delim('{') {
		return Value{}, p.NewErrorf("invalid JSON, expected %q, got %q", json.Delim('{'), tok)
	}

	vals := map[string]Value{}
	for dec.More() {
		tok, err := dec.Token()
		if err != nil {
			return Value{}, p.NewErrorf("error reading object attribute key token: %w", err)
		}
		key, ok := tok.(string)
		if !ok {
			return Value{}, p.NewErrorf("object attribute key was %T with value %v, not string", tok, tok)
		}
		innerPath := p.WithAttributeName(key)
		attrType, ok := attrTypes[key]
		if !ok {
			if opts.IgnoreUndefinedAttributes {
				// We are trying to ignore the key and value of any unsupported attribute.
				_ = dec.Decode(new(json.RawMessage))
				continue
			}

			return Value{}, innerPath.NewErrorf("unsupported attribute %q", key)
		}

		var rawVal json.RawMessage
		err = dec.Decode(&rawVal)
		if err != nil {
			return Value{}, innerPath.NewErrorf("error decoding value: %w", err)
		}
		val, err := jsonUnmarshal(rawVal, attrType, innerPath, opts)
		if err != nil {
			return Value{}, err
		}
		vals[key] = val
	}

	tok, err = dec.Token()
	if err != nil {
		return Value{}, p.NewErrorf("error reading token: %w", err)
	}
	if tok != json.Delim('}') {
		return Value{}, p.NewErrorf("invalid JSON, expected %q, got %q", json.Delim('}'), tok)
	}

	// make sure we have a value for every attribute
	for k, typ := range attrTypes {
		if _, ok := vals[k]; !ok {
			vals[k] = NewValue(typ, nil)
		}
	}

	return NewValue(Object{
		AttributeTypes: attrTypes,
	}, vals), nil
}
