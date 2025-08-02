// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package slices

import (
	"iter"
)

// AppliedToEach returns an iterator that yields the slice elements transformed by the function `f`.
func AppliedToEach[S ~[]E, E any, T any](s S, f func(E) T) iter.Seq[T] {
	return func(yield func(T) bool) {
		for _, v := range s {
			if !yield(f(v)) {
				return
			}
		}
	}
}
