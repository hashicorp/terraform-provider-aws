---
applyTo: "internal/service/**/*_test.go"
---
<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# Test Helpers, Data Source Tests, List Resource Tests, Unit Tests

Scope: Exists/Destroy helpers, data source tests, list resource tests, unit tests. See `acceptance-tests.instructions.md` for per-resource basics.

## Exists / Destroy helpers

`testAccCheck<Resource>Exists` should:

- Look up the resource state with `s.RootModule().Resources[name]`.
- Verify `rs.Primary.ID != ""`.
- Use `acctest.ProviderMeta(ctx, t).<Service>Client(ctx)` for the client — never construct a fresh client.
- Call the **exported** finder `tf<svc>.Find<Name>ByID(ctx, conn, ...)`.

`testAccCheck<Resource>Destroy` should:

- Skip rows where `rs.Type != "aws_<svc>_<thing>"`.
- Treat `retry.NotFound(err)` as success (`return nil`).
- Use `errs.IsA[*awstypes.<NotFoundException>]` rather than type assertions if it inspects errors directly.

Wrap real failures with `create.Error(names.<Service>, create.ErrActionCheckingExistence|ErrActionCheckingDestroyed, tf<svc>.ResName<Name>, id, err)` rather than `fmt.Errorf` / `errors.New`.

## Data source tests

Data source tests follow the resource conventions with these differences:

- Reference `dataSourceName := "data.aws_<svc>_<thing>.test"`.
- Prefer `resource.TestCheckResourceAttrPair(dataSourceName, attr, resourceName, attr)` over hard-coded values when the data source mirrors a resource.
- No `_disappears` test.
- A `CheckDestroy` is still required when the test creates a backing resource.

## List resource tests

List resources require three scenarios for parity:

- `_List_basic` — basic listing.
- `_List_includeResource` — with `include_resource = true` and full attribute checks.
- `_List_regionOverride` — region override; requires `acctest.PreCheckMultipleRegion(t, 2)`.

List resource tests use static testdata, not inline configs:

```go
ConfigDirectory: config.StaticDirectory("testdata/<Resource>/list_<scenario>/"),
ConfigVariables: config.Variables{ acctest.CtRName: config.StringVariable(rName), ... },
```

A separate `Step` with `Query: true` exercises the list operation. Identity assertions use `tfstatecheck.Identity()` / `identity.GetIdentity(resourceName)` and the `tfquerycheck.*` helpers under `internal/acctest/querycheck`.

List resource tests also require a Terraform version floor (currently 1.14):

```go
TerraformVersionChecks: []tfversion.TerraformVersionCheck{
    tfversion.SkipBelow(tfversion.Version1_14_0),
},
```

## Unit tests

Unit tests are for logic that doesn't touch AWS — parsers, custom flatteners/expanders, ID composition, validators. They:

- Run in parallel (`t.Parallel()` at top and inside subtests).
- Are table-driven with `t.Run(tc.TestName, ...)`.
- Must not call `acctest.Context`, `acctest.PreCheck`, or instantiate an AWS client.

Flag unit tests added for trivial pass-through flatteners/expanders — they're noise. Flag acceptance-style tests mis-named without the `TestAcc` prefix.
