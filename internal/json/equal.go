// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package json

import (
	"reflect"
)

// EqualBytes returns whether the JSON documents in the given byte slices are equal.
func EqualBytes(b1, b2 []byte) bool {
	var v1 any
	if err := DecodeFromBytes(b1, &v1); err != nil {
		return false
	}

	var v2 any
	if err := DecodeFromBytes(b2, &v2); err != nil {
		return false
	}

	return reflect.DeepEqual(v1, v2)
}

// EqualStrings returns whether the JSON documents in the given strings are equal.
func EqualStrings(s1, s2 string) bool {
	return EqualBytes([]byte(s1), []byte(s2))
}
