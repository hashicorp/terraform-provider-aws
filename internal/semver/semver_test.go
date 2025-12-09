// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package semver

import (
	"testing"
)

func TestSemVerLessThan(t *testing.T) {
	t.Parallel()

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
		lt := LessThan(tc.s1, tc.s2)
		if tc.lt != lt {
			t.Fatalf("SemVerLessThan(%q, %q) should be: %t", tc.s1, tc.s2, tc.lt)
		}
	}
}

func TestSemVerGreaterThanOrEqual(t *testing.T) {
	t.Parallel()

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
		lt := GreaterThanOrEqual(tc.s1, tc.s2)
		if tc.lt != lt {
			t.Fatalf("SemVerGreaterThanOrEqual(%q, %q) should be: %t", tc.s1, tc.s2, tc.lt)
		}
	}
}

func TestSemVerMajorMinor(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		s1 string
		s2 string
	}{
		{"1.0", "1.0"},
		{"2.3.4", "2.3"},
		{"4", "4.0"},
		{"v5.6.7", "5.6"},
		{"bloop", ""},
	} {
		got, _ := MajorMinor(tc.s1)
		if got != tc.s2 {
			t.Fatalf("MajorMinor(%q) should be: %q", tc.s1, tc.s2)
		}
	}
}
