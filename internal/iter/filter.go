// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iter

import (
	"iter"
)

// Predicate represents a predicate (boolean-valued function) of one argument.
type Predicate[T any] func(T) bool

// Filtered returns an iterator over the filtered elements of the sequence.
func Filtered[T any](seq iter.Seq[T], pred Predicate[T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		for e := range seq {
			if pred(e) && !yield(e) {
				return
			}
		}
	}
}
