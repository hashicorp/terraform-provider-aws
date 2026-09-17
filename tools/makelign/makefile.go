// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"bufio"
	"os"
	"strings"
)

// Target describes a GNUmakefile target rule.
//
// A target is "documented" if its definition line carries an inline comment
// of the form `## description`. Documented targets are surfaced in `make help`
// and are expected to appear in the cheat sheet, unless IsInternal is true.
type Target struct {
	Name        string
	Description string // text after ## (without [CI] or [internal] prefix)
	IsCI        bool   // description began with "[CI]"
	IsInternal  bool   // description began with "[internal]" — shown in source, hidden from make help and cheat sheet
	IsLegacy    bool   // description mentions "Legacy" (case-insensitive)
	HasDoc      bool   // line carried a `##` description
	Line        int    // 1-based source line of the definition
}

// MakefileData is the parsed view of a GNUmakefile relevant to alignment checks.
type MakefileData struct {
	// Targets, keyed by target name. The first definition wins if a target
	// happens to be defined multiple times (rare in practice).
	Targets map[string]*Target

	// Order is target names in source order, useful for stable reporting.
	Order []string

	// Phony is the list of names appearing in the `.PHONY:` declaration,
	// preserving source order so we can report alphabetization issues.
	Phony []string

	// PhonyLine is the 1-based line number of the `.PHONY:` declaration,
	// used as the location for findings about the phony list itself.
	PhonyLine int
}

// ParseMakefile reads a GNUmakefile from disk and extracts targets and the
// .PHONY list. Errors are returned only for I/O problems; malformed Make
// syntax is tolerated since the tool is descriptive, not a full parser.
func ParseMakefile(path string) (*MakefileData, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	data := &MakefileData{
		Targets: make(map[string]*Target),
	}

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	lineNum := 0
	inPhony := false

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// .PHONY collection has its own state machine and short-circuits
		// other parsing for the duration of the declaration.
		if inPhony {
			cont := collectPhonyLine(line, &data.Phony)
			if !cont {
				inPhony = false
			}
			continue
		}
		if strings.HasPrefix(line, ".PHONY:") {
			data.PhonyLine = lineNum
			rest := strings.TrimPrefix(line, ".PHONY:")
			cont := collectPhonyLine(rest, &data.Phony)
			inPhony = cont
			continue
		}

		// Target rules: not indented, name : ... [## desc]
		if t, ok := parseTargetLine(line, lineNum); ok {
			if _, exists := data.Targets[t.Name]; !exists {
				data.Targets[t.Name] = t
				data.Order = append(data.Order, t.Name)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return data, nil
}

// parseTargetLine attempts to interpret `line` as a target rule definition.
// It rejects indented lines (recipes), variable assignments, and conditional
// directives. When a target is found, the returned Target carries any inline
// `## description` and derived flags.
func parseTargetLine(line string, lineNum int) (*Target, bool) {
	// Recipe lines start with a tab; skip immediately.
	if line == "" || line[0] == '\t' {
		return nil, false
	}
	// Indented logical lines (4+ spaces, etc.) are also not targets.
	if line[0] == ' ' {
		return nil, false
	}

	// Split off the inline description, if any.
	var desc string
	hasDoc := false
	main := line
	if idx := strings.Index(line, "##"); idx >= 0 {
		main = line[:idx]
		desc = strings.TrimSpace(line[idx+2:])
		hasDoc = true
	}
	main = strings.TrimRight(main, " \t")

	// A target needs `:` not preceded by `=` (otherwise it's a variable).
	colonIdx := strings.Index(main, ":")
	if colonIdx < 0 {
		return nil, false
	}
	if eqIdx := strings.Index(main, "="); eqIdx >= 0 && eqIdx < colonIdx {
		return nil, false
	}
	// Reject `:=` style assignments. Double colons (`::`) are allowed.
	if colonIdx+1 < len(main) && main[colonIdx+1] == '=' {
		return nil, false
	}

	name := strings.TrimSpace(main[:colonIdx])
	if !isValidTargetName(name) {
		return nil, false
	}

	t := &Target{
		Name:     name,
		HasDoc:   hasDoc,
		Line:     lineNum,
		IsLegacy: strings.Contains(strings.ToLower(desc), "legacy"),
	}
	if hasDoc {
		// `## [CI] foo`       => IsCI true, Description "foo"
		// `## [internal] foo` => IsInternal true, Description "foo"
		const ciPrefix = "[CI]"
		const internalPrefix = "[internal]"
		switch {
		case strings.HasPrefix(desc, ciPrefix):
			t.IsCI = true
			desc = strings.TrimSpace(strings.TrimPrefix(desc, ciPrefix))
		case strings.HasPrefix(desc, internalPrefix):
			t.IsInternal = true
			desc = strings.TrimSpace(strings.TrimPrefix(desc, internalPrefix))
		}
		t.Description = desc
	}
	return t, true
}

// collectPhonyLine extracts target names from a single line of a `.PHONY:`
// declaration (possibly with line continuation `\`). Returns true if the
// declaration continues on the next line.
func collectPhonyLine(line string, phony *[]string) bool {
	trimmed := strings.TrimRight(line, " \t")
	cont := strings.HasSuffix(trimmed, "\\")
	body := strings.TrimSuffix(trimmed, "\\")
	for _, tok := range strings.Fields(body) {
		if isValidTargetName(tok) {
			*phony = append(*phony, tok)
		}
	}
	return cont
}

// isValidTargetName checks whether s is a plausible Make target name. We
// deliberately disallow `$` and `(` so we don't capture variable references
// that happen to look like `name:` after expansion.
func isValidTargetName(s string) bool {
	if s == "" {
		return false
	}
	for i, r := range s {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r == '_':
		case i > 0 && r >= '0' && r <= '9':
		case i > 0 && r == '-':
		default:
			return false
		}
	}
	return true
}
