---
applyTo: "internal/service/**/*.go"
description: "Code review guidelines for Terraform AWS Provider resources and datasources"
---

# Code Review Instructions for Resources and Datasources

## DONOTCOPY Comment Requirement

All newly created resources and datasources **must** include the following comment above the `package` declaration:

```go
// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.
```

### Identifying Resources and Datasources

A Go file is considered a resource or datasource if it contains any of the following annotations in comments:

- `@SDKResource` - SDK-based resource
- `@SDKDataSource` - SDK-based datasource
- `@FrameworkResource` - Framework-based resource
- `@FrameworkDataSource` - Framework-based datasource

### Example Structure

For a properly formatted resource file:

```go
// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package servicename

// ... imports ...

// @SDKResource("aws_service_resource_name", name="Resource Name")
func resourceResourceName() *schema.Resource {
    // ...
}
```

### Review Checklist

When reviewing new resource or datasource files, verify:

1. **DONOTCOPY comment is present** - The comment `// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.` must appear above the `package` declaration
2. **Comment placement** - The DONOTCOPY comment should appear after copyright/license headers but before the `package` statement

### Suggesting Changes

If the `DONOTCOPY` comment is missing from a new resource or datasource file, suggest the following change to the contributor:

**Suggestion message:**
> Missing required `DONOTCOPY` comment. Please add the following comment above the `package` declaration (after any copyright/license headers):
>
> ```go
> // DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.
> ```
>
> This comment is required for all new resources and datasources to discourage copying old patterns. Please use the `skaff` tool to scaffold new resources instead.

**Example fix:**

If the file looks like this:
```go
// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicename
```

Suggest changing it to:
```go
// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// DONOTCOPY: Copying old resources spreads bad habits. Use skaff instead.

package servicename
```


