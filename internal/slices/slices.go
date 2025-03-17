// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package slices

import (
	"slices"
)

// Reverse returns a reversed copy of the slice `s`.
func Reverse[S ~[]E, E any](s S) S {
	n := len(s)
	v := S(make([]E, 0, n))

	for i := range n {
		v = append(v, s[n-(i+1)])
	}

	return v
}

// RemoveAll removes all occurrences of the specified value `r` from a slice `s`.
func RemoveAll[S ~[]E, E comparable](s S, vs ...E) S {
	v := S(make([]E, 0, len(s)))

	for _, e := range s {
		if !slices.Contains(vs, e) {
			v = append(v, e)
		}
	}

	return slices.Clip(v)
}

// ApplyToAll returns a new slice containing the results of applying the function `f` to each element of the original slice `s`.
func ApplyToAll[S ~[]E1, E1, E2 any](s S, f func(E1) E2) []E2 {
	v := make([]E2, len(s))

	for i, e := range s {
		v[i] = f(e)
	}

	return v
}

func ApplyToAllWithError[S ~[]E1, E1, E2 any](s S, f func(E1) (E2, error)) ([]E2, error) {
	v := make([]E2, len(s))

	for i, e1 := range s {
		e2, err := f(e1)
		if err != nil {
			return nil, err
		}
		v[i] = e2
	}

	return v, nil
}

// Values returns a new slice containing values from the pointers in each element of the original slice `s`.
func Values[S ~[]*E, E any](s S) []E {
	return ApplyToAll(s, func(e *E) E {
		return *e
	})
}

// Predicate represents a predicate (boolean-valued function) of one argument.
type Predicate[T any] func(T) bool

// Filter returns a new slice containing all values that return `true` for the filter function `f`.
func Filter[S ~[]E, E any](s S, f Predicate[E]) S {
	v := S(make([]E, 0, len(s)))

	for _, e := range s {
		if f(e) {
			v = append(v, e)
		}
	}

	return slices.Clip(v)
}

// All returns `true` if the filter function `f` returns `true` for all items in slice `s`.
func All[S ~[]E, E any](s S, f Predicate[E]) bool {
	for _, e := range s {
		if !f(e) {
			return false
		}
	}
	return true
}

// Any returns `true` if the filter function `f` returns `true` for any item in slice `s`.
func Any[S ~[]E, E any](s S, f Predicate[E]) bool {
	return slices.ContainsFunc(s, f)
}

// AppendUnique appends unique (not already in the slice) values to a slice.
func AppendUnique[S ~[]E, E comparable](s S, vs ...E) S {
	for _, v := range vs {
		var exists bool

		if slices.Contains(s, v) {
			exists = true
		}

		if !exists {
			s = append(s, v)
		}
	}

	return s
}

// IndexOf returns the index of the first occurrence of `v` in `s`, or -1 if not present.
// This function is similar to the `Index` function in the Go standard `slices` package,
// the difference being that `s` is a slice of `any` and a runtime type check is made.
func IndexOf[S ~[]any, E comparable](s S, v E) int {
	for i := range s {
		if e, ok := s[i].(E); ok && v == e {
			return i
		}
	}
	return -1
}

type signed interface {
	~int | ~int32 | ~int64
}

// Range returns a slice of integers from `start` to `stop` (exclusive) using the specified `step`.
func Range[T signed](start, stop, step T) []T {
	v := make([]T, 0)

	switch {
	case step > 0:
		if start >= stop {
			return nil
		}
		for i := start; i < stop; i += step {
			v = append(v, i)
		}
	case step < 0:
		if start <= stop {
			return nil
		}
		for i := start; i > stop; i += step {
			v = append(v, i)
		}
	default:
		return nil
	}

	return v
}
