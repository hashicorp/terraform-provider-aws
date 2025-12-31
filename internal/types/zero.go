// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"reflect"
)

// IsZero returns true if `v` is the zero value of `T`, `nil`, is a pointer to `nil`, or points to the zero value of `T`.
func IsZero[T any](v T) bool {
	val := reflect.ValueOf(v)
	val = reflect.Indirect(val)

	if val.Kind() == reflect.Interface {
		val = val.Elem()
	}

	if !val.IsValid() || val.IsZero() {
		return true
	}

	return false
}

// Zero returns the zero value for T.
func Zero[T any]() T {
	var z T
	return z
}
