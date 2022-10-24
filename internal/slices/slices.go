package slices

// Reverse reverses a slice in place.
func Reverse[S ~[]E, E any](s S) S {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}

	return s
}

// Reversed returns a reversed copy of the slice.
func Reversed[S ~[]E, E any](s S) S {
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
