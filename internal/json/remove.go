// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package json

import (
	"bytes"
	"encoding/json/jsontext"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
)

// RemoveOptions configures values removed by RemoveJSONValues.
type RemoveOptions struct {
	// Fields contains decoded object field names to remove at every nesting level.
	Fields []string
	// Null removes null values from objects and arrays.
	Null bool
	// EmptyArrays removes arrays that are empty after their elements are transformed.
	EmptyArrays bool
	// EmptyObjects removes objects that are empty after their fields are transformed.
	EmptyObjects bool
	// EmptyStringFields removes empty strings from object fields, but not from arrays.
	EmptyStringFields bool
}

// RemoveJSONValues removes selected values from a JSON document and compacts the result.
// The top-level value is always retained.
func RemoveJSONValues(in string, options RemoveOptions) (string, error) {
	decoder := jsontext.NewDecoder(
		strings.NewReader(in),
		jsontext.AllowDuplicateNames(true),
		jsontext.AllowInvalidUTF8(true),
	)
	out := make([]byte, 0, len(in))

	if _, err := transformJSONValue(decoder, &out, options, false, true); err != nil {
		return "", err
	}

	if _, err := decoder.ReadToken(); !errors.Is(err, io.EOF) {
		if err == nil {
			return "", fmt.Errorf("multiple top-level JSON values")
		}
		return "", err
	}

	return string(out), nil
}

func transformJSONValue(decoder *jsontext.Decoder, out *[]byte, options RemoveOptions, objectField, root bool) (bool, error) {
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
			keep = root || !((options.Null && value.Kind() == jsontext.KindNull) ||
				(options.EmptyStringFields && objectField && bytes.Equal(value, []byte(`""`))))
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

func transformJSONObject(decoder *jsontext.Decoder, out *[]byte, options RemoveOptions, root bool) (bool, error) {
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

		if len(options.Fields) != 0 {
			decodedName, err := jsontext.AppendUnquote(nil, name)
			// The decoder has already validated the string syntax. Unquoting can
			// only fail here for invalid UTF-8, which is preserved for backwards
			// compatibility and cannot match a decoded field name.
			if err == nil && slices.Contains(options.Fields, string(decodedName)) {
				if err := decoder.SkipValue(); err != nil {
					return false, err
				}
				continue
			}
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

	return root || members != 0 || !options.EmptyObjects, nil
}

func transformJSONArray(decoder *jsontext.Decoder, out *[]byte, options RemoveOptions, root bool) (bool, error) {
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

	return root || elements != 0 || !options.EmptyArrays, nil
}
