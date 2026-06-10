---
applyTo: "internal/service/**/*_test.go"
---

# Acceptance & Unit Tests

Scope: resources, data sources, and list resources. Helper-test patterns (Exists/Destroy, data source tests, list resource tests, unit tests) are in `acceptance-tests-helpers.instructions.md`. Ephemeral resources, provider functions, and actions are reviewed rarely; these rules don't target them.

## What every new resource test file must contain

A new resource needs at least:

- `TestAcc<Service><Resource>_basic` — full happy path, including an `ImportState` step, check all attributes.
- `TestAcc<Service><Resource>_disappears` — verifies the provider re-creates a resource that's gone out-of-band.

Tag tests and identity tests are **generated** for resources annotated with `@Tags` and identity annotations. Flag PRs that add hand-written `_tags*` or `_Identity_*` tests for new resources.

## Naming

- Acceptance tests: `TestAcc<Service><Resource>_<scenario>`.
- Data source acceptance tests: `TestAcc<Service><DataSource>DataSource_<scenario>`.
- List resource acceptance tests: `TestAcc<Service><Resource>_List_<scenario>`.
- Unit tests: anything **without** the `TestAcc` prefix; flag any unit test that calls AWS.

## TestCase essentials

Acceptance tests normally start with `ctx := acctest.Context(t)` and use `acctest.ParallelTest(ctx, t, resource.TestCase{...})`.

The `TestCase` must set:

- `PreCheck` calling `acctest.PreCheck(ctx, t)`, `acctest.PreCheckPartitionHasService(t, names.<Service>EndpointID)`, and the package's `testAccPreCheck(ctx, t)`.
- `ErrorCheck: acctest.ErrorCheck(t, names.<Service>ServiceID)`.
- `ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories`.
- `CheckDestroy: testAccCheck<Resource>Destroy(ctx, t)`.

Flag missing or replaced versions of any of those.

For tests expected to run more than ~5 minutes, add the long-running guard right after `acctest.Context(t)`:

```go
if testing.Short() {
    t.Skip("skipping long-running test in short mode")
}
```

## Random naming

Use `sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)`. Flag hard-coded names or bare `acctest.RandString(...)`.

## PreCheck pattern

`testAccPreCheck` issues a single cheap List/Describe call and skips the test on partition/permission errors via `acctest.PreCheckSkipError(err)`. Flag PreChecks that make multiple API calls, return errors instead of calling `t.Skipf` / `t.Fatalf`, or do not run through `PreCheckSkipError`.

## ImportState step

The `_basic` test's last step verifies import:

```go
{
    ResourceName:      resourceName,
    ImportState:       true,
    ImportStateVerify: true,
}
```

`ImportStateVerifyIgnore` is for write-only fields the AWS API doesn't return (e.g., passwords, `apply_immediately`). Flag broad ignore lists used to paper over genuine drift.

## Disappears test (Framework vs SDKv2)

- Framework: `acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tf<svc>.Resource<Name>, resourceName)`.
- SDKv2: `acctest.CheckResourceDisappears(ctx, acctest.Provider, tf<svc>.Resource<Name>(), resourceName)`.

## exports_test.go

Test files reach into the package via a sibling `exports_test.go`:

```go
package <svc>

var (
    Resource<Name> = new<Name>Resource   // Framework
    // or: Resource<Name> = resource<Name> // SDKv2
    Find<Name>ByID = find<Name>ByID
)
```

Flag PRs that:

- Export production identifiers (capitalize the real `find<Name>ByID`) instead of using `exports_test.go`.
- Reach into the package via build tags or `internal/` traversal hacks.

The package itself is imported in tests as `tf<svc> "github.com/hashicorp/terraform-provider-aws/internal/service/<svc>"`.

## Regex and ARN checks

- Use `github.com/YakDriver/regexache`, not stdlib `regexp`. Flag any new test that imports `regexp`.
- For ARN attributes use `acctest.MatchResourceAttrRegionalARN` / `acctest.CheckResourceAttrRegionalARN` (or the global / alternate-region variants). Flag manual ARN assembly with `fmt.Sprintf` containing account ID or region.
