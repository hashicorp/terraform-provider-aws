# sdkmap
<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

A tool that loads each service package in `internal/service/*` and reports every resource annotated with `@SDKResource`, along with its CRUD, `CustomizeDiff`, and custom import functions.

Annotations are recognized using the same rules as the service package generator (`internal/generate/servicepackage`):

- Only the top-level package of each service directory is scanned (subpackages and `test-fixtures` are not).
- Methods are ignored.
- The whole comment line must be an annotation, and the resource type name must be valid (`aws_...`).

Build constraints are honored when selecting a package's files.

## Usage

From the repository root:

```bash
go run tools/sdkmap/main.go
```

**Options:**

- `-root DIR` - Path to the root of the repository (default: `.`)

Output is tab-separated, with paths relative to the repository root. Each resource is followed by an indented line for each function it references:

```
<service>	<resource type name>	<factory function>	<file>:<line>
	<Create|Read|Update|Delete|CustomizeDiff|Import>	<reference>	<file>:<line>
```

- `CustomizeDiff` functions combined with `customdiff.All` or `customdiff.Sequence` are listed individually.
- `Import` is listed only for custom importers, not `schema.ImportStatePassthroughContext`.
- The location is the function's package-level declaration (for a call, the called function's), or the function literal or call itself. It is empty for references outside the package, e.g. `verify.SetTagsDiff` or `schema.NoopContext`, and for method values.

A total count is written to stderr.

## Building

```bash
go build
```
