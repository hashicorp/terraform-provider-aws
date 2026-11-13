// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"path/filepath"
	"testing"
)

func TestParseTargetLine(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		line         string
		wantName     string
		wantOK       bool
		wantDoc      bool
		wantCI       bool
		wantInternal bool
		wantDesc     string
	}{
		"plain target": {
			line:     "build:",
			wantName: "build",
			wantOK:   true,
		},
		"target with deps": {
			line:     "ci: tools build",
			wantName: "ci",
			wantOK:   true,
		},
		"documented target": {
			line:     "build: prereq-go ## Build provider",
			wantName: "build",
			wantOK:   true,
			wantDoc:  true,
			wantDesc: "Build provider",
		},
		"CI documented target": {
			line:     "ci: tools build ## [CI] Run all CI checks",
			wantName: "ci",
			wantOK:   true,
			wantDoc:  true,
			wantCI:   true,
			wantDesc: "Run all CI checks",
		},
		"internal target": {
			line:         "test-single-service: ## [internal] test single service",
			wantName:     "test-single-service",
			wantOK:       true,
			wantDoc:      true,
			wantInternal: true,
			wantDesc:     "test single service",
		},
		"variable assignment rejected": {
			line:   "GO_VER := go1.22",
			wantOK: false,
		},
		"variable simple equals rejected": {
			line:   "PKG_NAME = internal",
			wantOK: false,
		},
		"variable conditional rejected": {
			line:   "PKG_NAME ?= internal",
			wantOK: false,
		},
		"recipe line rejected": {
			line:   "\t@echo hello",
			wantOK: false,
		},
		"indented logical line rejected": {
			line:   "    target: deps",
			wantOK: false,
		},
		"empty line rejected": {
			line:   "",
			wantOK: false,
		},
		"double colon target accepted": {
			line:     "all:: build test",
			wantName: "all",
			wantOK:   true,
		},
		"target with hyphens and digits": {
			line:     "golangci-lint5: ## [CI] golangci-lint Checks / 5 of 5",
			wantName: "golangci-lint5",
			wantOK:   true,
			wantDoc:  true,
			wantCI:   true,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got, ok := parseTargetLine(tc.line, 1)
			if ok != tc.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tc.wantOK)
			}
			if !ok {
				return
			}
			if got.Name != tc.wantName {
				t.Errorf("Name = %q, want %q", got.Name, tc.wantName)
			}
			if got.HasDoc != tc.wantDoc {
				t.Errorf("HasDoc = %v, want %v", got.HasDoc, tc.wantDoc)
			}
			if got.IsCI != tc.wantCI {
				t.Errorf("IsCI = %v, want %v", got.IsCI, tc.wantCI)
			}
			if got.IsInternal != tc.wantInternal {
				t.Errorf("IsInternal = %v, want %v", got.IsInternal, tc.wantInternal)
			}
			if tc.wantDesc != "" && got.Description != tc.wantDesc {
				t.Errorf("Description = %q, want %q", got.Description, tc.wantDesc)
			}
		})
	}
}

func TestIsValidTargetName(t *testing.T) {
	t.Parallel()
	good := []string{"a", "build", "ci-quick", "golangci-lint5", "_underscore", "fix-imports-core"}
	bad := []string{"", "1bad", "-bad", "$(VAR)", "has space", "has.dot"}
	for _, s := range good {
		if !isValidTargetName(s) {
			t.Errorf("expected %q to be valid", s)
		}
	}
	for _, s := range bad {
		if isValidTargetName(s) {
			t.Errorf("expected %q to be invalid", s)
		}
	}
}

func TestParseMakefile_Golden(t *testing.T) {
	t.Parallel()
	mf, err := ParseMakefile(filepath.Join("testdata", "golden", "GNUmakefile"))
	if err != nil {
		t.Fatalf("ParseMakefile: %v", err)
	}

	// Spot-check structural expectations against the curated fixture so
	// the test fails loudly if the parser regresses on a documented case.
	wantTargets := []string{"build", "ci", "deps", "docs", "missing-from-phony"}
	for _, name := range wantTargets {
		if _, ok := mf.Targets[name]; !ok {
			t.Errorf("missing target %q in parsed targets", name)
		}
	}
	if mf.Targets["ci"] == nil || !mf.Targets["ci"].IsCI {
		t.Errorf("ci target should have IsCI=true")
	}
	if mf.Targets["build"] == nil || mf.Targets["build"].Description != "Build provider" {
		t.Errorf("build target description = %q, want %q",
			mf.Targets["build"].Description, "Build provider")
	}
	wantPhony := []string{"build", "ci", "docs", "ghost-target", "deps"}
	if len(mf.Phony) != len(wantPhony) {
		t.Fatalf("Phony len = %d, want %d (%v)", len(mf.Phony), len(wantPhony), mf.Phony)
	}
	for i, p := range wantPhony {
		if mf.Phony[i] != p {
			t.Errorf("Phony[%d] = %q, want %q", i, mf.Phony[i], p)
		}
	}
}

func TestCollectPhonyLine(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		line     string
		wantList []string
		wantCont bool
	}{
		"continuation": {
			line:     "\tbuild \\",
			wantList: []string{"build"},
			wantCont: true,
		},
		"final line no continuation": {
			line:     "\tbuild",
			wantList: []string{"build"},
			wantCont: false,
		},
		"multiple targets one line": {
			line:     "\tbuild test docs \\",
			wantList: []string{"build", "test", "docs"},
			wantCont: true,
		},
		"empty line stops": {
			line:     "",
			wantList: nil,
			wantCont: false,
		},
		"trailing whitespace before backslash": {
			line:     "\tbuild  \\",
			wantList: []string{"build"},
			wantCont: true,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			var got []string
			cont := collectPhonyLine(tc.line, &got)
			if cont != tc.wantCont {
				t.Errorf("cont = %v, want %v", cont, tc.wantCont)
			}
			if len(got) != len(tc.wantList) {
				t.Fatalf("list = %v, want %v", got, tc.wantList)
			}
			for i := range got {
				if got[i] != tc.wantList[i] {
					t.Errorf("list[%d] = %q, want %q", i, got[i], tc.wantList[i])
				}
			}
		})
	}
}
