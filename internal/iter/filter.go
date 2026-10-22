// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iter

import (
	"iter"
)

// Predicate represents a predicate (boolean-valued function) of one argument.
type Predicate[T any] func(T) bool

// Filter returns an iterator over the filtered elements of the sequence.
func Filter[T any](seq iter.Seq[T], pred Predicate[T]) iter.Seq[T] {
	return func(yield func(T) bool) {
		for e := range seq {
			if pred(e) && !yield(e) {
				return
			}
		}
	}
}

// FilterValues returns an iterator over seq without the values pred rejects.
// Pairs carrying an error always pass, so a filter cannot quietly turn a failed
// listing into a short one.
func FilterValues[T any](seq iter.Seq2[T, error], pred Predicate[T]) iter.Seq2[T, error] {
	return func(yield func(T, error) bool) {
		for v, err := range seq {
			if err != nil || pred(v) {
				if !yield(v, err) {
					return
				}
			}
		}
	}
}
