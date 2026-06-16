// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"testing"
)

func TestSuggest(t *testing.T) {
	t.Parallel()
	candidates := []string{"gh-workflow-lint", "go-build", "go-misspell", "gen", "build"}
	tests := map[string]struct {
		name string
		want string
	}{
		"single typo": {
			name: "gh-workflows-lint",
			want: "gh-workflow-lint",
		},
		"close variant": {
			name: "go-bild",
			want: "go-build",
		},
		"too far": {
			name: "completely-unrelated-target-name",
			want: "",
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got := suggest(tc.name, candidates)
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestLevenshtein(t *testing.T) {
	t.Parallel()
	tests := []struct {
		a, b string
		want int
	}{
		{"", "", 0},
		{"a", "", 1},
		{"", "abc", 3},
		{"kitten", "sitting", 3},
		{"build", "build", 0},
		{"gh-workflow-lint", "gh-workflows-lint", 1},
	}
	for _, tc := range tests {
		t.Run(tc.a+"->"+tc.b, func(t *testing.T) {
			t.Parallel()
			if got := levenshtein(tc.a, tc.b); got != tc.want {
				t.Errorf("levenshtein(%q,%q) = %d, want %d", tc.a, tc.b, got, tc.want)
			}
		})
	}
}
