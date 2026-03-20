// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package slices

import (
	"iter"
)

// BackwardValues returns an iterator that yields the slice elements in reverse order.
// It is a values-only equivalent of `slices.Backward`.
func BackwardValues[Slice ~[]E, E any](s Slice) iter.Seq[E] {
	return func(yield func(E) bool) {
		for i := len(s) - 1; i >= 0; i-- {
			if !yield(s[i]) {
				return
			}
		}
	}
}
