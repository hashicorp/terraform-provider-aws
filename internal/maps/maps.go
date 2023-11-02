// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package maps

// ApplyToAllKeys returns a new map containing the results of applying the function `f` to each key of the original map `m`.
func ApplyToAllKeys[K1, K2 comparable, V any](m map[K1]V, f func(K1) K2) map[K2]V {
	n := make(map[K2]V, len(m))

	for k, v := range m {
		n[f(k)] = v
	}

	return n
}

// ApplyToAllValues returns a new map containing the results of applying the function `f` to each value of the original map `m`.
func ApplyToAllValues[K comparable, V1, V2 any](m map[K]V1, f func(V1) V2) map[K]V2 {
	n := make(map[K]V2, len(m))

	for k, v := range m {
		n[k] = f(v)
	}

	return n
}
