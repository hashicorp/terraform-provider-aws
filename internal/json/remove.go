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
