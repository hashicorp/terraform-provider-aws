// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testSDKDisappearsWithoutPlanCheck(t *testing.T) {
	ctx := acctest.Context(t)

	const resourceName = "aws_example_thing.test"

	acctest.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExists(ctx, t, resourceName),
					// ruleid: disappears-expect-resource-action
					acctest.CheckSDKResourceDisappears(ctx, t, ResourceThing(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testSDKDisappearsWithPlanCheck(t *testing.T) {
	ctx := acctest.Context(t)

	const resourceName = "aws_example_thing.test"

	acctest.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExists(ctx, t, resourceName),
					// ok: disappears-expect-resource-action
					acctest.CheckSDKResourceDisappears(ctx, t, ResourceThing(), resourceName),
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

func testSDKDisappearsWithWrongAction(t *testing.T) {
	ctx := acctest.Context(t)

	const resourceName = "aws_example_thing.test"

	acctest.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExists(ctx, t, resourceName),
					// ruleid: disappears-expect-resource-action
					acctest.CheckSDKResourceDisappears(ctx, t, ResourceThing(), resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func testFrameworkDisappearsWithoutPlanCheck(t *testing.T) {
	ctx := acctest.Context(t)

	const resourceName = "aws_example_thing.test"

	acctest.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckExists(ctx, t, resourceName),
					// ruleid: disappears-expect-resource-action
					acctest.CheckFrameworkResourceDisappears(ctx, t, ResourceThing, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testFrameworkDisappearsWithPlanCheck(t *testing.T) {
	ctx := acctest.Context(t)

	const resourceName = "aws_example_thing.test"

	acctest.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckExists(ctx, t, resourceName),
					// ok: disappears-expect-resource-action
					acctest.CheckFrameworkResourceDisappears(ctx, t, ResourceThing, resourceName),
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

func testFrameworkDisappearsWithWrongAction(t *testing.T) {
	ctx := acctest.Context(t)

	const resourceName = "aws_example_thing.test"

	acctest.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckExists(ctx, t, resourceName),
					// ruleid: disappears-expect-resource-action
					acctest.CheckFrameworkResourceDisappears(ctx, t, ResourceThing, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func testFrameworkStateFuncDisappearsWithoutPlanCheck(t *testing.T) {
	ctx := acctest.Context(t)

	const resourceName = "aws_example_thing.test"

	acctest.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExists(ctx, t, resourceName),
					// ruleid: disappears-expect-resource-action
					acctest.CheckFrameworkResourceDisappearsWithStateFunc(ctx, t, ResourceThing, resourceName, stateFunc),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testFrameworkStateFuncDisappearsWithPlanCheck(t *testing.T) {
	ctx := acctest.Context(t)

	const resourceName = "aws_example_thing.test"

	acctest.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExists(ctx, t, resourceName),
					// ok: disappears-expect-resource-action
					acctest.CheckFrameworkResourceDisappearsWithStateFunc(ctx, t, ResourceThing, resourceName, stateFunc),
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

func testFrameworkStateFuncDisappearsWithWrongAction(t *testing.T) {
	ctx := acctest.Context(t)

	const resourceName = "aws_example_thing.test"

	acctest.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExists(ctx, t, resourceName),
					// ruleid: disappears-expect-resource-action
					acctest.CheckFrameworkResourceDisappearsWithStateFunc(ctx, t, ResourceThing, resourceName, stateFunc),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}
