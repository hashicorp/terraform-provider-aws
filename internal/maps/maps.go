// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package maps

// ApplyToAllKeys returns a new map containing the results of applying the function `f` to each key of the original map `m`.
func ApplyToAllKeys[M ~map[K1]V, K1, K2 comparable, V any](m M, f func(K1) K2) map[K2]V {
	n := make(map[K2]V, len(m))

	for k, v := range m {
		n[f(k)] = v
	}

	return n
}

// ApplyToAllValues returns a new map containing the results of applying the function `f` to each value of the original map `m`.
func ApplyToAllValues[M ~map[K]V1, K comparable, V1, V2 any](m M, f func(V1) V2) map[K]V2 {
	n := make(map[K]V2, len(m))

	for k, v := range m {
		n[k] = f(v)
	}

	return n
}

func ApplyToAllValuesWithError[M ~map[K]V1, K comparable, V1, V2 any](m M, f func(V1) (V2, error)) (map[K]V2, error) {
	n := make(map[K]V2, len(m))

	for k, v1 := range m {
		v2, err := f(v1)
		if err != nil {
			return nil, err
		}
		n[k] = v2
	}

	return n, nil
}

// Keys returns the keys of the map `m`.
// The keys will be in an indeterminate order.
func Keys[M ~map[K]V, K comparable, V any](m M) []K {
	r := make([]K, 0, len(m))

	for k := range m {
		r = append(r, k)
	}

	return r
}

// Values returns the values of the map `m`.
// The values will be in an indeterminate order.
func Values[M ~map[K]V, K comparable, V any](m M) []V {
	r := make([]V, 0, len(m))

	for _, v := range m {
		r = append(r, v)
	}

	return r
}
