// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iter

import (
	"iter"
)

// Null returns an empty iterator.
func Null[E any]() iter.Seq[E] {
	return func(yield func(E) bool) {}
}
