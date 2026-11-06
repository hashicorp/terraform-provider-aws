// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package json

import (
	"bytes"
	"encoding/json/jsontext"
	"errors"
	"fmt"
	"io"
)

type removeOptions struct {
	fields            map[string]struct{}
	null              bool
	emptyArrays       bool
	emptyObjects      bool
	emptyStringFields bool
}

// RemoveFields removes the specified fields from a valid JSON string.
func RemoveFields(in string, fields ...string) string {
	removeFields := make(map[string]struct{}, len(fields))
	for _, field := range fields {
		removeFields[field] = struct{}{}
	}

	out, err := removeJSONValues([]byte(in), removeOptions{fields: removeFields})
	if err != nil {
		return ""
	}

	return string(out)
}

// RemoveEmptyFields removes null, empty array, and empty object values from a valid JSON document.
func RemoveEmptyFields(in []byte) []byte {
	out, err := removeJSONValues(in, removeOptions{
		null:         true,
		emptyArrays:  true,
		emptyObjects: true,
	})
	if err != nil {
		return nil
	}

	return out
}

// RemoveEmptyStringFields removes empty string values from object fields in a valid JSON document.
func RemoveEmptyStringFields(in []byte) []byte {
	out, err := removeJSONValues(in, removeOptions{emptyStringFields: true})
	if err != nil {
		return nil
	}

	return out
}

// RemoveNullAndEmptyArrayValues removes null and empty array values from a valid JSON document.
func RemoveNullAndEmptyArrayValues(in []byte) []byte {
	out, err := removeJSONValues(in, removeOptions{
		null:        true,
		emptyArrays: true,
	})
	if err != nil {
		return nil
	}

	return out
}

func removeJSONValues(in []byte, options removeOptions) ([]byte, error) {
	decoder := jsontext.NewDecoder(
		bytes.NewReader(in),
		jsontext.AllowDuplicateNames(true),
		jsontext.AllowInvalidUTF8(true),
	)
	out := make([]byte, 0, len(in))

	if _, err := transformJSONValue(decoder, &out, options, false, true); err != nil {
		return nil, err
	}

	if _, err := decoder.ReadToken(); !errors.Is(err, io.EOF) {
		if err == nil {
			return nil, fmt.Errorf("multiple top-level JSON values")
		}
		return nil, err
	}

	return out, nil
}

func transformJSONValue(decoder *jsontext.Decoder, out *[]byte, options removeOptions, objectField, root bool) (bool, error) {
	start := len(*out)

	var (
		keep bool
		err  error
	)
	switch decoder.PeekKind() {
	case jsontext.KindBeginObject:
		keep, err = transformJSONObject(decoder, out, options, root)
	case jsontext.KindBeginArray:
		keep, err = transformJSONArray(decoder, out, options, root)
	default:
		var value jsontext.Value
		value, err = decoder.ReadValue()
		if err == nil {
			keep = root || !((options.null && value.Kind() == jsontext.KindNull) ||
				(options.emptyStringFields && objectField && bytes.Equal(value, []byte(`""`))))
			if keep {
				*out = append(*out, value...)
			}
		}
	}
	if err != nil {
		return false, err
	}
	if !keep {
		*out = (*out)[:start]
	}

	return keep, nil
}

func transformJSONObject(decoder *jsontext.Decoder, out *[]byte, options removeOptions, root bool) (bool, error) {
	if _, err := decoder.ReadToken(); err != nil {
		return false, err
	}
	*out = append(*out, '{')

	members := 0
	for decoder.PeekKind() != jsontext.KindEndObject {
		name, err := decoder.ReadValue()
		if err != nil {
			return false, err
		}

		if _, ok := options.fields[string(name)]; ok {
			if err := decoder.SkipValue(); err != nil {
				return false, err
			}
			continue
		}

		start := len(*out)
		if members != 0 {
			*out = append(*out, ',')
		}
		*out = append(*out, name...)
		*out = append(*out, ':')

		keep, err := transformJSONValue(decoder, out, options, true, false)
		if err != nil {
			return false, err
		}
		if !keep {
			*out = (*out)[:start]
			continue
		}
		members++
	}

	if _, err := decoder.ReadToken(); err != nil {
		return false, err
	}
	*out = append(*out, '}')

	return root || members != 0 || !options.emptyObjects, nil
}

func transformJSONArray(decoder *jsontext.Decoder, out *[]byte, options removeOptions, root bool) (bool, error) {
	if _, err := decoder.ReadToken(); err != nil {
		return false, err
	}
	*out = append(*out, '[')

	elements := 0
	for decoder.PeekKind() != jsontext.KindEndArray {
		start := len(*out)
		if elements != 0 {
			*out = append(*out, ',')
		}

		keep, err := transformJSONValue(decoder, out, options, false, false)
		if err != nil {
			return false, err
		}
		if !keep {
			*out = (*out)[:start]
			continue
		}
		elements++
	}

	if _, err := decoder.ReadToken(); err != nil {
		return false, err
	}
	*out = append(*out, ']')

	return root || elements != 0 || !options.emptyArrays, nil
}
