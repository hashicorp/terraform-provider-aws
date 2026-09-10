// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"cmp"
	"fmt"
	"slices"
)

// Severity classifies a finding's impact.
//
// Errors fail the run; warnings are informational unless `-strict` is set.
// We deliberately keep the set small (no "info") so the CLI output is easy
// to scan and filter.
type Severity int

const (
	SevWarn Severity = iota
	SevError
)

func (s Severity) String() string {
	switch s {
	case SevError:
		return "ERROR"
	case SevWarn:
		return "WARN"
	default:
		return "?"
	}
}

// Finding is one alignment problem detected by validation.
//
// File and Line locate the problem in source where possible; Code is a
// stable short identifier (e.g. "phony-ghost") suitable for filtering or
// machine consumption; Suggestion is optional human-readable advice such
// as "did you mean X?" produced from edit-distance matching.
type Finding struct {
	Severity   Severity
	Code       string
	File       string
	Line       int
	Message    string
	Suggestion string
}

// Inputs bundles the parsed view of every file involved in alignment.
type Inputs struct {
	MakefilePath   string
	CheatSheetPath string
	CIDocPath      string

	Make       *MakefileData
	CheatSheet []CheatSheetRow
	CIDoc      *CIDoc
}

// Validate runs every alignment rule against the parsed inputs and returns
// findings sorted by file, line, then code so output is deterministic.
func Validate(in *Inputs) []Finding {
	var findings []Finding

	cheatByName := make(map[string]CheatSheetRow, len(in.CheatSheet))
	for _, r := range in.CheatSheet {
		cheatByName[r.Target] = r
	}
	phonySet := make(map[string]bool, len(in.Make.Phony))
	for _, p := range in.Make.Phony {
		phonySet[p] = true
	}

	findings = append(findings, checkPhonyAlignment(in, phonySet)...)
	findings = append(findings, checkCheatSheetAlignment(in, cheatByName)...)
	findings = append(findings, checkCIDocAlignment(in, cheatByName)...)
	findings = append(findings, checkOrdering(in)...)

	slices.SortStableFunc(findings, func(a, b Finding) int {
		if c := cmp.Compare(a.File, b.File); c != 0 {
			return c
		}
		if c := cmp.Compare(a.Line, b.Line); c != 0 {
			return c
		}
		return cmp.Compare(a.Code, b.Code)
	})
	return findings
}

// checkPhonyAlignment verifies that every Makefile target appears in the
// .PHONY list and vice-versa. These are the highest-impact checks because
// a `.PHONY` mismatch can cause silent breakage when a same-named file
// happens to exist in the working directory.
func checkPhonyAlignment(in *Inputs, phonySet map[string]bool) []Finding {
	var out []Finding
	targetNames := slices.Clone(in.Make.Order)

	for _, name := range in.Make.Order {
		if phonySet[name] {
			continue
		}
		out = append(out, Finding{
			Severity: SevError,
			Code:     "phony-missing",
			File:     in.MakefilePath,
			Line:     in.Make.Targets[name].Line,
			Message:  fmt.Sprintf("target %q is not in the .PHONY list", name),
		})
	}

	for _, p := range in.Make.Phony {
		if _, ok := in.Make.Targets[p]; ok {
			continue
		}
		f := Finding{
			Severity: SevError,
			Code:     "phony-ghost",
			File:     in.MakefilePath,
			Line:     in.Make.PhonyLine,
			Message:  fmt.Sprintf(".PHONY entry %q does not match any target", p),
		}
		if hint := suggest(p, targetNames); hint != "" {
			f.Suggestion = fmt.Sprintf("did you mean %q?", hint)
		}
		out = append(out, f)
	}
	return out
}

// checkCheatSheetAlignment verifies the cheat sheet references real
// targets (errors) and that documented targets are listed (warnings).
//
// A "documented" target is one with a `## description` comment, which is
// the same set of targets exposed by `make help`. Targets without `##` are
// considered internal and are not expected to appear in the cheat sheet.
func checkCheatSheetAlignment(in *Inputs, cheatByName map[string]CheatSheetRow) []Finding {
	var out []Finding
	targetNames := slices.Clone(in.Make.Order)

	for _, row := range in.CheatSheet {
		t, ok := in.Make.Targets[row.Target]
		if !ok {
			f := Finding{
				Severity: SevError,
				Code:     "cheatsheet-ghost",
				File:     in.CheatSheetPath,
				Line:     row.Line,
				Message:  fmt.Sprintf("cheat sheet target %q is not defined in GNUmakefile", row.Target),
			}
			if hint := suggest(row.Target, targetNames); hint != "" {
				f.Suggestion = fmt.Sprintf("did you mean %q?", hint)
			}
			out = append(out, f)
			continue
		}
		// CI flag must agree.
		if t.IsCI != row.IsCI {
			out = append(out, Finding{
				Severity: SevError,
				Code:     "ci-flag-mismatch",
				File:     in.CheatSheetPath,
				Line:     row.Line,
				Message: fmt.Sprintf(
					"target %q: GNUmakefile has [CI]=%t but cheat sheet CI? column is %s",
					row.Target, t.IsCI, checkmark(row.IsCI),
				),
			})
		}
		// Legacy flag must agree.
		if t.IsLegacy != row.IsLegacy {
			out = append(out, Finding{
				Severity: SevWarn,
				Code:     "legacy-flag-mismatch",
				File:     in.CheatSheetPath,
				Line:     row.Line,
				Message: fmt.Sprintf(
					"target %q: GNUmakefile mentions Legacy=%t but cheat sheet Legacy? column is %s",
					row.Target, t.IsLegacy, checkmark(row.IsLegacy),
				),
			})
		}
	}

	for _, name := range in.Make.Order {
		t := in.Make.Targets[name]
		if !t.HasDoc || t.IsInternal {
			continue
		}
		if _, ok := cheatByName[name]; ok {
			continue
		}
		out = append(out, Finding{
			Severity: SevWarn,
			Code:     "cheatsheet-missing",
			File:     in.MakefilePath,
			Line:     t.Line,
			Message:  fmt.Sprintf("documented target %q is missing from the cheat sheet", name),
		})
	}
	return out
}

// checkCIDocAlignment ensures every CI-marked target is mentioned in
// continuous-integration.md.
//
// Meta targets (marked `<sup>M</sup>` in the cheat sheet) are exempt:
// they're aggregations like `ci`, `docs`, or `website` whose constituent
// targets carry the actual CI documentation. Surfacing the parent would
// just create noise.
//
// We warn rather than error so the rule never blocks a PR by itself --
// CI documentation lags target additions in practice and we want this
// check to be informative, not punitive.
func checkCIDocAlignment(in *Inputs, cheatByName map[string]CheatSheetRow) []Finding {
	var out []Finding
	for _, name := range in.Make.Order {
		t := in.Make.Targets[name]
		if !t.IsCI {
			continue
		}
		if row, ok := cheatByName[name]; ok && row.IsMeta {
			continue
		}
		if in.CIDoc.Refs[name] {
			continue
		}
		out = append(out, Finding{
			Severity: SevWarn,
			Code:     "ci-doc-missing",
			File:     in.MakefilePath,
			Line:     t.Line,
			Message:  fmt.Sprintf("CI target %q is not referenced in continuous-integration.md", name),
		})
	}
	return out
}

// checkOrdering reports the first out-of-order pair in the .PHONY list and
// in the cheat sheet table. The Makefile and cheat sheet both carry "keep
// alphabetical" comments, so a single warning per file is enough to nudge
// the contributor without spamming.
func checkOrdering(in *Inputs) []Finding {
	var out []Finding

	for i := 1; i < len(in.Make.Phony); i++ {
		if in.Make.Phony[i-1] > in.Make.Phony[i] {
			out = append(out, Finding{
				Severity: SevWarn,
				Code:     "phony-order",
				File:     in.MakefilePath,
				Line:     in.Make.PhonyLine,
				Message: fmt.Sprintf(
					".PHONY list is not alphabetized: %q comes before %q",
					in.Make.Phony[i-1], in.Make.Phony[i],
				),
			})
			break
		}
	}

	names := make([]string, len(in.CheatSheet))
	for i, r := range in.CheatSheet {
		names[i] = r.Target
	}
	for i := 1; i < len(names); i++ {
		if names[i-1] > names[i] {
			out = append(out, Finding{
				Severity: SevWarn,
				Code:     "cheatsheet-order",
				File:     in.CheatSheetPath,
				Line:     in.CheatSheet[i-1].Line,
				Message: fmt.Sprintf(
					"cheat sheet is not alphabetized: %q comes before %q",
					names[i-1], names[i],
				),
			})
			break
		}
	}
	return out
}

// checkmark renders a boolean as the symbol used in the cheat sheet, for
// finding messages that describe what the column "is".
func checkmark(v bool) string {
	if v {
		return "✔"
	}
	return "(empty)"
}

// suggest returns the candidate closest to `name` by Levenshtein distance,
// provided the distance is small enough to be plausibly a typo. Empty
// string means no suggestion.
func suggest(name string, candidates []string) string {
	const maxDistance = 3
	best := ""
	bestDist := maxDistance + 1
	for _, c := range candidates {
		if c == name {
			continue
		}
		d := levenshtein(name, c)
		if d < bestDist {
			bestDist = d
			best = c
		}
	}
	if bestDist > maxDistance {
		return ""
	}
	return best
}

// levenshtein computes the edit distance between a and b using the classic
// O(len(a)*len(b)) DP. Inputs in this tool are short target names, so the
// allocation cost is negligible.
func levenshtein(a, b string) int {
	if a == b {
		return 0
	}
	ra, rb := []rune(a), []rune(b)
	if len(ra) == 0 {
		return len(rb)
	}
	if len(rb) == 0 {
		return len(ra)
	}
	prev := make([]int, len(rb)+1)
	curr := make([]int, len(rb)+1)
	for j := range prev {
		prev[j] = j
	}
	for i := 1; i <= len(ra); i++ {
		curr[0] = i
		for j := 1; j <= len(rb); j++ {
			cost := 1
			if ra[i-1] == rb[j-1] {
				cost = 0
			}
			curr[j] = min3(
				prev[j]+1,      // deletion
				curr[j-1]+1,    // insertion
				prev[j-1]+cost, // substitution
			)
		}
		prev, curr = curr, prev
	}
	return prev[len(rb)]
}

func min3(a, b, c int) int {
	m := a
	if b < m {
		m = b
	}
	if c < m {
		m = c
	}
	return m
}
