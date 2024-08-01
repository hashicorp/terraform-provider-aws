// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package json

import (
	"unicode"
	"unicode/utf8"

	"github.com/hashicorp/terraform-provider-aws/internal/json/ujson"
)

// KeyFirstLower converts the first letter of each key to lowercase.
func KeyFirstLower(in []byte) []byte {
	out := make([]byte, 0, len(in))

	err := ujson.Walk(in, func(_ int, key, value []byte) bool {
		// Write to output.
		if len(out) != 0 && ujson.ShouldAddComma(value, out[len(out)-1]) {
			out = append(out, ',')
		}
		// key is the raw key of the current object or empty otherwise.
		// It can be a double-quoted string or empty.
		switch len(key) {
		case 0:
		case 1, 2:
			// Empty key.
			out = append(out, key...)
		default:
		}
		if len(key) > 0 {
			out = append(out, '"')
			r, n := utf8.DecodeRune(key[1:])
			r = unicode.ToLower(r)
			low := make([]byte, utf8.RuneLen(r))
			utf8.EncodeRune(low, r)
			out = append(out, low...)
			out = append(out, key[n+1:]...)
			out = append(out, ':')
		}
		out = append(out, value...)

		return true
	})

	if err != nil {
		return nil
	}

	return out
}
