---
applyTo: "internal/service/**/*.go"
---
<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# Helpers, Sweepers, Data Sources, List Resources

Scope: finders, status, waiters, sweepers, data sources, list resources.

## Finders, status, waiters

- Finder signature: `find<Resource>ByID(ctx context.Context, conn *<svc>.Client, id string) (*awstypes.<Type>, error)`. Variants by ARN/Name use the same shape.
- On `*awstypes.ResourceNotFoundException`, return `smarterr.NewError(&retry.NotFoundError{LastError: err})`.
- Status function reuses the finder and returns `retry.StateRefreshFunc`. Design status so one function powers create, update, and delete waiters.
- Waiters use `retry.StateChangeConf`. Created/updated waiters typically set `NotFoundChecks: 20` and `ContinuousTargetOccurence: 2`. Deleted waiters use empty `Target` with `Pending` covering deletion-in-progress states.
- Prefer SDK-provided status constants (e.g. `awstypes.StatusInProgress`) over package-level string consts.
- Finders and `ResName<Name>` constants referenced in tests must be re-exported via `exports_test.go`.

Flag finders that return raw errors (must wrap with `smarterr.NewError`), status that duplicates finder logic, or hand-rolled polling loops in place of `retry.StateChangeConf`.

## Sweepers

Each new resource needs a sweeper. Iterate the SDK paginator, build via `framework.NewSweepResource(new<Resource>Resource, client, framework.NewAttribute(names.AttrID, aws.ToString(v.<Thing>Id)))` (where `framework` is `internal/sweep/framework`), and register in the package's `sweep.go` with `awsv2.Register("aws_<svc>_<thing>", sweep<Resource>s, ...optionalDeps)`. Pass multiple `framework.NewAttribute(...)` arguments for composite identity.

Flag new resources without a sweeper, sweepers that don't propagate paginator errors via `smarterr.NewError`, and sweepers using import aliases other than `framework` for `internal/sweep/framework`.

## Data sources

Data sources have only a `Read` method.

- Use `framework.DataSourceWithModel[T]`.
- Schema attributes are `Required` or `Optional` for search criteria; everything else is `Computed`.
- Attributes that are `Required` on the corresponding resource are typically `Computed` on the data source unless they form lookup criteria.
- No configurable timeouts.
- Tagged data sources expose a single computed `tags` attribute (no `tags_all`).

## List resources

Framework path embeds the corresponding underlying resource. SDKv2 path uses `framework.ListResourceWithSDKv2Resource`.

The `List` method:

1. Get the client.
2. Fetch the config (only when the list takes query attributes such as a parent ID).
3. Stream results from a paginated AWS List API.
4. Set logging fields per item (typically the ARN) via `tflog.SetField(ctx, logging.ResourceAttributeKey(...), ...)`.
5. Set identifying attributes for each result.
6. Set `result.DisplayName` to a human-readable identifier (typically the resource name).

The listing helper uses an iterator pattern over the SDK paginator:

```go
func list<Thing>s(ctx context.Context, conn *<svc>.Client, input *<svc>.List<Thing>sInput) iter.Seq2[awstypes.<Thing>, error]
```

The flatten function shared by Read and List lives in the resource file (`r.flatten` for Framework, `resource<Name>Flatten` for SDKv2).

Flag list resources that don't set `DisplayName`, that re-implement flatten logic instead of sharing with Read, or that omit the `tflog.SetField` per-item logging hook.
