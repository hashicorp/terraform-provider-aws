// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iter

import (
	"iter"
)

// Null returns an empty iterator.
func Null[V any]() iter.Seq[V] {
	return func(yield func(V) bool) {}
}

// Null2 returns an empty value pair iterator.
func Null2[K, V any]() iter.Seq2[K, V] {
	return func(yield func(K, V) bool) {}
}
