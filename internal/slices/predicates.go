// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package slices

// PredicateAnd returns a Predicate that evaluates to true if all of the specified predicates evaluate to true.
func PredicateAnd[T any](predicates ...Predicate[T]) Predicate[T] {
	return func(v T) bool {
		for _, predicate := range predicates {
			if !predicate(v) {
				return false
			}
		}

		return true
	}
}

// PredicateOr returns a Predicate that evaluates to true if any of the specified predicates evaluate to true.
func PredicateOr[T any](predicates ...Predicate[T]) Predicate[T] {
	return func(v T) bool {
		for _, predicate := range predicates {
			if predicate(v) {
				return true
			}
		}

		return false
	}
}

// PredicateEquals returns a Predicate that evaluates to true if the predicate's argument equals `v`.
func PredicateEquals[T comparable](v T) Predicate[T] {
	return func(x T) bool {
		return x == v
	}
}

// PredicateTrue returns a Predicate that always evaluates to true.
func PredicateTrue[T any]() Predicate[T] {
	return func(T) bool {
		return true
	}
}

func PredicateValue[T any](predicate Predicate[*T]) Predicate[T] {
	return func(v T) bool {
		return predicate(&v)
	}
}
