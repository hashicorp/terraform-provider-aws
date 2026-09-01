# Design: Surfacing RDS Events (FRB-7061)
<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

Status: Draft (delete after implementation). **Component A (§4, §5) — DONE,
verified 2026-09-01: `TestAccRDSEventsDataSource_basic` passed
(`go test ./internal/service/rds/... -run TestAccRDSEventsDataSource_basic`,
437.49s); no leftover AWS resources confirmed via
`describe-db-instances`.** Component B (§6) not started.
Tracking: FRB-7061; GitHub #41037 (open enhancement); #29861 (closed/fixed — do not reopen)

> Sections 1–9 are the design to implement. Rationale, resolved decisions,
> open questions, and correction history live in the Appendices.

---

## 1. Problem

When RDS rejects or silently defers a modify — most visibly a major/minor
`engine_version` upgrade that fails a pre-check, or enhanced monitoring that
fails to configure — the AWS API returns **no error**. The reason is recorded
only in the RDS **event stream** (`DescribeEvents`). The provider never reads
those events, so the user sees either the confusing `Provider produced
inconsistent final plan` error (#29861) or a silent success followed by a
perpetual diff (#41037).

The plan-consistency crash was already mitigated in v4.62.0
(`compareActualEngineVersion` in `verify.go` + the `engine_version`
`DiffSuppressFunc` in `cluster.go`). What is missing, and what this work adds,
is **surfacing the reason** a change did not take effect.

## 2. Solution overview

Two components, both delivered in **one branch / one PR**:

- **Component A — `aws_rds_events` data source.** A thin, read-only wrapper over
  `DescribeEvents` for declarative/ad-hoc inspection. Exposes an `events`
  attribute; emits no diagnostics.
- **Component B — inline surfacing.** In the SDKv2 resource CRUD paths, after a
  modify/create that did not take effect, query the relevant RDS events and emit
  a **warning** diagnostic explaining why.

Both use one shared finder (§4). They are complementary: the data source is
decoupled from any specific operation (loose timing, re-reads each plan); inline
surfacing is correctly timed to the failing apply but only fires on instrumented
paths.

## 3. AWS API reference (aws-sdk-go-v2 rds v1.126.1)

`rds.DescribeEventsInput` fields: `SourceIdentifier *string` (requires
`SourceType`), `SourceType types.SourceType`, `EventCategories []string`,
`Duration *int32` (minutes, default 60), `StartTime *time.Time`,
`EndTime *time.Time`, `MaxRecords *int32`.

`types.Event`: `Date`, `EventCategories []string`, `Message *string`,
`SourceArn`, `SourceIdentifier`, `SourceType`. **No event ID** — filtering uses
category + message text only.

Paginator: `rds.NewDescribeEventsPaginator`; output field `Events`.

Time semantics (verified, API Reference → Request Parameters): `StartTime`,
`EndTime`, `Duration` are all independent and optional. `StartTime` alone
returns events through *now*; `EndTime` is not required. The 60-minute
`Duration` default applies only when neither `StartTime` nor `EndTime` is set —
so an explicit `StartTime` (Component B) is unaffected by long applies. RDS
retains events for 14 days.

### 3.1 Event categories by source type (verified)

Source: [USER_Events.Messages.html](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/USER_Events.Messages.html).

| Source | Event | Message (abbrev.) | Category |
|---|---|---|---|
| db-cluster | RDS-EVENT-0412 | pre-check for cluster engine upgrade **failed or timed out** | `maintenance` |
| db-cluster | RDS-EVENT-0174 | cluster is in a state that cannot be upgraded | `maintenance` |
| db-instance | RDS-EVENT-0270 | engine upgrade **failed**, rollback succeeded | `maintenance` |
| db-instance | RDS-EVENT-0188 | instance can't be upgraded (MySQL 5.7→8.0 rollback) | `failure` |
| db-instance | RDS-EVENT-0158 | instance in a state that cannot be upgraded | `notification` |
| db-instance | (enhanced monitoring config failure, #41037) | unable to configure enhanced monitoring… | `failure` |

**Implication:** upgrade failures span `failure` **and** `maintenance` (the
Aurora pre-check failure is `maintenance`), so the upgrade gate queries
`{"failure","maintenance"}`. The create-time monitoring failure is `failure`,
so that gate queries `{"failure"}`.

## 4. Shared finder

Location: `internal/service/rds/events_data_source.go` (same file as Component
A — package convention: finders live in their implementation file, not a
separate `find.go`/`events.go`; any file in the package can call a func
regardless of which file defines it, so Component B calls these same
package-level funcs from `instance.go`/`cluster.go`/`cluster_instance.go`/
`blue_green.go` without a new import). Precedent: `findEventCategoriesMaps`
lives in `event_categories_data_source.go`.

```go
// findEvents returns all events matching input (paginated).
func findEvents(ctx context.Context, conn *rds.Client, input *rds.DescribeEventsInput) ([]awstypes.Event, error) {
	var output []awstypes.Event
	pages := rds.NewDescribeEventsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		output = append(output, page.Events...)
	}
	return output, nil
}

// findEventMessagesAfter returns message text for one source, in the given
// categories, since a start time. Categories are a parameter because the
// relevant categories differ per scenario (§3.1).
func findEventMessagesAfter(ctx context.Context, conn *rds.Client, sourceID string, sourceType awstypes.SourceType, since time.Time, categories []string) ([]string, error) {
	input := &rds.DescribeEventsInput{
		SourceIdentifier: aws.String(sourceID),
		SourceType:       sourceType,
		EventCategories:  categories,
		StartTime:        aws.Time(since),
	}
	events, err := findEvents(ctx, conn, input)
	if err != nil {
		return nil, err
	}
	return tfslices.ApplyToAll(events, func(e awstypes.Event) string {
		return aws.ToString(e.Message)
	}), nil
}

var (
	upgradeEventCategories    = []string{"failure", "maintenance"} // update-path upgrade gate (§6.2)
	createTimeEventCategories = []string{"failure"}                // create-time gate (§6.4)
)
```

## 5. Component A — `aws_rds_events` data source

File: `internal/service/rds/events_data_source.go` (Terraform Plugin Framework;
modeled on `snapshots_data_source.go`). Registration is annotation-driven — run
`make gen PKG=rds`.

### 5.1 Schema

| Attribute | Mode | Type | Notes |
|---|---|---|---|
| `source_identifier` | Optional | String | if set, `source_type` required |
| `source_type` | Optional | String | `enum.Validate[types.SourceType]()` |
| `event_categories` | Optional | Set(String) | passthrough; empty = all categories |
| `duration` | Optional | Int64 | minutes; conflicts with `start_time`/`end_time` |
| `start_time` | Optional | `timetypes.RFC3339` | conflicts with `duration` |
| `end_time` | Optional | `timetypes.RFC3339` | conflicts with `duration` |
| `events` | Computed | List(Object) | `framework.DataSourceComputedListOfObjectAttribute[eventModel](ctx)` |

`eventModel`: `date` (`timetypes.RFC3339`), `event_categories`
(`fwtypes.ListOfString`), `message`, `source_arn`, `source_identifier`,
`source_type` (all String).

Validators: `source_identifier` ⇒ `source_type` (`stringvalidator.AlsoRequires`);
`duration` conflicts with `start_time`/`end_time` (`int64validator.ConflictsWith`).

### 5.2 Model & Read

```go
type eventsDataSourceModel struct {
	framework.WithRegionModel
	SourceIdentifier types.String                                `tfsdk:"source_identifier"`
	SourceType       types.String                                `tfsdk:"source_type"`
	EventCategories  fwtypes.SetOfString                         `tfsdk:"event_categories"`
	Duration         types.Int64                                 `tfsdk:"duration"`
	StartTime        timetypes.RFC3339                           `tfsdk:"start_time"`
	EndTime          timetypes.RFC3339                           `tfsdk:"end_time"`
	Events           fwtypes.ListNestedObjectValueOf[eventModel] `tfsdk:"events"`
}

func (d *eventsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	conn := d.Meta().RDSClient(ctx)

	var data eventsDataSourceModel
	smerr.AddEnrich(ctx, &resp.Diagnostics, req.Config.Get(ctx, &data))
	if resp.Diagnostics.HasError() {
		return
	}

	var input rds.DescribeEventsInput
	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Expand(ctx, data, &input))
	if resp.Diagnostics.HasError() {
		return
	}

	out, err := findEvents(ctx, conn, &input)
	if err != nil {
		smerr.AddError(ctx, &resp.Diagnostics, err)
		return
	}

	smerr.AddEnrich(ctx, &resp.Diagnostics, flex.Flatten(ctx, out, &data.Events))
	if resp.Diagnostics.HasError() {
		return
	}

	// Synthetic ID (region + source + window); data source has no natural ID.
	smerr.AddEnrich(ctx, &resp.Diagnostics, resp.State.Set(ctx, &data))
}
```

### 5.3 Docs & tests

- `website/docs/d/rds_events.html.markdown` — description must not start with
  "The/A/An". Document the `depends_on`-after-apply pattern and that `events` is
  time-dependent (reflects a wall-clock window).
- `TestAccRDSEventsDataSource_basic`: create a `db-instance`, read events by
  `source_identifier`/`source_type`, assert loose (attribute presence, not exact
  messages — content is timing-dependent).

## 6. Component B — inline surfacing

### 6.1 Sites to instrument

| Resource | Site | `SourceType` | Gate |
|---|---|---|---|
| `aws_rds_cluster_instance` | `resourceClusterInstanceUpdate` — modify 516 / wait 524 | `db-instance` | upgrade (§6.2) |
| `aws_rds_cluster` | `resourceClusterUpdate` — modify 1748 / wait 1772 | `db-cluster` | upgrade (§6.2) |
| `aws_db_instance` (update, direct) | `resourceInstanceUpdate` — `dbInstanceModify` at 2346 | `db-instance` | upgrade (§6.2) |
| `aws_db_instance` (update, Blue/Green) | `modifyTarget` (blue_green.go:159) | `db-instance` | upgrade (§6.3) |
| `aws_db_instance` (create) | create tail of `resourceInstanceCreate` (~1923) | `db-instance` | create-time (§6.4) |

Not instrumented: create-time cluster/instance modifies, Blue/Green cleanup
(deletion-protection) modify, `db_instance_parameter_group_name` retry modify.

The shared helper `dbInstanceModify` keeps its current signature (returns
`error`). Gating is done by each caller.

### 6.2 Upgrade gate (update paths)

Surface only when **all** hold, using the **post-modify state of the instance
that was actually modified**:

1. an upgrade-relevant site (§6.1);
2. `d.HasChange(names.AttrEngineVersion)`;
3. requested `≠` applied: `requested != aws.ToString(obj.EngineVersion)`;
4. not deferred to a maintenance window:
   `obj.PendingModifiedValues == nil || obj.PendingModifiedValues.EngineVersion == nil`.

Query categories `upgradeEventCategories` (`{"failure","maintenance"}`),
`StartTime = modifyStart` (captured immediately before the modify), keyed on the
modified identifier. Best-effort: a `DescribeEvents`/describe error is logged,
never fatal.

**`aws_db_instance` direct-modify (site 2346)** — obtain the post-modify instance
with an explicit describe of the modified identifier:

```go
modifyStart := time.Now().UTC()

if err := dbInstanceModify(ctx, conn, d.Id(), input, deadline.Remaining()); err != nil {
	// ...existing error handling...
}

if d.HasChange(names.AttrEngineVersion) {
	if instance, err := findDBInstanceByID(ctx, conn, aws.ToString(input.DBInstanceIdentifier)); err == nil {
		requested := d.Get(names.AttrEngineVersion).(string)
		if pending := instance.PendingModifiedValues != nil && instance.PendingModifiedValues.EngineVersion != nil; !pending &&
			requested != aws.ToString(instance.EngineVersion) {
			diags = append(diags, surfaceUpgradeEvents(ctx, conn,
				aws.ToString(input.DBInstanceIdentifier), awstypes.SourceTypeDbInstance, modifyStart)...)
		}
	}
}
```

**`aws_rds_cluster` / `aws_rds_cluster_instance`** — these modify-and-wait on the
same `d.Id()`, so the object returned by their waiters
(`waitDBClusterUpdated` → `*types.DBCluster`, `waitDBClusterInstanceAvailable`
→ `*types.DBInstance`) *is* the modified resource. Capture it and apply the same
gate; no extra describe needed.

```go
func surfaceUpgradeEvents(ctx context.Context, conn *rds.Client, sourceID string, st awstypes.SourceType, since time.Time) diag.Diagnostics {
	var diags diag.Diagnostics
	msgs, err := findEventMessagesAfter(ctx, conn, sourceID, st, since, upgradeEventCategories)
	if err != nil {
		log.Printf("[WARN] describing RDS events for %s: %s", sourceID, err)
		return diags
	}
	for _, m := range msgs {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "RDS reported an event during update that may explain why the change did not take effect",
			Detail:   m + "\n\nReview RDS events (aws rds describe-events) and any pre-upgrade check log for details.",
		})
	}
	return diags
}
```

### 6.3 Upgrade gate (Blue/Green `modifyTarget`)

In `modifyTarget`, `modifyInput.DBInstanceIdentifier` is the **green** instance,
but `dbInstanceModify(ctx, h.conn, d.Id(), …)` waits on `d.Id()` (the **blue**
source — `waitDBInstanceAvailable` polls by its `id` argument). So:

- key the gate and event query on the **green** `identifier`, not `d.Id()`;
- add a **green-keyed** `waitDBInstanceAvailable(ctx, conn, identifier, timeout)`
  inside `modifyTarget` immediately after the modify, and gate off **its**
  returned `*types.DBInstance` (do not add a second `findDBInstanceByID`, and do
  not use `dbInstanceModify`'s blue-keyed wait);
- run this before switchover, while the green instance still exists.

Otherwise the gate is identical to §6.2 (`surfaceUpgradeEvents`, keyed on the
green identifier).

### 6.4 Create-time gate (`aws_db_instance` only — #41037)

**Anchor:** the tail of `resourceInstanceCreate` (after the
`requiresModifyDbInstance` block, ~1923; before `resourceInstanceRead`), with
`createStart` captured at the top of the function. This is downstream of both
routes by which `monitoring_interval`/`monitoring_role_arn` reach the instance —
the plain `CreateDBInstanceInput` (869/873) and the restore-path post-create
modify at site 1914 (1348/1594) — so one gate covers all create sub-paths.

This gate differs from §6.2 and must **not** share its function or categories:

- **non-comparative**: no prior state at create; compare requested input vs the
  configured value the API returns on the created instance;
- **no `PendingModifiedValues`**: enhanced monitoring has no maintenance-window
  deferral;
- **categories `createTimeEventCategories` (`{"failure"}`)**: confirmed by the
  #41037 repro output.

```go
// createStart := time.Now().UTC()  // at the TOP of resourceInstanceCreate

// At the create tail:
if instance, err := findDBInstanceByID(ctx, conn, identifier); err == nil {
	requestedInterval, intOK := d.GetOk("monitoring_interval")
	requestedRoleARN, arnOK := d.GetOk("monitoring_role_arn")
	monitoringRequested := arnOK && requestedRoleARN.(string) != "" && intOK && requestedInterval.(int) > 0
	monitoringDropped := monitoringRequested &&
		(instance.MonitoringInterval == nil || aws.ToInt32(instance.MonitoringInterval) == 0)

	if monitoringDropped {
		diags = append(diags, surfaceCreateTimeEvents(ctx, conn,
			identifier, awstypes.SourceTypeDbInstance, createStart)...)
	}
}

func surfaceCreateTimeEvents(ctx context.Context, conn *rds.Client, sourceID string, st awstypes.SourceType, since time.Time) diag.Diagnostics {
	var diags diag.Diagnostics
	msgs, err := findEventMessagesAfter(ctx, conn, sourceID, st, since, createTimeEventCategories)
	if err != nil {
		log.Printf("[WARN] describing RDS events for %s: %s", sourceID, err)
		return diags
	}
	for _, m := range msgs {
		diags = append(diags, diag.Diagnostic{
			Severity: diag.Warning,
			Summary:  "RDS reported a failure event during create that may explain an unapplied setting",
			Detail:   m + "\n\nReview RDS events (aws rds describe-events) for details.",
		})
	}
	return diags
}
```

Scope: `monitoring_interval`/`monitoring_role_arn` only. Other create-time
settings (IAM auth, Performance Insights, log exports) may share this failure
mode but are excluded pending their own reported case and evidence.

## 7. Cross-cutting

- **IAM**: both components need `rds:DescribeEvents` (note in docs).
- **Region**: data-source model embeds `framework.WithRegionModel`; inline uses
  the existing resource client.
- **Diagnostics**: `smerr.AddEnrich`/`smerr.AddError` in Framework code;
  `sdkdiag` in SDKv2 code.
- **CHANGELOG**: one PR, multiple `.changelog/<PR>.txt` entries —
  `new-data-source` for A; `enhancement` for each Component B resource.
- **AI disclosure**: `🤖🤖🤖` in PR title + note in description.

## 8. Testing & verification

- `make gen PKG=rds` after adding the annotation.
- `make test PKG=rds` (unit); acceptance tests
  `make t T=TestAccRDSEventsDataSource_ K=rds` (require approval; create real
  resources).
- `make quick-fix PKG=rds` (fmt, imports, lint, semgrep, copyright, build).
- `make swissshepherd` for docs.

## 9. Implementation order (single branch / single PR)

1. ~~Shared finder (§4) — unit-testable, no external surface.~~ **Done** —
   merged into `events_data_source.go` per package convention (finders live
   alongside their implementation file; any file in the package can call a
   func regardless of which file defines it).
2. ~~Component A data source + docs + acceptance test (§5).~~ **Done** —
   `internal/service/rds/events_data_source.go`,
   `internal/service/rds/events_data_source_test.go`,
   `website/docs/d/rds_events.html.markdown`. Registered via `make gen PKG=rds`.
   `TestAccRDSEventsDataSource_basic` passed.
3. `aws_rds_cluster_instance` upgrade gate (§6.1–6.2).
4. `aws_rds_cluster` upgrade gate (§6.1–6.2).
5. `aws_db_instance` update: direct-modify (§6.2) + Blue/Green (§6.3).
6. `aws_db_instance` create-time gate (§6.4).

---

# Appendix A — Decisions & rationale

- **Two components, not one.** A data source alone can't time a warning to the
  failing apply (re-reads each plan, loose correlation); inline surfacing alone
  only covers instrumented paths. Ship both.
- **No `emit_warnings` on the data source.** No precedent in `internal/service`
  for a schema-gated informational-warning toggle (the only data-source warnings
  are Route 53 arg-validation and EC2 AMI `allow_unsafe_filter` error→warning
  downgrade). Would re-warn every plan (alarm fatigue) and duplicate Component B.
  Data source stays a pure read; `-json`/`output` already expose `events`.
- **Data source default `event_categories` = empty (all).** A faithful wrapper
  shouldn't editorialize; no flooding concern without `emit_warnings`.
- **Finder in `events_data_source.go`.** The rds package has no standalone finder
  files; finders live in implementation files (precedent:
  `findEventCategoriesMaps` in `event_categories_data_source.go`).
- **`dbInstanceModify` signature unchanged.** Returning its post-wait instance
  would hand the Blue/Green `modifyTarget` path the **blue** instance (the helper
  waits on `d.Id()`, not the green `input.DBInstanceIdentifier`), producing a
  false-positive on every Blue/Green upgrade. Each caller gates explicitly
  instead (§6.2/§6.3).
- **Upgrade categories `{"failure","maintenance"}`.** The Aurora cluster
  pre-check failure (RDS-EVENT-0412) and db-instance upgrade-failed/rollback
  (RDS-EVENT-0270) are `maintenance`; `failure`-only would miss them (§3.1). The
  broad `maintenance` filter is safe only because the gate restricts surfacing to
  the "requested, not applied, not deferred" state.
- **Create-time gate is separate.** Non-comparative, no `PendingModifiedValues`,
  `{"failure"}`-only — different evidence and shape than the upgrade gate; do not
  merge (§6.4).
- **Create-time anchor = create tail.** Monitoring reaches the instance via two
  routes (plain `CreateDBInstanceInput`; restore-path post-create modify at site
  1914); a single tail gate is downstream of both.
- **`compareActualEngineVersion` does not gate surfacing.** It only reconciles
  the `engine_version`/`engine_version_actual` state attributes; condition 4 of
  the upgrade gate must be wired explicitly.
- **Verified call shapes.** `flex.Expand(ctx, data, &input)` order matches
  `snapshots_data_source.go:84` et al.; `waitDBInstanceAvailable` polls by its
  `id` arg; pending-version precedent at `instance.go:2786`, `cluster.go:2014`,
  `cluster_instance.go:643` (generic `!inttypes.IsZero(pmv)` at `cluster.go:2093`).

# Appendix B — Open questions

- **Message-substring refine.** Whether to add a heuristic message filter
  (`fail`, `cannot`, `rolled back`, `timed out`) on top of the category filter.
  Deferred until acceptance testing shows whether benign `maintenance` messages
  leak through the upgrade gate.
- **Synthetic data-source ID.** Confirm the region+source+window composition has
  no collision concerns across repeated reads.

# Appendix C — Flagged, out of scope

- **`modifyTarget` waits on the wrong id (pre-existing).**
  `dbInstanceModify(ctx, h.conn, d.Id(), …)` waits on the blue source after
  modifying the green target; likely vestigial because `resourceInstanceUpdate`
  already waits on the green instance (`waitDBInstanceAvailable(...,
  targetARN.Identifier, ...)`, ~2231) *before* `modifyTarget`. Nothing waits on
  the green instance *after* its modify — a possible race before `Switchover`.
  This design does not depend on it (§6.3 adds its own green-keyed wait), but it
  warrants a separate look if Blue/Green reliability is investigated.

# Appendix D — Correction history (design process)

- Initial anchor cited `instance.go` ~1913/1920 — wrong (that's
  `resourceInstanceCreate`, not update). Corrected to the verified site map (§6.1).
- Proposed instrumenting the shared `dbInstanceModify` and returning its
  post-wait instance — wrong for Blue/Green (returns blue). Switched to
  explicit per-caller gating (§6.2/§6.3).
- Proposed `EventCategories=["failure"]` — wrong; would miss the Aurora
  `maintenance` pre-check failure. Corrected to `{"failure","maintenance"}` (§3.1).
- Create-time (#41037) anchor: first assumed site 1914, then "1914 excluded
  entirely" — both wrong. Monitoring routes through both create input and site
  1914; final anchor is the create tail (§6.4).
- `emit_warnings` on the data source — considered, rejected (Appendix A).
