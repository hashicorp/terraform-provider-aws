// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package json

import (
	evanphxjsonpatch "github.com/evanphx/json-patch"
	mattbairdjsonpatch "github.com/mattbaird/jsonpatch"
)

// `CreatePatchFromStrings` creates an [RFC6902](https://datatracker.ietf.org/doc/html/rfc6902) JSON Patch from two JSON strings.
// `a` is the original JSON document and `b` is the modified JSON document.
// The patch is returned as an array of operations (which can be encoded to JSON).
func CreatePatchFromStrings(a, b string) ([]mattbairdjsonpatch.JsonPatchOperation, error) {
	return mattbairdjsonpatch.CreatePatch([]byte(a), []byte(b))
}

// `CreateMergePatchFromStrings` creates an [RFC7396](https://datatracker.ietf.org/doc/html/rfc7396) JSON merge patch from two JSON strings.
// `a` is the original JSON document and `b` is the modified JSON document.
// The patch is returned as a JSON string.
func CreateMergePatchFromStrings(a, b string) (string, error) {
	patch, err := evanphxjsonpatch.CreateMergePatch([]byte(a), []byte(b))
	if err != nil {
		return "", err
	}

	return string(patch), nil
}
