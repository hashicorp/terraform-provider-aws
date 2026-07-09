# literally
<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

A tool for analyzing and replacing string literals in Go code.

## Modes

### Analysis Mode (default)

Scans Go code to find repeated string literals and reports them with occurrence counts.

```bash
literally [options]
```

**Common options:**

- `--mincount N` - Minimum occurrences to report (default: 5)
- `--minpkgcount N` - Minimum packages literal must appear in (default: 4)
- `--minlen N` - Minimum string length (default: 1)
- `--maxlen N` - Maximum string length (default: 50)
- `--schemaonly` - Only report schema map keys
- `--output FILE` - Write results to file (default: stdout)

### Replace Mode

AST-based replacement of string literals with constants. Only replaces actual Go string literals, not struct tags or import paths.

```bash
literally --replace \
  --literal "string_to_replace" \
  --constant constantName \
  --constants-file path/to/attr_names.go \
  --package path/to/package \
  [--dry-run]
```

**Example:**

```bash
literally --replace \
  --literal "checksum" \
  --constant attrChecksum \
  --constants-file internal/service/lexmodels/attr_names.go \
  --package internal/service/lexmodels \
  --dry-run
```

**What it does:**

1. Creates constants file if it doesn't exist
2. Adds constant to file in alphabetical order (if not present)
3. Replaces all matching string literals in non-test `.go` files
4. Preserves formatting with `go/format`

**Replace options:**

- `--replace` - Enable replace mode
- `--literal` - String literal to replace (required)
- `--constant` - Constant name to use (required)
- `--constants-file` - Path to constants file (required)
- `--package` - Package directory to process (default: `.`)
- `--dry-run` - Preview changes without modifying files

## Building

```bash
go build
```
