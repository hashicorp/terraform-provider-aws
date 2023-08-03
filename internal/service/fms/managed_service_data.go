// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fms

import (
	"github.com/hashicorp/terraform-provider-aws/internal/service/fms/ujson"
)

// removeEmptyFieldsFromJSON removes `null` and empty array (`[]`) fields from a valid JSON string.
func removeEmptyFieldsFromJSON(in string) string {
	out := make([]byte, 0, len(in))
	lenBeforeArray := 0

	err := ujson.Walk([]byte(in), func(_ int, key, value []byte) bool {
		n := len(out)

		// For valid JSON, value will never be empty.
		skip := false
		switch value[0] {
		case 'n': // Null (null)
			skip = true
		case '[': // Start of array
			lenBeforeArray = n
		case ']': // End of array
			if out[n-1] == '[' {
				// Truncate output.
				out = out[:lenBeforeArray]
				lenBeforeArray = 0
				skip = true
			}
		}

		if skip {
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
		return ""
	}

	return string(out)
}
