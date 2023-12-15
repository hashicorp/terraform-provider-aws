// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package json

import (
	"bytes"

	"github.com/hashicorp/terraform-provider-aws/internal/json/ujson"
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
func RemoveEmptyFields(in string) string {
	n := 0
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
func removeEmptyFields(in string) (string, int) {
	out := make([]byte, 0, len(in))
	lenBefore := 0
	removed := 0

	err := ujson.Walk([]byte(in), func(_ int, key, value []byte) bool {
		n := len(out)

		// For valid JSON, value will never be empty.
		skip := false
		switch value[0] {
		case 'n': // Null (null)
			skip = true
		case '[': // Start of array
			lenBefore = n
		case ']': // End of array
			if out[n-1] == '[' {
				// Truncate output.
				out = out[:lenBefore]
				lenBefore = 0
				skip = true
			}
		case '{': // Start of object
			lenBefore = n
		case '}': // End of object
			if n > 1 && out[n-1] == '{' {
				// Truncate output.
				out = out[:lenBefore]
				lenBefore = 0
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
		return "", 0
	}

	return string(out), removed
}
