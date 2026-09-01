// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package batch_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBatchJobQueueDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, "tf_acc_test_")
	resourceName := "aws_batch_job_queue.test"
	dataSourceName := "data.aws_batch_job_queue.by_name"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccJobQueueDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "compute_environment_order.#", resourceName, "compute_environment_order.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrPriority, resourceName, names.AttrPriority),
					// resource.TestCheckResourceAttrPair(dataSourceName, "scheduling_policy_arn", resourceName, "scheduling_policy_arn"),
					resource.TestCheckResourceAttr(dataSourceName, "scheduling_policy_arn", ""),
					resource.TestCheckNoResourceAttr(resourceName, "scheduling_policy_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrState, resourceName, names.AttrState),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{})),
				},
			},
		},
	})
}

func TestAccBatchJobQueueDataSource_schedulingPolicy(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, "tf_acc_test_")
	resourceName := "aws_batch_job_queue.test"
	dataSourceName := "data.aws_batch_job_queue.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccJobQueueDataSourceConfig_schedulingPolicy(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "compute_environment_order.#", resourceName, "compute_environment_order.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "job_state_time_limit_action.#", resourceName, "job_state_time_limit_action.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrPriority, resourceName, names.AttrPriority),
					resource.TestCheckResourceAttrPair(dataSourceName, "scheduling_policy_arn", resourceName, "scheduling_policy_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrState, resourceName, names.AttrState),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{})),
				},
			},
		},
	})
}

func testAccJobQueueDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccJobQueueConfig_base(rName),
		fmt.Sprintf(`
resource "aws_batch_job_queue" "test" {
  name     = "%[1]s"
  state    = "ENABLED"
  priority = 1
  compute_environment_order {
    compute_environment = aws_batch_compute_environment.test.arn
    order               = 1
  }
}

resource "aws_batch_job_queue" "wrong" {
  name     = "%[1]s_wrong"
  state    = "ENABLED"
  priority = 2
  compute_environment_order {
    compute_environment = aws_batch_compute_environment.test.arn
    order               = 1
  }

}

data "aws_batch_job_queue" "by_name" {
  name = aws_batch_job_queue.test.name
}
`, rName))
}

func testAccJobQueueDataSourceConfig_schedulingPolicy(rName string) string {
	return acctest.ConfigCompose(
		testAccJobQueueConfig_base(rName),
		fmt.Sprintf(`
resource "aws_batch_scheduling_policy" "test" {
  name = %[1]q

  fair_share_policy {
    compute_reservation = 1
    share_decay_seconds = 3600

    share_distribution {
      share_identifier = "A1*"
      weight_factor    = 0.1
    }
  }
}

resource "aws_batch_job_queue" "test" {
  name                  = %[1]q
  scheduling_policy_arn = aws_batch_scheduling_policy.test.arn
  state                 = "ENABLED"
  priority              = 1
  compute_environment_order {
    compute_environment = aws_batch_compute_environment.test.arn
    order               = 1
  }


  job_state_time_limit_action {
    action           = "CANCEL"
    max_time_seconds = 600
    reason           = "MISCONFIGURATION:JOB_RESOURCE_REQUIREMENT"
    state            = "RUNNABLE"
  }
}

data "aws_batch_job_queue" "test" {
  name = aws_batch_job_queue.test.name
}
`, rName))
}
