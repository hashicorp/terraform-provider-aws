# smarterr Quick Start & Migration Guide

`smarterr` is a config-driven library for formatting and annotating Terraform AWS Provider errors in a consistent, helpful, and composable way. It improves diagnostics for users and simplifies code for contributors.

## What smarterr Does

smarterr consumes:

* Configuration files (`smarterr.hcl`) with tokens, templates, transforms, etc.
* Go error values (wrapped with context-aware stack info)
* `context.Context`
* Optional key/value arguments (`keyvals`)
* Live and captured call stack frames

It then produces consistent diagnostics with clear user messages across:

* SDKv2-style diagnostics (`sdkdiag.Diagnostics`)
* Framework-style diagnostics (`fwdiag.Diagnostics`)
* Logs (if configured)

---

## `smerr`: Provider Integration

The `internal/smerr` package is a thin wrapper over smarterr that:

* Injects private provider context (e.g., resource name, service name)
* Simplifies usage for both SDKv2 and Framework-style resources

Use `smerr` in most cases instead of importing smarterr directly.

---

## smarterr Configuration

Located in:

* `internal/smarterr.hcl` – global config
* `internal/service/<service>/smarterr.hcl` – service-specific overrides

These files support:

* `token` – building blocks like `id`, `error`, or `happening`
* `template` – Go `text/template` blocks for message formatting
* `parameter` – static values like service name
* `stack_match` – extract info from call stack
* `transform` – text cleanup utilities
* `smarterr` – optional global configuration

See [the smarterr docs](https://github.com/YakDriver/smarterr/tree/main/docs) for more.

---

## Migration Guide

### Step 1: Determine Resource Style

#### Framework-style

* Uses `"github.com/hashicorp/terraform-plugin-framework/resource"`
* Has `@FrameworkResource` or `@FrameworkDataSource` comments
* Defines methods like `Schema`, `Create`, `Read`, `Update`, `Delete`

#### SDKv2-style

* Uses `"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"`
* Has `@SDKResource` comment and returns `*schema.Resource`

---

### SDKv2 Migration Steps

1. **Replace diagnostic append calls**
   Replace any of the following:

   * `sdkdiag.AppendErrorf`
   * `sdkdiag.AppendFromErr`
   * `create.AppendDiagError`

   **Before:**

   ```go
   return sdkdiag.AppendErrorf(diags, "creating (%s): %s", id, err)
   ```

   **After:**

   ```go
   return smerr.Append(ctx, diags, err, smerr.ID, id)
   ```

2. **Wrap all error return values**
   **Before:**

   ```go
   return nil, err
   ```

   **After:**

   ```go
   return nil, smarterr.NewError(err)
   ```

   For helpers like `AssertSingleValueResult`:

   ```go
   return smarterr.Assert(tfresource.AssertSingleValueResult(...))
   ```

---

### Framework Migration Steps

1. **Replace `AddError`**
   **Before:**

   ```go
   resp.Diagnostics.AddError("creating...", err.Error())
   ```

   **After:**

   ```go
   smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, state.RuleName.String())
   ```

2. **Replace `Append`**
   If **no ID available** (e.g., early in `Create`):

   ```go
   smerr.EnrichAppend(ctx, &resp.Diagnostics, req.State.Get(ctx, &state))
   ```

   **Important!** Use **ID if available**, even if it wasn't in the original `Append` call:

   ```go
   smerr.EnrichAppend(ctx, &resp.Diagnostics, fwflex.Flatten(ctx, out, &state), smerr.ID, state.RuleName.String())
   ```

   For example, replace (ID not available, early in function):

   ```go
   resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
   ```

   With:

   ```go
   smerr.EnrichAppend(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))
   ```

   For example, replace (ID _is_ available, middle to end of function):

   ```go
   resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
   ```

   With:

   ```go
   smerr.EnrichAppend(ctx, &resp.Diagnostics, resp.State.Set(ctx, &state), smerr.ID, state.ResourceArn.String())
   ```

3. **Wrap error-producing calls**
   Same as SDKv2:

   ```go
   return nil, smarterr.NewError(err)
   ```

---

## Why Wrap Errors?

Wrapping errors with `smarterr.NewError()` captures call stack information at the time of failure. This enables smarterr to:

* Determine **subaction** like "finding", "waiting", etc.
* Avoid duplicative wrapping (“walls of text”)
* Format the **summary** and **detail** portions idiomatically

---

## AI/Human Migration Tip

Use this migration checklist:

* ✅ Wrapped all `return nil, err` with `smarterr.NewError()`
* ✅ Replaced `AppendErrorf()` and friends with `smerr.Append()`
* ✅ Replaced `AddError()` with `smerr.AddError()`
* ✅ Replaced `Append()` with `smerr.EnrichAppend()`
* ✅ Preserved `ctx`, `diags`, and `id` correctly
* ✅ Did not modify resource schema or unrelated logic
