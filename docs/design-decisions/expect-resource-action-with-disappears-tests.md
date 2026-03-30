<!-- Copyright IBM Corp. 2014, 2026 -->
<!-- SPDX-License-Identifier: MPL-2.0 -->

# Use `plancheck.ExpectResourceAction` with disappears acceptance tests

**Summary:** Acceptance tests exercising out of band deletion (colloquially named "disappears" tests) should utilize the terraform-plugin-testing library's plancheck package to assert expected post apply actions.  
**Created**: 2025-04-11  
**Author**: [@jar-b](https://github.com/jar-b)  

---

## Background

Resources implemented in the Terraform AWS provider commonly include a “disappears” acceptance test, which exercises the expected behavior of the resource following an out of band deletion (e.g. deletion outside of Terraform). The goal of these tests is to ensure that the resource’s `Read` operation correctly detects when the resource cannot be found and removes it from state. The [contributor guide](https://hashicorp.github.io/terraform-provider-aws/running-and-writing-acceptance-tests/#disappears-acceptance-tests) includes the following example of a “disappears” test.

```
func TestAccExampleThing_disappears(t *testing.T) {
  ctx := acctest.Context(t)
  rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
  resourceName := "aws_example_thing.test"

  resource.ParallelTest(t, resource.TestCase{
    PreCheck:                 func() { acctest.PreCheck(ctx, t) },
    ErrorCheck:               acctest.ErrorCheck(t, names.ExampleServiceID),
        ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
    CheckDestroy:             testAccCheckExampleThingDestroy(ctx),
    Steps: []resource.TestStep{
      {
        Config: testAccExampleThingConfigName(rName),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckExampleThingExists(ctx, resourceName),
          acctest.CheckResourceDisappears(ctx, acctest.Provider, ResourceExampleThing(), resourceName),
        ),
        ExpectNonEmptyPlan: true,
      },
    },
  })
}
```

“Disappears” tests are distinctive in that they *expect* a non-empty plan because a deletion was intentionally triggered during the check phase via the [`acctest.CheckResourceDisappears`](https://github.com/hashicorp/terraform-provider-aws/blob/3f01f342585ac1656fe6421f196a5850a9c3b685/internal/acctest/acctest.go#L1474-L1493) helper. The impacted resource should be dropped from state during the post apply refresh, and therefore planned to be re-created. `ExpectNonEmptyPlan` ensures we don’t fail to detect out of band deletions, however, it doesn’t explicitly verify *what* changes are planned.

In [`v1.2.0`](https://github.com/hashicorp/terraform-plugin-testing/releases/tag/v1.2.0), the [`terraform-plugin-testing`](https://github.com/hashicorp/terraform-plugin-testing) library introduced a `plancheck` package with an `ExpectResourceAction` built-in plan check, which asserts that a given resource will have a specific resource change type in the plan. With this check “disappears” tests can explicitly verify the plan of the deleted resource.

## Decision

The contributor guide and [`skaff`](https://hashicorp.github.io/terraform-provider-aws/skaff/) will be updated to include a post-apply, post-refresh plan check to verify that the “disappeared” resource is planned for creation. For example,

```
func TestAccExampleThing_disappears(t *testing.T) {
  ctx := acctest.Context(t)
  rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
  resourceName := "aws_example_thing.test"

  resource.ParallelTest(t, resource.TestCase{
    PreCheck:                 func() { acctest.PreCheck(ctx, t) },
    ErrorCheck:               acctest.ErrorCheck(t, names.ExampleServiceID),
        ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
    CheckDestroy:             testAccCheckExampleThingDestroy(ctx),
    Steps: []resource.TestStep{
      {
        Config: testAccExampleThingConfigName(rName),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckExampleThingExists(ctx, resourceName),
          acctest.CheckResourceDisappears(ctx, acctest.Provider, ResourceExampleThing(), resourceName),
        ),
        ExpectNonEmptyPlan: true,
        ConfigPlanChecks: resource.ConfigPlanChecks{
          PostApplyPostRefresh: []plancheck.PlanCheck{
            plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
          },
        },
      },
    },
  })
}
```

## Consequences/Future Work

Given the additional `ConfigPlanCheck` has only a single variable reference (`resourceName`) which is a standard naming convention across all acceptance tests, we should explore automating the addition of this check to existing “disappears” tests as well.
