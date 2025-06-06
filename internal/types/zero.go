// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"reflect"
)

// IsZero returns true if `v` is `nil`, is a pointer to `nil`, or points to the zero value of `T`.
func IsZero[T any](v *T) bool {
	if v == nil {
		return true
	}

	if val := reflect.ValueOf(*v); !val.IsValid() || val.IsZero() {
		return true
	}

	return false
}
