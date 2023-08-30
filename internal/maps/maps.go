// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package maps

// ApplyToAll returns a new map containing the results of applying the function `f` to each element of the original map `m`.
func ApplyToAll[K comparable, T, U any](m map[K]T, f func(T) U) map[K]U {
	n := make(map[K]U, len(m))

	for k, v := range m {
		n[k] = f(v)
	}

	return n
}
