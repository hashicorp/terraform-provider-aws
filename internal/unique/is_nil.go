// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package unique

import (
	"unique"
)

// IsHandleNil checks whether a Handle has had a value assigned.
func IsHandleNil[T comparable](h unique.Handle[T]) bool {
	return isZero(h)
}

func isZero[T comparable](v T) bool {
	var zero T
	return v == zero
}
