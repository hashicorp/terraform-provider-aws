// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

// Command makelign validates that the Terraform AWS Provider's GNUmakefile
// stays in alignment with three pieces of supporting documentation:
//
//  1. The `.PHONY:` declaration inside GNUmakefile itself.
//  2. docs/makefile-cheat-sheet.md (the targets table).
//  3. docs/continuous-integration.md (CI target references).
//
// It reports two classes of issues:
//
//   - ERROR: hard alignment failures (.PHONY ghosts, missing .PHONY entries,
//     cheat sheet referring to nonexistent targets, CI flag mismatches).
//   - WARN:  soft alignment issues (documented target missing from the cheat
//     sheet, CI target not referenced in the CI doc, ordering).
//
// Errors fail the run with exit code 1. Warnings can be promoted to errors
// with `-strict`, which is the recommended setting for CI.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	defaultMakefile   = "GNUmakefile"
	defaultCheatSheet = "docs/makefile-cheat-sheet.md"
	defaultCIDoc      = "docs/continuous-integration.md"
)

// options bundles command-line flags so they can be threaded through the
// run function for testability without touching package globals.
type options struct {
	repoRoot string
	strict   bool
	noColor  bool
}

func main() {
	opts, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	code := run(opts, os.Stdout, os.Stderr)
	os.Exit(code)
}

// parseArgs builds an options struct from the supplied CLI arguments.
// Returning an error rather than calling os.Exit keeps the function
// usable from tests.
func parseArgs(args []string) (options, error) {
	fs := flag.NewFlagSet("makelign", flag.ContinueOnError)
	fs.SetOutput(io.Discard) // we render usage ourselves on error
	var opts options
	fs.BoolVar(&opts.strict, "strict", false, "treat warnings as errors")
	fs.BoolVar(&opts.noColor, "no-color", false, "disable ANSI color output")
	fs.Usage = func() {
		fmt.Fprint(os.Stderr, usage)
	}
	if err := fs.Parse(args); err != nil {
		return opts, fmt.Errorf("makelign: %w\n\n%s", err, usage)
	}
	switch fs.NArg() {
	case 0:
		opts.repoRoot = "."
	case 1:
		opts.repoRoot = fs.Arg(0)
	default:
		return opts, fmt.Errorf("makelign: unexpected extra arguments\n\n%s", usage)
	}
	return opts, nil
}

const usage = `Usage: makelign [flags] [repo-root]

Validates alignment between:
  - GNUmakefile target rules
  - GNUmakefile .PHONY list
  - docs/makefile-cheat-sheet.md
  - docs/continuous-integration.md

If repo-root is omitted, the current directory is used.

Flags:
  -strict      Treat warnings as errors (recommended for CI).
  -no-color    Disable ANSI color output.
  -h, -help    Show this help.

Exit codes:
  0  No errors (warnings allowed unless -strict).
  1  At least one error (or warning under -strict).
  2  Usage or I/O error.
`

// run executes one validation pass and returns a process exit code.
//
// stdout receives the human-readable findings table; stderr is reserved
// for I/O failures so that downstream tools parsing stdout aren't fooled
// by error messages.
func run(opts options, stdout, stderr io.Writer) int {
	makePath := filepath.Join(opts.repoRoot, defaultMakefile)
	sheetPath := filepath.Join(opts.repoRoot, defaultCheatSheet)
	ciPath := filepath.Join(opts.repoRoot, defaultCIDoc)

	mf, err := ParseMakefile(makePath)
	if err != nil {
		fmt.Fprintf(stderr, "makelign: cannot read %s: %v\n", makePath, err)
		return 2
	}
	sheet, err := ParseCheatSheet(sheetPath)
	if err != nil {
		fmt.Fprintf(stderr, "makelign: cannot read %s: %v\n", sheetPath, err)
		return 2
	}
	cidoc, err := ParseCIDoc(ciPath)
	if err != nil {
		fmt.Fprintf(stderr, "makelign: cannot read %s: %v\n", ciPath, err)
		return 2
	}

	findings := Validate(&Inputs{
		MakefilePath:   defaultMakefile,
		CheatSheetPath: defaultCheatSheet,
		CIDocPath:      defaultCIDoc,
		Make:           mf,
		CheatSheet:     sheet,
		CIDoc:          cidoc,
	})

	useColor := !opts.noColor && isTerminal(stdout)
	report(stdout, findings, useColor)

	errs, warns := countSeverity(findings)
	fmt.Fprintf(stdout, "\nSummary: %d error(s), %d warning(s)\n", errs, warns)

	switch {
	case errs > 0:
		return 1
	case opts.strict && warns > 0:
		return 1
	default:
		return 0
	}
}

// report renders findings in a fixed-width, grep-friendly format:
//
//	SEVERITY  file:line  message  (suggestion)
//
// We intentionally do not use a tabwriter so column positions are
// predictable in CI logs and editor jump-to-line features work.
func report(w io.Writer, findings []Finding, useColor bool) {
	if len(findings) == 0 {
		fmt.Fprintln(w, "makelign: no alignment issues found ✓")
		return
	}
	for _, f := range findings {
		sev := f.Severity.String()
		if useColor {
			sev = colorize(sev, f.Severity)
		}
		loc := f.File
		if f.Line > 0 {
			loc = fmt.Sprintf("%s:%d", f.File, f.Line)
		}
		if f.Suggestion != "" {
			fmt.Fprintf(w, "%-5s  %s  %s  (%s)  [%s]\n",
				sev, loc, f.Message, f.Suggestion, f.Code)
		} else {
			fmt.Fprintf(w, "%-5s  %s  %s  [%s]\n",
				sev, loc, f.Message, f.Code)
		}
	}
}

func countSeverity(findings []Finding) (errs, warns int) {
	for _, f := range findings {
		switch f.Severity {
		case SevError:
			errs++
		case SevWarn:
			warns++
		}
	}
	return errs, warns
}

// colorize adds ANSI escapes for terminal output. We only color the
// severity tag so log scrubbing tools see plain message text.
func colorize(s string, sev Severity) string {
	switch sev {
	case SevError:
		return "\x1b[31m" + s + "\x1b[0m" // red
	case SevWarn:
		return "\x1b[33m" + s + "\x1b[0m" // yellow
	default:
		return s
	}
}

// isTerminal returns true when w looks like a TTY. We avoid pulling in
// golang.org/x/term to keep the module dependency-free.
func isTerminal(w io.Writer) bool {
	f, ok := w.(*os.File)
	if !ok {
		return false
	}
	st, err := f.Stat()
	if err != nil {
		return false
	}
	return st.Mode()&os.ModeCharDevice != 0
}
