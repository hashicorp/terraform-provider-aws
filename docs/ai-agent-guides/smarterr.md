<!-- Copyright IBM Corp. 2014, 2025 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# smarterr Migration Guide

**AI and Human Contributors:** Follow these exact patterns to migrate Terraform AWS Provider error handling to smarterr/smerr.

## Core Concept

- **Use `smerr`** (provider wrapper) for all diagnostic calls
- **Use `smarterr`** only for bare error returns
- **Framework uses "Add" verbs, SDKv2 uses "Append" verbs**

## Migration Patterns

### 1. Legacy Diagnostic Calls → smerr

| Legacy | Replace With |
|--------|-------------|
| `sdkdiag.AppendFromErr(diags, err)` | `smerr.Append(ctx, diags, err)` |
| `sdkdiag.AppendErrorf(diags, "msg", err)` | `smerr.Append(ctx, diags, err, smerr.ID, id)` |
| `response.Diagnostics.AddError("msg", err.Error())` | `smerr.AddError(ctx, &response.Diagnostics, err)` |
| `create.AppendDiagError(diags, ..., err)` | `smerr.Append(ctx, diags, err, smerr.ID, id)` |

### 2. Bare Error Returns → smarterr

```go
// Before
return nil, err

// After
return nil, smarterr.NewError(err)
```

### 3. tfresource Helpers → smarterr

```go
// Before
return tfresource.AssertSingleValueResult(...)

// After
return smarterr.Assert(tfresource.AssertSingleValueResult(...))
```

### 4. Direct Diagnostics Calls

**Framework:**

```go
// Before: resp.Diagnostics.Append(...)
// After:  smerr.AddEnrich(ctx, &resp.Diagnostics, ...)
```

**SDKv2:**

```go
// Before: return append(diags, someFunc()...)
// After:  return smerr.AppendEnrich(ctx, diags, someFunc())
```

## Function Reference

**Framework (Add verbs):**

- `smerr.AddError(ctx, &response.Diagnostics, err, smerr.ID, id)`
- `smerr.AddEnrich(ctx, &response.Diagnostics, diagnosticFunc())`

**SDKv2 (Append verbs):**

- `smerr.Append(ctx, diags, err, smerr.ID, id)`
- `smerr.AppendEnrich(ctx, diags, diagnosticFunc())`

**Both contexts:**

- `smarterr.NewError(err)` - Wrap bare returns
- `smarterr.Assert(tfresource.AssertSingleValueResult(...))` - Wrap helpers

## Key Rules

1. **Always pass `ctx` first**
2. **Always include `smerr.ID, resourceID` when available** (e.g., `d.Id()`, `state.Name.String()`)
3. **Framework = Add verbs, SDKv2 = Append verbs**
4. **Preserve all existing IDs and context**

## Identify Framework vs SDKv2

- **Framework:** `@FrameworkResource`, uses `terraform-plugin-framework`
- **SDKv2:** `@SDKResource`, uses `terraform-plugin-sdk/v2`, returns `*schema.Resource`

---

**Apply these patterns exactly. No schema or logic changes.**
