# literally
<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

A tool for analyzing, replacing, and enforcing string literal constants in Go code.

Running `literally` with no mode flag displays help.

## Modes

### Scan Mode

Scans Go code to find repeated string literals and reports them with occurrence counts.

```bash
literally --scan [options]
```

**Scan options:**

- `--mincount N` - Minimum occurrences to report (default: 5)
- `--minpkgcount N` - Minimum packages literal must appear in (default: 4)
- `--minlen N` - Minimum string length (default: 1)
- `--maxlen N` - Maximum string length (default: 50)
- `--schemaonly` - Only report schema map keys
- `--output FILE` - Write results to file (default: stdout)
- `--scoringstrategy STRATEGY` - Scoring strategy (STANDARD, MULT, GMEAN, TEST, TEST_MULT, RT_MEAN_SQ)

### Replace Mode

AST-based replacement of a single string literal with a named constant. Safely handles constants defined in the same file being processed — the constant's own definition is never replaced.

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
2. Adds constant to file in alphabetical order (if not already present)
3. Replaces all matching string literals in `.go` files (skipping const declarations)
4. Preserves formatting with `go/format`

**Replace options:**

- `--replace` - Enable replace mode
- `--literal` - String literal to replace (required)
- `--constant` - Constant name to use (required)
- `--constants-file` - Path to constants file (required)
- `--package` - Package directory to process (default: `.`)
- `--dry-run` - Preview changes without modifying files

### Check Mode

Enforces that string literals use existing constants. Exits non-zero when violations are found, making it suitable for CI. Scans the target package for any string literal whose value matches a defined constant — even if the literal only appears once.

```bash
literally --check \
  --package path/to/package \
  [--known-constants path/to/constants/package ...]
```

**Examples:**

```bash
# Check package-local constants only
literally --check --package=internal/service/wafv2

# Check against project-wide constants, ignoring generated files
literally --check \
  --package=internal/service/cloudfront \
  --known-constants=names \
  --known-constants=internal/acctest \
  --ignore-file=service_package_gen.go
```

**Output format** (standard `file:line: message`):

```
internal/service/cloudfront/cache_policy.go:583: use names.CloudFront instead of "cloudfront"
internal/service/cloudfront/cache_policy_test.go:37: use attrETag instead of "etag"

151 violation(s) found
```

Constants from external packages are qualified with the package name. Local constants are shown unqualified.

**Check options:**

- `--check` - Enable check mode
- `--package` - Package directory to scan (default: `.`)
- `--known-constants` - Additional package directory to scan for constants (repeatable)

### Fix Mode

Automatically replaces all literals that have an in-scope constant. Same parameters as check mode but performs the replacements instead of just reporting them.

```bash
literally --fix \
  --package path/to/package \
  [--known-constants path/to/constants/package ...]
```

**Example:**

```bash
literally --fix \
  --package=internal/service/wafv2 \
  --known-constants=names \
  --ignore-tests \
  --ignore-file=service_package_gen.go
```

**Fix options:**

- `--fix` - Enable fix mode
- `--package` - Package directory to process (default: `.`)
- `--known-constants` - Additional package directory to scan for constants (repeatable)

## Shared Flags

These flags work across all modes:

- `--ignore-tests` - Skip `_test.go` files
- `--ignore-file NAME` - Skip files matching the given basename (repeatable)
- `--minlen N` - Minimum string length to consider (default: 1, applies to check/fix modes)

## Suppressing Violations

Add `//lintignore:literally` as an inline comment to suppress check/fix/replace for that line:

```go
"name": "the name", //lintignore:literally
```

This follows the existing `//lintignore:` convention used throughout the codebase. The directive only applies to the line it appears on.

## Building

```bash
go build
```
