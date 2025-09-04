// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package json

import (
	mattbairdjsonpatch "github.com/mattbaird/jsonpatch"
)

// `CreatePatchFromStrings` creates an [RFC6902](https://datatracker.ietf.org/doc/html/rfc6902) JSON Patch from two JSON strings.
// `a` is the original JSON document and `b` is the modified JSON document.
// The patch is returned as an array of operations (which can be encoded to JSON).
func CreatePatchFromStrings(a, b string) ([]mattbairdjsonpatch.JsonPatchOperation, error) {
	return mattbairdjsonpatch.CreatePatch([]byte(a), []byte(b))
}
