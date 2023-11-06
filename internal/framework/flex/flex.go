// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package flex

type Set[T comparable] []T

// Difference find the elements in two sets that are not similar.
func (s Set[T]) Difference(ns Set[T]) Set[T] {
	m := make(map[T]struct{})
	for _, v := range ns {
		m[v] = struct{}{}
	}

	var result []T
	for _, v := range s {
		if _, ok := m[v]; !ok {
			result = append(result, v)
		}
	}
	return result
}
