// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"errors"
	"io/fs"
	"path/filepath"
	"testing"
)

func TestParseCIDoc_Golden(t *testing.T) {
	t.Parallel()
	doc, err := ParseCIDoc(filepath.Join("testdata", "golden", "docs", "continuous-integration.md"))
	if err != nil {
		t.Fatalf("ParseCIDoc: %v", err)
	}
	want := []string{"build", "ci", "docs"}
	for _, w := range want {
		if !doc.Refs[w] {
			t.Errorf("expected %q to be referenced", w)
		}
	}
	// "tools" is mentioned in prose only; the regex should still capture it
	// when prefixed by "make". An unrelated word like "implementation" must
	// not be captured.
	if doc.Refs["implementation"] {
		t.Errorf("captured a non-make word as reference")
	}
}

func TestMakeRefRegex(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		input string
		want  []string
	}{
		"plain": {
			input: "Run `make ci` to do everything.",
			want:  []string{"ci"},
		},
		"multiple": {
			input: "Use `make ci` or `make ci-quick`.",
			want:  []string{"ci", "ci-quick"},
		},
		"in code block": {
			input: "```\nmake build\n```\n",
			want:  []string{"build"},
		},
		"avoid matching 'remake'": {
			// `remake` ends with `make` but should NOT trigger.
			input: "We need to remake foo.",
			want:  nil,
		},
		"shell prompt prefix": {
			input: "$ make gh-workflow-lint",
			want:  []string{"gh-workflow-lint"},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			matches := reMakeRef.FindAllStringSubmatch(tc.input, -1)
			var got []string
			for _, m := range matches {
				got = append(got, m[1])
			}
			if len(got) != len(tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
			for i := range got {
				if got[i] != tc.want[i] {
					t.Errorf("got[%d] = %q, want %q", i, got[i], tc.want[i])
				}
			}
		})
	}
}

func TestParseCIDoc_MissingFile(t *testing.T) {
	t.Parallel()
	_, err := ParseCIDoc(filepath.Join(t.TempDir(), "nope.md"))
	if !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("expected not-exist error, got %v", err)
	}
}
