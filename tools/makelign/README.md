<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# makelign

`makelign` validates that the Terraform AWS Provider's `GNUmakefile` stays in
alignment with its supporting documentation. It is intentionally narrow: it
parses four sources, applies a fixed set of rules, and prints findings.

## What it checks

| # | Source                                       | Source              |
|---|----------------------------------------------|---------------------|
| 1 | Target rules                                 | `GNUmakefile`       |
| 2 | `.PHONY:` declaration                        | `GNUmakefile`       |
| 3 | Targets table                                | `docs/makefile-cheat-sheet.md` |
| 4 | `make <target>` references                   | `docs/continuous-integration.md` |

## Rules

Each finding carries a stable code so they can be filtered or suppressed at
the report layer if necessary.

### Errors (fail CI)

| Code              | Meaning |
|-------------------|---------|
| `phony-missing`   | A target has a recipe in `GNUmakefile` but is absent from the `.PHONY` list. Without `.PHONY`, a same-named file in the working directory would silently shadow the target. |
| `phony-ghost`     | A name appears in `.PHONY` but no rule defines it. Includes a "did you mean" suggestion via Levenshtein distance. |
| `cheatsheet-ghost`| The cheat sheet documents a target that does not exist in `GNUmakefile`. |
| `ci-flag-mismatch`| The `## [CI]` prefix on the target description disagrees with the ✔ in the cheat sheet's `CI?` column. |

### Warnings (informational, fail under `-strict`)

| Code                    | Meaning |
|-------------------------|---------|
| `cheatsheet-missing`    | A target with a `## description` comment is not listed in the cheat sheet. |
| `ci-doc-missing`        | A `## [CI]` target is not referenced in `continuous-integration.md`. Meta targets (marked `<sup>M</sup>` in the cheat sheet) are exempt. |
| `legacy-flag-mismatch`  | The Makefile description mentions "Legacy" but the cheat sheet's `Legacy?` column disagrees, or vice versa. |
| `phony-order`           | The `.PHONY` list is not alphabetized. |
| `cheatsheet-order`      | The cheat sheet table is not alphabetized. |

## Usage

From the repository root:

```console
$ make makefile-lint
```

Or directly:

```console
$ cd tools/makelign
$ go run . ../..
```

### Flags

```text
-strict      Treat warnings as errors (recommended for CI).
-no-color    Disable ANSI color output.
```

### Exit codes

| Code | Meaning |
|------|---------|
| 0    | No errors (warnings allowed unless `-strict`). |
| 1    | At least one error (or warning under `-strict`). |
| 2    | Usage or I/O error. |

## Output format

```text
SEVERITY  file:line  message  (suggestion)  [code]
```

Example:

```text
ERROR  GNUmakefile:1193  .PHONY entry "gh-workflows-lint" does not match any target  (did you mean "gh-workflow-lint"?)  [phony-ghost]
ERROR  GNUmakefile:259   target "copyright-fix" is not in the .PHONY list  [phony-missing]
WARN   GNUmakefile:165   documented target "clean-go-cache-trim" is missing from the cheat sheet  [cheatsheet-missing]

Summary: 2 error(s), 1 warning(s)
```

The columns are space-separated and grep-friendly. Editors that recognize
`file:line` jump syntax (vim, VS Code, IntelliJ) navigate directly to each
finding.

## Design choices

* **Stdlib only.** No third-party dependencies. The parsers are line-oriented
  and tolerant: malformed Make syntax produces nothing rather than a fatal
  error so an in-progress edit can still be linted.
* **Single binary, no subcommands.** This tool does one thing.
* **Severity, not categories.** Errors block; warnings nudge. Anything more
  granular tends to drift toward never being acted on.
* **Suggestions where useful.** Levenshtein-based "did you mean" hints turn
  most ghost-PHONY findings into a one-character fix.
* **Stable finding codes.** Codes such as `phony-ghost` are part of the
  contract: they're greppable in CI logs and easy to reference in PRs.

## Testing

```console
$ go test ./...
```

The fixtures under `testdata/golden/` exercise every rule. Add a new fixture
when adding a new rule rather than mutating the existing one.

## Adding a rule

1. Add a check function to `validate.go` returning `[]Finding`.
2. Wire it into `Validate`.
3. Document it in this README's rules table.
4. Add a fixture or extend `testdata/golden/` and assert the finding from
   `TestValidate_Golden` (or, preferably, `TestValidate_RulesIsolated`).
