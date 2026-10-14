# sdkmap
<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

A tool that loads each service package in `internal/service/*` and reports every resource annotated with `@SDKResource`, along with its CRUD, `CustomizeDiff`, and custom import functions and the AWS SDK client methods it calls.

Annotations are recognized using the same rules as the service package generator (`internal/generate/servicepackage`):

- Only the top-level package of each service directory is scanned (subpackages and `test-fixtures` are not).
- Annotated methods are ignored.
- The whole comment line must be an annotation, and the resource type name must be valid (`aws_...`).

Build constraints are honored when selecting a package's files.

## Usage

From the repository root:

```bash
go run tools/sdkmap/main.go
```

**Options:**

- `-root DIR` - Path to the root of the repository (default: `.`)

Output is tab-separated, with paths relative to the repository root. Each resource line is followed by indented lines for the functions it references, then for the AWS SDK client methods it calls:

```
<service>	<resource type name>	<factory function>	<file>:<line>
	<Create|Read|Update|Delete|CustomizeDiff|Import>	<reference>	<file>:<line>
	SDKCall	<sdk package>.<method>
```

Function references:

- `CustomizeDiff` functions combined with `customdiff.All` or `customdiff.Sequence` are listed individually.
- `Import` is listed only for custom importers, not `schema.ImportStatePassthroughContext`.
- The location is the function's package-level declaration (for a call, the called function's), or the function literal or call itself. It is empty for references outside the package, e.g. `verify.SetTagsDiff` or `schema.NoopContext`, and for method values.

SDK calls:

- `SDKCall` lists, once each, the AWS SDK client methods reachable from the resource's factory function. The prefix is the SDK package name, which can differ from the service, e.g. `applicationautoscaling` for `aws_appautoscaling_*`.
- References are followed through package-level functions, variables, and constants; exported declarations in other service packages (e.g. `tfec2.FindSubnetByID`); and methods called on a value whose type can be determined from the source. A call on an interface type follows every type in the package that implements it.
- A client is a variable, parameter, or struct field of type `*<sdk package>.Client`, or the result of a `conns.AWSClient` client method (e.g. `S3Client(ctx)`). Paginators (e.g. `s3.NewListObjectsV2Paginator`) count as calls to their operation.
- For resources with `@Tags(identifierAttribute=...)`, the calls made by the tags interceptor through the service package's `ListTags` and `UpdateTags` are included. Where these are shared by several resource types (e.g. in S3), every resource includes the calls for all of them.
- The analysis is syntactic, so some calls are missed: those made through SDK waiters, through non-service packages (e.g. `manager.NewUploader` in S3), or on values whose type cannot be determined from the source. Variable scoping is ignored.

A resource with no SDK calls found produces a warning on stderr. A total count is written to stderr.

## Building

```bash
go build
```
