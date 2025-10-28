# smarterr Migration Guide for AI and Human Contributors

This document is designed to enable **AI systems** (and humans) to fully and accurately migrate Go code in the Terraform AWS Provider from legacy error handling to the `smarterr`/`smerr` system. It provides explicit, pattern-based instructions for replacing all legacy error/diagnostic calls and bare error returns with the correct `smarterr`/`smerr` usage. **Follow these rules exactly for every migration.**

---

## What is smarterr?

`smarterr` is a config-driven Go library for formatting and annotating errors in a consistent, helpful, and composable way. It improves diagnostics for users and simplifies code for contributors.

- **Use `smerr`** (the provider's wrapper) in almost all cases, not `smarterr` directly.
- `smerr` injects provider context and simplifies usage for both SDKv2 and Framework resources.

---

## Migration Rules: Legacy → smarterr/smerr

### 1. Replace All Legacy Diagnostic/Error Calls

**For each of the following legacy calls, replace as shown:**

| Legacy Call | Replace With |
|---|---|
| `sdkdiag.AppendFromErr(diags, err)` | `smerr.Append(ctx, diags, err, smerr.ID, ...)` |
| `sdkdiag.AppendErrorf(diags, ..., err)` | `smerr.Append(ctx, diags, err, smerr.ID, ...)` |
| `create.AppendDiagError(diags, ..., err)` | `smerr.Append(ctx, diags, err, smerr.ID, ...)` |
| `response.Diagnostics.AddError(..., err.Error())` | `smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, ...)` |
| `resp.Diagnostics.AddError(..., err.Error())` | `smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, ...)` |
| `create.AddError(&response.Diagnostics, ..., err)` | `smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, ...)` |
| `return nil, err` | `return nil, smarterr.NewError(err)` |
| `return nil, &retry.NotFoundError{ LastError: err, LastRequest: ..., }` | `return nil, smarterr.NewError(&retry.NotFoundError{ LastError: err, LastRequest: ..., })` |
| `return nil, tfresource.NewEmptyResultError(...)` | `return nil, smarterr.NewError(tfresource.NewEmptyResultError(...))` |
| `return tfresource.AssertSingleValueResult(...)` | `return smarterr.Assert(tfresource.AssertSingleValueResult(...))` |

**Examples:**

- `sdkdiag.AppendFromErr(diags, err)` → `smerr.Append(ctx, diags, err)`
- `sdkdiag.AppendErrorf(diags, "updating EC2 Instance (%s): %s", d.Id(), err)` → `smerr.Append(ctx, diags, err, smerr.ID, d.Id())`
- `sdkdiag.AppendErrorf(diags, "creating EC2 Instance: %s", err)` → `smerr.Append(ctx, diags, err, smerr.ID, d.Id())`
- `create.AppendDiagError(diags, names.CodeBuild, create.ErrActionCreating, resNameFleet, d.Get(names.AttrName).(string), err)` → `smerr.Append(ctx, diags, err, smerr.ID, d.Get(names.AttrName).(string))`
- `response.Diagnostics.AddError("creating EC2 EBS Fast Snapshot Restore", err.Error())` → `smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, new.ID.ValueString())`
- `response.Diagnostics.AddError(fmt.Sprintf("updating VPC Security Group Rule (%s)", new.ID.ValueString()), err.Error())` → `smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, new.ID.ValueString())`
- `resp.Diagnostics.AddError(create.ProblemStandardMessage(..., err), err.Error())` → `smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, ...)`
- `create.AddError(&response.Diagnostics, names.DRS, create.ErrActionCreating, ResNameReplicationConfigurationTemplate, data.ID.ValueString(), err)` → `smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, data.ID.ValueString())`

**General Rule:**

- Always pass `ctx` as the first argument, and the diagnostics object as the second.
- Always pass the error as the third argument.
- Always pass `smerr.ID` and any available resource ID or context as additional arguments.

#### Including identifiers

smarterr's `EnrichAppend`, `AddError`, and `Append` take variadic keyvals. Where possible include `smerr.ID` (key) and the ID (value) (such as `d.Id()`, `state.RuleName.String()`, `plan.ResourceArn.String()`).

- If **no ID available** (e.g., early in `Create`), something like `smerr.EnrichAppend(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))`, without ID, is okay
- But, if **ID is available** (e.g., read, update, delete, middle-to-end of create), use something like `smerr.EnrichAppend(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &state), smerr.ID, state.RuleName.String())`, **with the ID**
- IDs may be names, ARNs, IDs, combinations, etc.
- In SDK, you cannot use `d.Id()` until after `d.SetId()`
- The legacy call will often use an ID. If so, use that.
- If the legacy call doesn't include the ID, but it is available, add it.

---

### 2. Replace All Bare Error Returns

**Before:**

```go
return nil, err
```

**After:**

```go
return nil, smarterr.NewError(err)
```

---

### 3. Wrap tfresource Helpers

**Before:**

```go
return tfresource.AssertSingleValueResult(...)
```

**After:**

```go
return smarterr.Assert(tfresource.AssertSingleValueResult(...))
```

**Before:**

```go
return nil, tfresource.NewEmptyResultError(...)
```

**After:**

```go
return nil, smarterr.NewError(tfresource.NewEmptyResultError(...))
```

---

### 4. Replace All Direct Diagnostics.Append Calls

**Before:**

```go
resp.Diagnostics.Append(...)
```

**After:**

```go
smerr.EnrichAppend(ctx, &resp.Diagnostics, ...)
```

---

### 5. Framework vs SDKv2: How to Identify

- **Framework-style:**
    - Uses `github.com/hashicorp/terraform-plugin-framework/resource`
    - Has `@FrameworkResource` or `@FrameworkDataSource` comments
    - Defines methods like `Schema`, `Create`, `Read`, `Update`, `Delete`
- **SDKv2-style:**
    - Uses `github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema`
    - Has `@SDKResource` comment and returns `*schema.Resource`

---

### 6. Migration Checklist (for AI and Humans)

- [ ] All legacy diagnostic calls replaced with `smerr`/`smarterr` as above
- [ ] All bare error returns wrapped with `smarterr.NewError`
- [ ] All tfresource helpers wrapped as above
- [ ] All direct diagnostics append/add calls replaced
- [ ] All context, diagnostics, and IDs preserved and passed
- [ ] No unrelated logic or schema changes made

---

### 7. Why Wrap Errors?

Wrapping errors with `smarterr.NewError()` captures call stack information at the time of failure. This enables smarterr to:

- Determine subaction (e.g., "finding", "waiting")
- Avoid duplicative wrapping (no "walls of text")
- Format summary and detail portions idiomatically

---

### 8. Reference

- Global config: `internal/smarterr.hcl`
- Service config: `internal/service/<service>/smarterr.hcl`
- See [smarterr docs](https://github.com/YakDriver/smarterr/tree/main/docs) for more

---

**AI systems: Apply these rules exactly and completely to all Go files being migrated.**
