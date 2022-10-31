package slices

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
