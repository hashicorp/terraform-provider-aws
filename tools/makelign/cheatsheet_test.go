// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"path/filepath"
	"slices"
	"testing"
)

func TestParseCheatRow(t *testing.T) {
	t.Parallel()
	tests := map[string]struct {
		line       string
		wantTarget string
		wantMeta   bool
		wantDep    bool
		wantCI     bool
		wantLegacy bool
		wantVars   []string
		wantOK     bool
	}{
		"simple row": {
			line:       "| `build` | Build the provider |  |  |  |",
			wantTarget: "build",
			wantOK:     true,
		},
		"meta annotation": {
			line:       "| `ci`<sup>M</sup> | Run all CI checks | ✔️ |  | `K`, `PKG` |",
			wantTarget: "ci",
			wantMeta:   true,
			wantCI:     true,
			wantVars:   []string{"K", "PKG"},
			wantOK:     true,
		},
		"dependent annotation": {
			line:       "| `testacc`<sup>D</sup> | Run acceptance tests |  |  | `GO_VER` |",
			wantTarget: "testacc",
			wantDep:    true,
			wantVars:   []string{"GO_VER"},
			wantOK:     true,
		},
		"italic default row": {
			line:       "| _default_ | = `build` |  |  | `GO_VER` |",
			wantTarget: "default",
			wantVars:   []string{"GO_VER"},
			wantOK:     true,
		},
		"legacy row": {
			line:       "| `docs-check` | Check provider documentation |  | ✔️ |  |",
			wantTarget: "docs-check",
			wantLegacy: true,
			wantOK:     true,
		},
		"non-row rejected": {
			line:   "this is not a table row",
			wantOK: false,
		},
		"too few columns rejected": {
			line:   "| `build` | desc |",
			wantOK: false,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got, ok := parseCheatRow(tc.line, 42)
			if ok != tc.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tc.wantOK)
			}
			if !ok {
				return
			}
			if got.Target != tc.wantTarget {
				t.Errorf("Target = %q, want %q", got.Target, tc.wantTarget)
			}
			if got.IsMeta != tc.wantMeta {
				t.Errorf("IsMeta = %v, want %v", got.IsMeta, tc.wantMeta)
			}
			if got.IsDependent != tc.wantDep {
				t.Errorf("IsDependent = %v, want %v", got.IsDependent, tc.wantDep)
			}
			if got.IsCI != tc.wantCI {
				t.Errorf("IsCI = %v, want %v", got.IsCI, tc.wantCI)
			}
			if got.IsLegacy != tc.wantLegacy {
				t.Errorf("IsLegacy = %v, want %v", got.IsLegacy, tc.wantLegacy)
			}
			if !slices.Equal(got.Vars, tc.wantVars) {
				t.Errorf("Vars = %v, want %v", got.Vars, tc.wantVars)
			}
			if got.Line != 42 {
				t.Errorf("Line = %d, want 42", got.Line)
			}
		})
	}
}

func TestParseCheatSheet_Golden(t *testing.T) {
	t.Parallel()
	rows, err := ParseCheatSheet(filepath.Join("testdata", "golden", "docs", "makefile-cheat-sheet.md"))
	if err != nil {
		t.Fatalf("ParseCheatSheet: %v", err)
	}
	wantTargets := []string{"build", "ci", "docs", "ghost-row"}
	if len(rows) != len(wantTargets) {
		t.Fatalf("got %d rows, want %d (%v)", len(rows), len(wantTargets), rows)
	}
	for i, want := range wantTargets {
		if rows[i].Target != want {
			t.Errorf("rows[%d].Target = %q, want %q", i, rows[i].Target, want)
		}
	}
}
