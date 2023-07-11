// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package slices

import "golang.org/x/exp/slices"

// Reverse returns a reversed copy of the slice.
func Reverse[S ~[]E, E any](s S) S {
	v := S([]E{})
	n := len(s)

	for i := 0; i < n; i++ {
		v = append(v, s[n-(i+1)])
	}

	return v
}

// RemoveAll removes all occurrences of the specified value from a slice.
func RemoveAll[E comparable](s []E, r E) []E {
	v := []E{}

	for _, e := range s {
		if e != r {
			v = append(v, e)
		}
	}

	return v
}

// ApplyToAll returns a new slice containing the results of applying the function `f` to each element of the original slice `s`.
func ApplyToAll[T, U any](s []T, f func(T) U) []U {
	v := make([]U, len(s))

	for i, e := range s {
		v[i] = f(e)
	}

	return v
}

type FilterFunc[T any] func(T) bool

// Filter returns a new slice containing all values that return `true` for the filter function `f`
func Filter[T any](s []T, f FilterFunc[T]) []T {
	v := make([]T, 0, len(s))

	for _, e := range s {
		if f(e) {
			v = append(v, e)
		}
	}

	return slices.Clip(v)
}

// All returns `true` if the filter function `f` retruns `true` for all items
func All[T any](s []T, f FilterFunc[T]) bool {
	for _, e := range s {
		if !f(e) {
			return false
		}
	}
	return true
}

// Any returns `true` if the filter function `f` retruns `true` for any item
func Any[T any](s []T, f FilterFunc[T]) bool {
	for _, e := range s {
		if f(e) {
			return true
		}
	}
	return false
}

// Chunks returns a slice of S, each of the specified size (or less).
func Chunks[S ~[]E, E any](s S, size int) []S {
	chunks := make([]S, 0)

	for i := 0; i < len(s); i += size {
		end := i + size

		if end > len(s) {
			end = len(s)
		}

		chunks = append(chunks, s[i:end])
	}

	return chunks
}
