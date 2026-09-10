// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"bufio"
	"os"
	"regexp"
	"strings"
)

// CheatSheetRow is one row from the cheat sheet's target table.
//
// The "Target" column may render the name as a backtick code span with an
// optional `<sup>M</sup>` or `<sup>D</sup>` annotation, or as `_name_`
// italic for the synthetic `default` row. Both forms produce the same
// Target string here.
type CheatSheetRow struct {
	Target      string
	IsMeta      bool // M annotation
	IsDependent bool // D annotation
	Description string
	IsCI        bool     // ✔️ in the CI? column
	IsLegacy    bool     // ✔️ in the Legacy? column
	Vars        []string // names parsed from the Vars column (no backticks)
	Line        int      // 1-based source line
}

// ParseCheatSheet reads the makefile cheat sheet markdown and returns one
// CheatSheetRow per data row of the targets table. Non-target rows and
// header/separator rows are skipped silently.
func ParseCheatSheet(path string) ([]CheatSheetRow, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	var rows []CheatSheetRow
	inTable := false
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := scanner.Text()

		// The targets table is identified by its header row. Once entered,
		// we stay in table mode until we hit a non-table line.
		if !inTable {
			if isCheatSheetTableHeader(line) {
				inTable = true
			}
			continue
		}
		if !strings.HasPrefix(line, "|") {
			inTable = false
			continue
		}
		// Markdown separator row: `| --- | --- | ... |`.
		if strings.Contains(line, "---") {
			continue
		}
		if row, ok := parseCheatRow(line, lineNum); ok {
			rows = append(rows, row)
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return rows, nil
}

// isCheatSheetTableHeader recognizes the targets table header. Other tables
// in the document (e.g. variable references) do not start with "Target".
func isCheatSheetTableHeader(line string) bool {
	trimmed := strings.TrimSpace(line)
	return strings.HasPrefix(trimmed, "| Target |")
}

// reTargetCell matches the first cell of a targets-table row, capturing
// either a backtick-wrapped name (with optional M/D annotation) or an
// italic `_name_` cell used for the `default` row.
var reTargetCell = regexp.MustCompile(
	"^`([a-zA-Z][a-zA-Z0-9_-]*)`(?:<sup>([MD])</sup>)?$|" +
		"^_([a-zA-Z][a-zA-Z0-9_-]*)_$",
)

// reVarName extracts variable names from backticks in the Vars column.
var reVarName = regexp.MustCompile("`([A-Z][A-Z0-9_]*)`")

func parseCheatRow(line string, lineNum int) (CheatSheetRow, bool) {
	cells := splitMarkdownRow(line)
	if len(cells) < 5 {
		return CheatSheetRow{}, false
	}

	row := CheatSheetRow{Line: lineNum}
	m := reTargetCell.FindStringSubmatch(cells[0])
	switch {
	case m == nil:
		return CheatSheetRow{}, false
	case m[1] != "":
		row.Target = m[1]
		switch m[2] {
		case "M":
			row.IsMeta = true
		case "D":
			row.IsDependent = true
		}
	case m[3] != "":
		row.Target = m[3]
	}

	row.Description = cells[1]
	row.IsCI = strings.Contains(cells[2], "✔")
	row.IsLegacy = strings.Contains(cells[3], "✔")
	for _, vm := range reVarName.FindAllStringSubmatch(cells[4], -1) {
		row.Vars = append(row.Vars, vm[1])
	}
	return row, true
}

// splitMarkdownRow returns the trimmed cell contents of a markdown table
// row. Pipes inside backticks are not handled (the cheat sheet doesn't use
// them), keeping the parser straightforward.
func splitMarkdownRow(line string) []string {
	line = strings.TrimSpace(line)
	line = strings.TrimPrefix(line, "|")
	line = strings.TrimSuffix(line, "|")
	parts := strings.Split(line, "|")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}
	return parts
}
