package verify

import (
	"testing"
)

func TestSemVerLessThan(t *testing.T) {
	for _, tc := range []struct {
		s1 string
		s2 string
		lt bool
	}{
		{"1.0", "2.0", true},
		{"3.0", "2.0", false},
		{"4.0", "4.0", false},
		{"2", "10", true},
		{"abc", "xyz", false},
	} {
		lt := SemVerLessThan(tc.s1, tc.s2)
		if tc.lt != lt {
			t.Fatalf("SemVerLessThan(%q, %q) should be: %t", tc.s1, tc.s2, tc.lt)
		}
	}
}

func TestSemVerGreaterThanOrEqual(t *testing.T) {
	for _, tc := range []struct {
		s1 string
		s2 string
		lt bool
	}{
		{"1.0", "2.0", false},
		{"3.0", "2.0", true},
		{"4.0", "4.0", true},
		{"2", "10", false},
		{"abc", "xyz", false},
	} {
		lt := SemVerGreaterThanOrEqual(tc.s1, tc.s2)
		if tc.lt != lt {
			t.Fatalf("SemVerGreaterThanOrEqual(%q, %q) should be: %t", tc.s1, tc.s2, tc.lt)
		}
	}
}
