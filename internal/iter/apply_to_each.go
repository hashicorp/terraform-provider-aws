// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iter

import (
	"iter"
)

// Filtered returns an iterator over the filtered elements of the sequence.
func AppliedToEach[E, T any](seq iter.Seq[E], f func(E) T) iter.Seq[T] {
	return func(yield func(T) bool) {
		for v := range seq {
			if !yield(f(v)) {
				return
			}
		}
	}
}
