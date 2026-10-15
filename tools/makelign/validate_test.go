// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"path/filepath"
	"testing"
)

// TestValidate_Golden runs the full validation pipeline against a curated
// fixture set that includes one example of every rule violation. It serves
// as both an integration test and a guard against accidental rule churn.
func TestValidate_Golden(t *testing.T) {
	t.Parallel()
	in := loadGolden(t)
	got := Validate(in)

	// Index findings by code so the test doesn't depend on output order.
	byCode := make(map[string]int)
	for _, f := range got {
		byCode[f.Code]++
	}
	wantAtLeast := map[string]int{
		"phony-missing":      1, // missing-from-phony, default
		"phony-ghost":        1, // ghost-target
		"cheatsheet-ghost":   1, // ghost-row
		"cheatsheet-missing": 1, // deps (documented, not in cheat sheet)
	}
	for code, n := range wantAtLeast {
		if byCode[code] < n {
			t.Errorf("code %q: got %d, want >= %d (findings: %+v)", code, byCode[code], n, got)
		}
	}
}

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

// TestValidate_RulesIsolated drives each rule in isolation by constructing
// minimal Inputs structs. This makes failures easy to localize without
// having to interpret the larger golden fixture.
func TestValidate_RulesIsolated(t *testing.T) {
	t.Parallel()

	t.Run("phony-missing", func(t *testing.T) {
		t.Parallel()
		in := minimalInputs()
		in.Make.Targets["new-target"] = &Target{Name: "new-target", Line: 100, HasDoc: true}
		in.Make.Order = append(in.Make.Order, "new-target")
		got := Validate(in)
		if !hasFinding(got, "phony-missing", "new-target") {
			t.Errorf("expected phony-missing for new-target, got %+v", got)
		}
	})

	t.Run("phony-ghost with suggestion", func(t *testing.T) {
		t.Parallel()
		in := minimalInputs()
		in.Make.Phony = append(in.Make.Phony, "buld") // typo of "build"
		got := Validate(in)
		var f *Finding
		for i := range got {
			if got[i].Code == "phony-ghost" {
				f = &got[i]
				break
			}
		}
		if f == nil {
			t.Fatalf("expected phony-ghost finding")
		}
		if f.Suggestion == "" {
			t.Errorf("expected a 'did you mean' suggestion for buld -> build")
		}
	})

	t.Run("cheatsheet-ghost", func(t *testing.T) {
		t.Parallel()
		in := minimalInputs()
		in.CheatSheet = append(in.CheatSheet, CheatSheetRow{Target: "nonexistent", Line: 50})
		got := Validate(in)
		if !hasFinding(got, "cheatsheet-ghost", "nonexistent") {
			t.Errorf("expected cheatsheet-ghost for nonexistent")
		}
	})

	t.Run("ci-flag-mismatch", func(t *testing.T) {
		t.Parallel()
		in := minimalInputs()
		// build is non-CI in the makefile but cheat sheet marks it CI.
		in.CheatSheet = []CheatSheetRow{{Target: "build", Line: 10, IsCI: true}}
		got := Validate(in)
		if !hasFinding(got, "ci-flag-mismatch", "build") {
			t.Errorf("expected ci-flag-mismatch for build")
		}
	})

	t.Run("internal target exempt from cheatsheet-missing", func(t *testing.T) {
		t.Parallel()
		in := minimalInputs()
		in.Make.Targets["helper"] = &Target{Name: "helper", Line: 50, HasDoc: true, IsInternal: true}
		in.Make.Order = append(in.Make.Order, "helper")
		in.Make.Phony = append(in.Make.Phony, "helper")
		got := Validate(in)
		for _, f := range got {
			if f.Code == "cheatsheet-missing" && contains(f.Message, "helper") {
				t.Errorf("internal target should be exempt from cheatsheet-missing, got %v", f)
			}
		}
	})

	t.Run("meta target exempt from ci-doc-missing", func(t *testing.T) {
		t.Parallel()
		in := minimalInputs()
		in.Make.Targets["meta-ci"] = &Target{Name: "meta-ci", Line: 200, HasDoc: true, IsCI: true}
		in.Make.Order = append(in.Make.Order, "meta-ci")
		in.Make.Phony = append(in.Make.Phony, "meta-ci")
		in.CheatSheet = append(in.CheatSheet, CheatSheetRow{Target: "meta-ci", Line: 5, IsMeta: true, IsCI: true})
		got := Validate(in)
		// Should NOT have ci-doc-missing because meta targets are exempt.
		for _, f := range got {
			if f.Code == "ci-doc-missing" && contains(f.Message, "meta-ci") {
				t.Errorf("meta target should be exempt from ci-doc-missing, got %v", f)
			}
		}
	})
}

// minimalInputs builds an Inputs with one valid build target so individual
// tests can mutate just the field under test.
func minimalInputs() *Inputs {
	return &Inputs{
		MakefilePath:   "GNUmakefile",
		CheatSheetPath: "docs/cheat.md",
		CIDocPath:      "docs/ci.md",
		Make: &MakefileData{
			Targets: map[string]*Target{
				"build": {Name: "build", Line: 1, HasDoc: true, Description: "Build"},
			},
			Order:     []string{"build"},
			Phony:     []string{"build"},
			PhonyLine: 100,
		},
		CheatSheet: []CheatSheetRow{{Target: "build", Line: 1}},
		CIDoc:      &CIDoc{Refs: map[string]bool{"build": true}},
	}
}

func loadGolden(t *testing.T) *Inputs {
	t.Helper()
	dir := filepath.Join("testdata", "golden")
	mf, err := ParseMakefile(filepath.Join(dir, "GNUmakefile"))
	if err != nil {
		t.Fatalf("ParseMakefile: %v", err)
	}
	sheet, err := ParseCheatSheet(filepath.Join(dir, "docs", "makefile-cheat-sheet.md"))
	if err != nil {
		t.Fatalf("ParseCheatSheet: %v", err)
	}
	cidoc, err := ParseCIDoc(filepath.Join(dir, "docs", "continuous-integration.md"))
	if err != nil {
		t.Fatalf("ParseCIDoc: %v", err)
	}
	return &Inputs{
		MakefilePath:   "GNUmakefile",
		CheatSheetPath: "docs/makefile-cheat-sheet.md",
		CIDocPath:      "docs/continuous-integration.md",
		Make:           mf,
		CheatSheet:     sheet,
		CIDoc:          cidoc,
	}
}

func hasFinding(fs []Finding, code, mustContain string) bool {
	for _, f := range fs {
		if f.Code == code && contains(f.Message, mustContain) {
			return true
		}
	}
	return false
}

func contains(haystack, needle string) bool {
	if needle == "" {
		return true
	}
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
