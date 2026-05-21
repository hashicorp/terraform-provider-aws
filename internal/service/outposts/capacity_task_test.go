// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package outposts_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/outposts"
	awstypes "github.com/aws/aws-sdk-go-v2/service/outposts/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfoutposts "github.com/hashicorp/terraform-provider-aws/internal/service/outposts"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccOutpostsCapacityTask_basic(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var capacityTask outposts.GetCapacityTaskOutput
	resourceName := "aws_outposts_capacity_task.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOutpostsOutposts(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OutpostsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapacityTaskDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityTaskConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCapacityTaskExists(ctx, t, resourceName, &capacityTask),
					resource.TestCheckResourceAttrSet(resourceName, "capacity_task_id"),
					resource.TestCheckResourceAttrSet(resourceName, "outpost_identifier"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreationDate),
					resource.TestCheckResourceAttr(resourceName, "instance_pool.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_pool.0.count", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "instance_pool.0.instance_type"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"instances_to_exclude"},
			},
		},
	})
}

func TestAccOutpostsCapacityTask_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var capacityTask outposts.GetCapacityTaskOutput
	resourceName := "aws_outposts_capacity_task.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOutpostsOutposts(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OutpostsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapacityTaskDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityTaskConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCapacityTaskExists(ctx, t, resourceName, &capacityTask),
					// Invokes the resource's own Delete, which performs the cancel+wait
					// dispatch dictated by BR-Terminal-State-Matrix. This satisfies
					// BR-Disappears-Test-Mechanism without a bespoke test-side helper.
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfoutposts.ResourceCapacityTask, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccOutpostsCapacityTask_taskActionWaitForEvacuation(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var capacityTask outposts.GetCapacityTaskOutput
	resourceName := "aws_outposts_capacity_task.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOutpostsOutposts(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OutpostsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapacityTaskDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityTaskConfig_taskAction(string(awstypes.TaskActionOnBlockingInstancesWaitForEvacuation)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCapacityTaskExists(ctx, t, resourceName, &capacityTask),
					resource.TestCheckResourceAttr(resourceName, "task_action_on_blocking_instances", string(awstypes.TaskActionOnBlockingInstancesWaitForEvacuation)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"instances_to_exclude"},
			},
		},
	})
}

func TestAccOutpostsCapacityTask_instancesToExclude(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var capacityTask outposts.GetCapacityTaskOutput
	resourceName := "aws_outposts_capacity_task.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOutpostsOutposts(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OutpostsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapacityTaskDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityTaskConfig_instancesToExclude(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCapacityTaskExists(ctx, t, resourceName, &capacityTask),
					resource.TestCheckResourceAttr(resourceName, "instances_to_exclude.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instances_to_exclude.0.instances.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// BR-Read-Drift-WriteOnly: instances_to_exclude may not be echoed by the API;
				// importer cannot reconstruct it, so skip verification for this attribute.
				ImportStateVerifyIgnore: []string{"instances_to_exclude"},
			},
		},
	})
}

func TestAccOutpostsCapacityTask_multipleInstancePools(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var capacityTask outposts.GetCapacityTaskOutput
	resourceName := "aws_outposts_capacity_task.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOutpostsOutposts(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OutpostsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapacityTaskDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityTaskConfig_multipleInstancePools(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCapacityTaskExists(ctx, t, resourceName, &capacityTask),
					resource.TestCheckResourceAttr(resourceName, "instance_pool.#", "2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"instances_to_exclude"},
			},
		},
	})
}

func TestAccOutpostsCapacityTask_assetId(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var capacityTask outposts.GetCapacityTaskOutput
	resourceName := "aws_outposts_capacity_task.test"
	assetsDataSourceName := "data.aws_outposts_assets.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckOutpostsOutposts(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.OutpostsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapacityTaskDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityTaskConfig_assetId(),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCapacityTaskExists(ctx, t, resourceName, &capacityTask),
					resource.TestCheckResourceAttrSet(resourceName, "asset_id"),
					resource.TestCheckResourceAttrPair(resourceName, "asset_id", assetsDataSourceName, "asset_ids.0"),
					resource.TestCheckResourceAttr(resourceName, "instance_pool.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"instances_to_exclude"},
			},
		},
	})
}

func testAccCheckCapacityTaskDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).OutpostsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_outposts_capacity_task" {
				continue
			}

			_, err := tfoutposts.FindCapacityTaskByTwoPartKey(ctx, conn, rs.Primary.Attributes["outpost_identifier"], rs.Primary.Attributes["capacity_task_id"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			// CheckDestroy tolerates terminal states (COMPLETED / CANCELLED / FAILED) because
			// BR-Terminal-State-Matrix intentionally leaves these records in place on Terraform
			// destroy. Only non-terminal survivors are failures.
			if out, err2 := tfoutposts.FindCapacityTaskByTwoPartKey(ctx, conn, rs.Primary.Attributes["outpost_identifier"], rs.Primary.Attributes["capacity_task_id"]); err2 == nil {
				switch out.CapacityTaskStatus {
				case awstypes.CapacityTaskStatusCompleted,
					awstypes.CapacityTaskStatusCancelled,
					awstypes.CapacityTaskStatusFailed:
					continue
				}
			}

			return fmt.Errorf("Outposts Capacity Task (%s/%s) still exists",
				rs.Primary.Attributes["outpost_identifier"],
				rs.Primary.Attributes["capacity_task_id"])
		}

		return nil
	}
}

func testAccCheckCapacityTaskExists(ctx context.Context, t *testing.T, n string, v *outposts.GetCapacityTaskOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).OutpostsClient(ctx)

		output, err := tfoutposts.FindCapacityTaskByTwoPartKey(ctx, conn, rs.Primary.Attributes["outpost_identifier"], rs.Primary.Attributes["capacity_task_id"])
		if err != nil {
			return err
		}

		*v = *output
		return nil
	}
}

// testAccCapacityTaskConfig_base provides data-source lookups that every CapacityTask
// test re-uses. Avoids hardcoding Outpost IDs or instance types — values come from the
// test account's first available Outpost.
func testAccCapacityTaskConfig_base() string {
	return `
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost_instance_types" "test" {
  arn = tolist(data.aws_outposts_outposts.test.arns)[0]
}
`
}

func testAccCapacityTaskConfig_basic() string {
	return acctest.ConfigCompose(testAccCapacityTaskConfig_base(), `
resource "aws_outposts_capacity_task" "test" {
  outpost_identifier = tolist(data.aws_outposts_outposts.test.arns)[0]

  instance_pool {
    instance_type = tolist(data.aws_outposts_outpost_instance_types.test.instance_types)[0]
    count         = 1
  }
}
`)
}

func testAccCapacityTaskConfig_taskAction(taskAction string) string {
	return acctest.ConfigCompose(testAccCapacityTaskConfig_base(), fmt.Sprintf(`
resource "aws_outposts_capacity_task" "test" {
  outpost_identifier                = tolist(data.aws_outposts_outposts.test.arns)[0]
  task_action_on_blocking_instances = %[1]q

  instance_pool {
    instance_type = tolist(data.aws_outposts_outpost_instance_types.test.instance_types)[0]
    count         = 1
  }
}
`, taskAction))
}

func testAccCapacityTaskConfig_instancesToExclude() string {
	return acctest.ConfigCompose(testAccCapacityTaskConfig_base(), `
resource "aws_outposts_capacity_task" "test" {
  outpost_identifier = tolist(data.aws_outposts_outposts.test.arns)[0]

  instance_pool {
    instance_type = tolist(data.aws_outposts_outpost_instance_types.test.instance_types)[0]
    count         = 1
  }

  # Synthetic instance ID — acceptance test intentionally passes a non-existent ID so the
  # CapacityTask doesn't actually evacuate customer capacity. The API accepts the exclusion
  # list and the Outpost plan accounts for it, but nothing is actually stopped.
  instances_to_exclude {
    instances = ["i-00000000000000000"]
  }
}
`)
}

func testAccCapacityTaskConfig_multipleInstancePools() string {
	return acctest.ConfigCompose(testAccCapacityTaskConfig_base(), `
resource "aws_outposts_capacity_task" "test" {
  outpost_identifier = tolist(data.aws_outposts_outposts.test.arns)[0]

  instance_pool {
    instance_type = tolist(data.aws_outposts_outpost_instance_types.test.instance_types)[0]
    count         = 1
  }

  instance_pool {
    instance_type = tolist(data.aws_outposts_outpost_instance_types.test.instance_types)[1]
    count         = 1
  }
}
`)
}

func testAccCapacityTaskConfig_assetId() string {
	return acctest.ConfigCompose(testAccCapacityTaskConfig_base(), `
data "aws_outposts_assets" "test" {
  arn = tolist(data.aws_outposts_outposts.test.arns)[0]
}

resource "aws_outposts_capacity_task" "test" {
  outpost_identifier = tolist(data.aws_outposts_outposts.test.arns)[0]
  asset_id           = data.aws_outposts_assets.test.asset_ids[0]

  instance_pool {
    instance_type = tolist(data.aws_outposts_outpost_instance_types.test.instance_types)[0]
    count         = 1
  }
}
`)
}
