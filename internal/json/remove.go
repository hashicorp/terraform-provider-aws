// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package json

import (
	"bytes"

	"github.com/hashicorp/terraform-provider-aws/internal/json/ujson"
	"github.com/hashicorp/terraform-provider-aws/internal/types/stack"
)

// RemoveFields removes the specified fields from a valid JSON string.
func RemoveFields(in string, fields ...string) string {
	out := make([]byte, 0, len(in))

	err := ujson.Walk([]byte(in), func(_ int, key, value []byte) bool {
		if len(key) != 0 {
			for _, field := range fields {
				if bytes.Equal(key, []byte(field)) {
					// Remove the key and value from the output.
					return false
				}
			}
		}

		// Write to output.
		if len(out) != 0 && ujson.ShouldAddComma(value, out[len(out)-1]) {
			out = append(out, ',')
		}
		if len(key) > 0 {
			out = append(out, key...)
			out = append(out, ':')
		}
		out = append(out, value...)

		return true
	})

	if err != nil {
		return ""
	}

	return string(out)
}

// RemoveEmptyFields removes all empty fields from a valid JSON string.
func RemoveEmptyFields(in []byte) []byte {
	var n int
	for {
		in, n = removeEmptyFields(in)
		if n == 0 {
			break
		}
	}

	return in
}

// removeEmptyFields removes `null`, empty array (`[]`) and empty object (`{}`) fields from a valid JSON string.
// Returns the new JSON string and the number of empty fields removed.
func removeEmptyFields(in []byte) ([]byte, int) {
	out := make([]byte, 0, len(in))
	before := stack.New[int]()
	removed := 0

	err := ujson.Walk(in, func(_ int, key, value []byte) bool {
		n := len(out)

		// For valid JSON, value will never be empty.
		skip := false
		switch value[0] {
		case 'n': // Null (null)
			skip = true
		case '[': // Start of array
			before.Push(n)
		case ']': // End of array
			i := before.Pop().MustUnwrap()
			if out[n-1] == '[' {
				// Truncate output.
				out = out[:i]
				skip = true
			}
		case '{': // Start of object
			before.Push(n)
		case '}': // End of object
			i := before.Pop().MustUnwrap()
			if n > 1 && out[n-1] == '{' {
				// Truncate output.
				out = out[:i]
				skip = true
			}
		}

		if skip {
			removed++
			return false
		}

		if n != 0 && ujson.ShouldAddComma(value, out[n-1]) {
			out = append(out, ',')
		}
		if len(key) > 0 {
			out = append(out, key...)
			out = append(out, ':')
		}
		out = append(out, value...)

		return true
	})

	if err != nil {
		return nil, 0
	}

	return out, removed
}
