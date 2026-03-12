// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package json

import (
	"iter"

	"github.com/hashicorp/terraform-provider-aws/internal/json/ujson"
)

// TopLevelKeys returns the top-level keys for a valid JSON string.
func TopLevelKeys(in string) iter.Seq[string] {
	return func(yield func(string) bool) {
		ujson.Walk([]byte(in), func(level int, key, _ []byte) bool {
			if level == 1 {
				yield(string(key))
				return false
			}
			return true
		})
	}
}
