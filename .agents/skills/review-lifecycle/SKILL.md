---
name: review-lifecycle
description: "Review Terraform AWS Provider resource lifecycle: CRUD step order, smerr/smarterr error handling, and AutoFlex model-to-SDK conversion. Use when reviewing a PR that changes non-test resource logic in internal/service/**/*.go (Create/Read/Update/Delete, error wrapping, flex.Expand/Flatten/Diff)."
---

<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# Review: Resource Lifecycle

Assume the `@maintainer` persona. Scope: CRUD, errors, AutoFlex on non-test resource code (`internal/service/**/*.go`). Loaded from `review-pr`.

## Registration annotations

One above each factory function. Flag missing annotations or factory-name/resource-name mismatches.

| Construct | Annotation |
|---|---|
| Resource (Framework) | `// @FrameworkResource("aws_<svc>_<thing>", name="<Human Name>")` |
| Data source (Framework) | `// @FrameworkDataSource("aws_<svc>_<thing>", name="<Human Name>")` |
| List resource (Framework) | `// @FrameworkListResource("aws_<svc>_<thing>")` |
| List resource (SDKv2) | `// @SDKListResource("aws_<svc>_<thing>")` |

## Struct embeds

```go
type fooResource struct {
    framework.ResourceWithModel[fooResourceModel]
    framework.WithTimeouts          // only when a timeouts block is present
    framework.WithImportByIdentity  // not WithImportByID for new resources
}
```

- New resources use `framework.WithImportByIdentity`. Flag `framework.WithImportByID` on new resources.
- Use `framework.ResourceWithModel[T]` (not `ResourceWithConfigure`).
- Data sources: `framework.DataSourceWithModel[T]`.

## CRUD step order

**Create:** get client → fetch plan → `flex.Expand` → set tags via `input.Tags = getTagsIn(ctx)` for tagged resources → AWS Create call → `flex.Flatten` output back into the plan → wait → `resp.State.Set(ctx, plan)`. Flag Creates that don't read the output back into the plan — computed attributes stay unknown in state.

**Read:** get client → fetch state → finder → on `retry.NotFound(err)`, append `fwdiag.NewResourceNotFoundWarningDiagnostic(err)`, call `resp.State.RemoveResource(ctx)`, return → flatten → `resp.State.Set`.

**Update:** get client → fetch plan and state → `diff, d := flex.Diff(ctx, plan, state)` → gate on `diff.HasChanges()` → AWS modify → flatten output back into plan → wait → `resp.State.Set(ctx, &plan)`. Flag updates that always call the API without `HasChanges()`, or that re-fetch state after the modify.

Omit Update when the API has no update, every attribute has `RequiresReplace()`, or Create is reused for modify.

**Delete:** get client → fetch state → build input → AWS delete; silently swallow `errs.IsA[*awstypes.ResourceNotFoundException](err)` → wait.

## Errors

Use `smerr`/`smarterr`, never raw `resp.Diagnostics.AddError`.

- Upstream diagnostics: `smerr.AddEnrich(ctx, &resp.Diagnostics, req.Plan.Get(ctx, &plan))` (not the deprecated `EnrichAppend`).
- API errors: `smerr.AddError(ctx, &resp.Diagnostics, err, smerr.ID, plan.Name.String())`. Always pass `smerr.ID` with an identifier — `plan.Name` in Create, `state.ID` elsewhere.
- Inside finders/waiters, wrap returned errors with `smarterr.NewError(err)`.
- Detect AWS exceptions with `errs.IsA[*awstypes.<Exception>](err)` — never type assertions, never `strings.Contains(err.Error(), ...)`.
- Use `retry.NotFound(err)` for wrapped not-found from the finder layer.

Flag raw `resp.Diagnostics.AddError`, type assertions on AWS errors, and string-based error matching.

## AutoFlex

Use `flex.Expand` / `flex.Flatten` for model ↔ SDK conversion. Manual per-field `aws.String` / `aws.ToString` is wrong in new Framework code.

- `flex.WithFieldNamePrefix("<Thing>")` when AWS prefixes its fields (model `ID` ↔ SDK `ThingId`).
- AWS plural collections become singular blocks (`Parameters` ↔ `parameter`).
- AutoFlex does **not** copy tags through; see `review-tags`.
- Update path: `flex.Diff(ctx, plan, state)` then `diff.HasChanges()`.
