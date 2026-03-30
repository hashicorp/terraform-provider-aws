// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package batch_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/batch"
	awstypes "github.com/aws/aws-sdk-go-v2/service/batch/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfbatch "github.com/hashicorp/terraform-provider-aws/internal/service/batch"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBatchJobQueue_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var jobQueue1 awstypes.JobQueueDetail
	resourceName := "aws_batch_job_queue.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobQueueDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobQueueConfig_state(rName, string(awstypes.JQStateEnabled)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobQueueExists(ctx, t, resourceName, &jobQueue1),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "batch", "job-queue/{name}"),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_order.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "compute_environment_order.0.compute_environment", "aws_batch_compute_environment.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "job_state_time_limit_action.#", "0"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrPriority, "1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.JQStateEnabled)),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBatchJobQueue_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var jobQueue1 awstypes.JobQueueDetail
	resourceName := "aws_batch_job_queue.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobQueueDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobQueueConfig_state(rName, string(awstypes.JQStateEnabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobQueueExists(ctx, t, resourceName, &jobQueue1),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfbatch.ResourceJobQueue, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccBatchJobQueue_ComputeEnvironments_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	var jobQueue1 awstypes.JobQueueDetail
	resourceName := "aws_batch_job_queue.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobQueueDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobQueueConfig_ComputeEnvironments_multiple(rName, string(awstypes.JQStateEnabled)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobQueueExists(ctx, t, resourceName, &jobQueue1),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_order.#", "3"),
					resource.TestCheckResourceAttrPair(resourceName, "compute_environment_order.0.compute_environment", "aws_batch_compute_environment.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "compute_environment_order.1.compute_environment", "aws_batch_compute_environment.more.0", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "compute_environment_order.2.compute_environment", "aws_batch_compute_environment.more.1", names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccJobQueueConfig_ComputeEnvironments_multipleReorder(rName, string(awstypes.JQStateEnabled)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobQueueExists(ctx, t, resourceName, &jobQueue1),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_order.#", "3"),
					resource.TestCheckResourceAttrPair(resourceName, "compute_environment_order.0.compute_environment", "aws_batch_compute_environment.more.0", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "compute_environment_order.1.compute_environment", "aws_batch_compute_environment.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "compute_environment_order.2.compute_environment", "aws_batch_compute_environment.more.1", names.AttrARN),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBatchJobQueue_ComputeEnvironmentOrder_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	var jobQueue1 awstypes.JobQueueDetail
	resourceName := "aws_batch_job_queue.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobQueueDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobQueueConfig_ComputeEnvironmentOrder_multiple(rName, string(awstypes.JQStateEnabled), 1, 2, 3),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobQueueExists(ctx, t, resourceName, &jobQueue1),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_order.#", "3"),
					resource.TestCheckResourceAttrPair(resourceName, "compute_environment_order.0.compute_environment", "aws_batch_compute_environment.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "compute_environment_order.1.compute_environment", "aws_batch_compute_environment.more.0", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "compute_environment_order.2.compute_environment", "aws_batch_compute_environment.more.1", names.AttrARN),
				),
			},
			{
				Config: testAccJobQueueConfig_ComputeEnvironmentOrder_multiple(rName, string(awstypes.JQStateEnabled), 2, 1, 3),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckJobQueueExists(ctx, t, resourceName, &jobQueue1),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_order.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_order.0.order", "2"),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_order.1.order", "1"),
					resource.TestCheckResourceAttr(resourceName, "compute_environment_order.2.order", "3"),
					resource.TestCheckResourceAttrPair(resourceName, "compute_environment_order.0.compute_environment", "aws_batch_compute_environment.test", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "compute_environment_order.1.compute_environment", "aws_batch_compute_environment.more.0", names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "compute_environment_order.2.compute_environment", "aws_batch_compute_environment.more.1", names.AttrARN),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/8083
func TestAccBatchJobQueue_ComputeEnvironments_externalOrderUpdate(t *testing.T) {
	ctx := acctest.Context(t)
	var jobQueue1 awstypes.JobQueueDetail
	resourceName := "aws_batch_job_queue.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobQueueDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobQueueConfig_state(rName, string(awstypes.JQStateEnabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobQueueExists(ctx, t, resourceName, &jobQueue1),
					testAccCheckJobQueueComputeEnvironmentOrderUpdate(ctx, t, &jobQueue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBatchJobQueue_priority(t *testing.T) {
	ctx := acctest.Context(t)
	var jobQueue1, jobQueue2 awstypes.JobQueueDetail
	resourceName := "aws_batch_job_queue.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobQueueDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobQueueConfig_priority(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobQueueExists(ctx, t, resourceName, &jobQueue1),
					resource.TestCheckResourceAttr(resourceName, names.AttrPriority, "1"),
				),
			},
			{
				Config: testAccJobQueueConfig_priority(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobQueueExists(ctx, t, resourceName, &jobQueue2),
					resource.TestCheckResourceAttr(resourceName, names.AttrPriority, "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBatchJobQueue_schedulingPolicy(t *testing.T) {
	ctx := acctest.Context(t)
	var jobQueue1, jobQueue2 awstypes.JobQueueDetail
	resourceName := "aws_batch_job_queue.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	schedulingPolicyName1 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	schedulingPolicyName2 := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobQueueDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				// last variable selects the scheduling policy's arn. In this case, the first scheduling policy's arn.
				Config: testAccJobQueueConfig_schedulingPolicy(rName, schedulingPolicyName1, schedulingPolicyName2, "first"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobQueueExists(ctx, t, resourceName, &jobQueue1),
					resource.TestCheckResourceAttrSet(resourceName, "scheduling_policy_arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// test switching the scheduling_policy_arn by changing the last variable to select the second scheduling policy's arn.
				Config: testAccJobQueueConfig_schedulingPolicy(rName, schedulingPolicyName1, schedulingPolicyName2, "second"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobQueueExists(ctx, t, resourceName, &jobQueue2),
					resource.TestCheckResourceAttrSet(resourceName, "scheduling_policy_arn"),
				),
			},
		},
	})
}

func TestAccBatchJobQueue_state(t *testing.T) {
	ctx := acctest.Context(t)
	var jobQueue1, jobQueue2 awstypes.JobQueueDetail
	resourceName := "aws_batch_job_queue.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobQueueDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobQueueConfig_state(rName, string(awstypes.JQStateDisabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobQueueExists(ctx, t, resourceName, &jobQueue1),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.JQStateDisabled)),
				),
			},
			{
				Config: testAccJobQueueConfig_state(rName, string(awstypes.JQStateEnabled)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobQueueExists(ctx, t, resourceName, &jobQueue2),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.JQStateEnabled)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccBatchJobQueue_jobStateTimeLimitActionsMultiple(t *testing.T) {
	ctx := acctest.Context(t)
	var jobQueue1 awstypes.JobQueueDetail
	resourceName := "aws_batch_job_queue.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckJobQueueDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccJobQueueConfig_jobStateTimeLimitAction(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobQueueExists(ctx, t, resourceName, &jobQueue1),
					resource.TestCheckResourceAttr(resourceName, "job_state_time_limit_action.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccJobQueueConfig_jobStateTimeLimitActionUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobQueueExists(ctx, t, resourceName, &jobQueue1),
					resource.TestCheckResourceAttr(resourceName, "job_state_time_limit_action.#", "2"),
				),
			},
		},
	})
}

func TestAccBatchJobQueue_upgradeComputeEnvironments(t *testing.T) {
	ctx := acctest.Context(t)
	var jobQueue1 awstypes.JobQueueDetail
	resourceName := "aws_batch_job_queue.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.BatchServiceID),
		CheckDestroy: testAccCheckJobQueueDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.95.0",
					},
				},
				Config: testAccJobQueueConfig_StateUpgradeComputeEnvironments_initial(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobQueueExists(ctx, t, resourceName, &jobQueue1),
				),
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccJobQueueConfig_StateUpgradeComputeEnvironments_updated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckJobQueueExists(ctx, t, resourceName, &jobQueue1),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func testAccCheckJobQueueExists(ctx context.Context, t *testing.T, n string, v *awstypes.JobQueueDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).BatchClient(ctx)

		output, err := tfbatch.FindJobQueueByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckJobQueueDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_batch_job_queue" {
				continue
			}

			conn := acctest.ProviderMeta(ctx, t).BatchClient(ctx)
			_, err := tfbatch.FindJobQueueByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Batch Job Queue %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

// testAccCheckJobQueueComputeEnvironmentOrderUpdate simulates the change of a Compute Environment Order
// An external update to the Batch Job Queue (e.g. console) may trigger changes to the value of the Order
// parameter that do not affect the operation of the queue itself, but the resource logic needs to handle.
// For example, Terraform may set a single Compute Environment with Order 0, but the console updates it to 1.
func testAccCheckJobQueueComputeEnvironmentOrderUpdate(ctx context.Context, t *testing.T, jobQueue *awstypes.JobQueueDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).BatchClient(ctx)

		input := &batch.UpdateJobQueueInput{
			ComputeEnvironmentOrder: jobQueue.ComputeEnvironmentOrder,
			JobQueue:                jobQueue.JobQueueName,
		}
		name := aws.ToString(jobQueue.JobQueueName)

		if len(input.ComputeEnvironmentOrder) != 1 {
			return fmt.Errorf("expected one ComputeEnvironmentOrder in Batch Job Queue (%s)", name)
		}

		if aws.ToInt32(input.ComputeEnvironmentOrder[0].Order) != 1 {
			return fmt.Errorf("expected first ComputeEnvironmentOrder in Batch Job Queue (%s) to have existing Order value of 1", name)
		}

		input.ComputeEnvironmentOrder[0].Order = aws.Int32(1)

		_, err := conn.UpdateJobQueue(ctx, input)

		if err != nil {
			return fmt.Errorf("error updating Batch Job Queue (%s): %w", name, err)
		}

		return nil
	}
}

func testAccJobQueueConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_service_principal" "batch" {
  service_name = "batch"
}

data "aws_service_principal" "ec2" {
  service_name = "ec2"
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
    {
        "Action": "sts:AssumeRole",
        "Effect": "Allow",
        "Principal": {
        "Service": "${data.aws_service_principal.batch.name}"
        }
    }
    ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSBatchServiceRole"
}

resource "aws_iam_role" "ecs_instance_role" {
  name = "%[1]s-ecs"

  assume_role_policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
    {
        "Action": "sts:AssumeRole",
        "Effect": "Allow",
        "Principal": {
        "Service": "${data.aws_service_principal.ec2.name}"
        }
    }
    ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "ecs_instance_role" {
  role       = aws_iam_role.ecs_instance_role.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonEC2ContainerServiceforEC2Role"
}

resource "aws_iam_instance_profile" "ecs_instance_role" {
  name = aws_iam_role.ecs_instance_role.name
  role = aws_iam_role_policy_attachment.ecs_instance_role.role
}

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.1.1.0/24"
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_batch_compute_environment" "test" {
  name         = %[1]q
  service_role = aws_iam_role.test.arn
  type         = "MANAGED"

  compute_resources {
    instance_role      = aws_iam_instance_profile.ecs_instance_role.arn
    instance_type      = ["c5", "m5", "r5"]
    max_vcpus          = 1
    min_vcpus          = 0
    security_group_ids = [aws_security_group.test.id]
    subnets            = [aws_subnet.test.id]
    type               = "EC2"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName)
}

func testAccJobQueueConfig_priority(rName string, priority int) string {
	return acctest.ConfigCompose(
		testAccJobQueueConfig_base(rName),
		fmt.Sprintf(`
resource "aws_batch_job_queue" "test" {
  compute_environment_order {
    compute_environment = aws_batch_compute_environment.test.arn
    order               = 1
  }
  name     = %[1]q
  priority = %[2]d
  state    = "ENABLED"
}
`, rName, priority))
}

func testAccJobQueueConfig_baseSchedulingPolicy(rName string, rName2 string) string {
	return fmt.Sprintf(`
resource "aws_batch_scheduling_policy" "test1" {
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

resource "aws_batch_scheduling_policy" "test2" {
  name = %[2]q

  fair_share_policy {
    compute_reservation = 1
    share_decay_seconds = 3600

    share_distribution {
      share_identifier = "A2"
      weight_factor    = 0.2
    }
  }
}
`, rName, rName2)
}

func testAccJobQueueConfig_schedulingPolicy(rName string, schedulingPolicyName1 string, schedulingPolicyName2 string, selectSchedulingPolicy string) string {
	return acctest.ConfigCompose(
		testAccJobQueueConfig_base(rName),
		testAccJobQueueConfig_baseSchedulingPolicy(schedulingPolicyName1, schedulingPolicyName2),
		fmt.Sprintf(`
locals {
  select_scheduling_policy = %[2]q
}

resource "aws_batch_job_queue" "test" {
  compute_environment_order {
    compute_environment = aws_batch_compute_environment.test.arn
    order               = 1
  }
  name                  = %[1]q
  priority              = 1
  scheduling_policy_arn = local.select_scheduling_policy == "first" ? aws_batch_scheduling_policy.test1.arn : aws_batch_scheduling_policy.test2.arn
  state                 = "ENABLED"
}
`, rName, selectSchedulingPolicy))
}

func testAccJobQueueConfig_state(rName string, state string) string {
	return acctest.ConfigCompose(
		testAccJobQueueConfig_base(rName),
		fmt.Sprintf(`
resource "aws_batch_job_queue" "test" {
  compute_environment_order {
    compute_environment = aws_batch_compute_environment.test.arn
    order               = 1
  }

  name     = %[1]q
  priority = 1
  state    = %[2]q
}
`, rName, state))
}

func testAccJobQueueConfig_ComputeEnvironments_multiple(rName string, state string) string {
	return acctest.ConfigCompose(
		testAccJobQueueConfig_base(rName),
		fmt.Sprintf(`
resource "aws_batch_job_queue" "test" {
  dynamic "compute_environment_order" {
    for_each = concat(
      [aws_batch_compute_environment.test.arn],
      aws_batch_compute_environment.more[*].arn,
    )
    content {
      compute_environment = compute_environment_order.value
      order               = compute_environment_order.key + 1
    }
  }
  name     = %[1]q
  priority = 1
  state    = %[2]q
}

resource "aws_batch_compute_environment" "more" {
  count = 2

  name         = "%[1]s-${count.index + 1}"
  service_role = aws_iam_role.test.arn
  type         = "MANAGED"

  compute_resources {
    instance_role      = aws_iam_instance_profile.ecs_instance_role.arn
    instance_type      = ["c5", "m5", "r5"]
    max_vcpus          = 1
    min_vcpus          = 0
    security_group_ids = [aws_security_group.test.id]
    subnets            = [aws_subnet.test.id]
    type               = "EC2"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, state))
}

func testAccJobQueueConfig_ComputeEnvironmentOrder_multiple(rName string, state string, o1 int, o2 int, o3 int) string {
	return acctest.ConfigCompose(
		testAccJobQueueConfig_base(rName),
		fmt.Sprintf(`
resource "aws_batch_job_queue" "test" {
  compute_environment_order {
    order               = %[3]d
    compute_environment = aws_batch_compute_environment.test.arn
  }

  compute_environment_order {
    order               = %[4]d
    compute_environment = aws_batch_compute_environment.more[0].arn
  }

  compute_environment_order {
    order               = %[5]d
    compute_environment = aws_batch_compute_environment.more[1].arn
  }

  name     = %[1]q
  priority = 1
  state    = %[2]q
}

resource "aws_batch_compute_environment" "more" {
  count = 2

  name         = "%[1]s-${count.index + 1}"
  service_role = aws_iam_role.test.arn
  type         = "MANAGED"

  compute_resources {
    instance_role      = aws_iam_instance_profile.ecs_instance_role.arn
    instance_type      = ["c5", "m5", "r5"]
    max_vcpus          = 1
    min_vcpus          = 0
    security_group_ids = [aws_security_group.test.id]
    subnets            = [aws_subnet.test.id]
    type               = "EC2"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, state, o1, o2, o3))
}

func testAccJobQueueConfig_ComputeEnvironments_multipleReorder(rName string, state string) string {
	return acctest.ConfigCompose(
		testAccJobQueueConfig_base(rName),
		fmt.Sprintf(`
resource "aws_batch_job_queue" "test" {
  compute_environment_order {
    compute_environment = aws_batch_compute_environment.more[0].arn
    order               = 1
  }
  compute_environment_order {
    compute_environment = aws_batch_compute_environment.test.arn
    order               = 2
  }
  compute_environment_order {
    compute_environment = aws_batch_compute_environment.more[1].arn
    order               = 3
  }
  name     = %[1]q
  priority = 1
  state    = %[2]q
}

resource "aws_batch_compute_environment" "more" {
  count = 2

  name         = "%[1]s-${count.index + 1}"
  service_role = aws_iam_role.test.arn
  type         = "MANAGED"

  compute_resources {
    instance_role      = aws_iam_instance_profile.ecs_instance_role.arn
    instance_type      = ["c5", "m5", "r5"]
    max_vcpus          = 1
    min_vcpus          = 0
    security_group_ids = [aws_security_group.test.id]
    subnets            = [aws_subnet.test.id]
    type               = "EC2"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName, state))
}

func testAccJobQueueConfig_jobStateTimeLimitAction(rName string) string {
	return acctest.ConfigCompose(
		testAccJobQueueConfig_base(rName),
		fmt.Sprintf(`
resource "aws_batch_job_queue" "test" {
  compute_environment_order {
    compute_environment = aws_batch_compute_environment.test.arn
    order               = 1
  }

  name     = %[1]q
  priority = 1
  state    = "DISABLED"
  job_state_time_limit_action {
    action           = "CANCEL"
    max_time_seconds = 600
    reason           = "MISCONFIGURATION:JOB_RESOURCE_REQUIREMENT"
    state            = "RUNNABLE"
  }
  job_state_time_limit_action {
    action           = "CANCEL"
    max_time_seconds = 605
    reason           = "CAPACITY:INSUFFICIENT_INSTANCE_CAPACITY"
    state            = "RUNNABLE"
  }
}
`, rName))
}

func testAccJobQueueConfig_jobStateTimeLimitActionUpdated(rName string) string {
	return acctest.ConfigCompose(
		testAccJobQueueConfig_base(rName),
		fmt.Sprintf(`
resource "aws_batch_job_queue" "test" {
  compute_environment_order {
    compute_environment = aws_batch_compute_environment.test.arn
    order               = 1
  }
  name     = %[1]q
  priority = 1
  state    = "DISABLED"
  job_state_time_limit_action {
    action           = "CANCEL"
    max_time_seconds = 610
    reason           = "MISCONFIGURATION:JOB_RESOURCE_REQUIREMENT"
    state            = "RUNNABLE"
  }
  job_state_time_limit_action {
    action           = "CANCEL"
    max_time_seconds = 605
    reason           = "MISCONFIGURATION:COMPUTE_ENVIRONMENT_MAX_RESOURCE"
    state            = "RUNNABLE"
  }
}
`, rName))
}

// testAccJobQueueConfig_base returns the base configuration for the Terraform AWS Provider v5.95.0.
// WARNING: Legacy configuration, breaking changes have been made to the aws_batch_compute_environment resource.
func testAccJobQueueConfig_legacyBase(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_service_principal" "batch" {
  service_name = "batch"
}

data "aws_service_principal" "ec2" {
  service_name = "ec2"
}

resource "aws_iam_role" "test" {
  name               = %[1]q
  assume_role_policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
    {
        "Action": "sts:AssumeRole",
        "Effect": "Allow",
        "Principal": {
        "Service": "${data.aws_service_principal.batch.name}"
        }
    }
    ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSBatchServiceRole"
}

resource "aws_iam_role" "ecs_instance_role" {
  name = "%[1]s-ecs"

  assume_role_policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
    {
        "Action": "sts:AssumeRole",
        "Effect": "Allow",
        "Principal": {
        "Service": "${data.aws_service_principal.ec2.name}"
        }
    }
    ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "ecs_instance_role" {
  role       = aws_iam_role.ecs_instance_role.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonEC2ContainerServiceforEC2Role"
}

resource "aws_iam_instance_profile" "ecs_instance_role" {
  name = aws_iam_role.ecs_instance_role.name
  role = aws_iam_role_policy_attachment.ecs_instance_role.role
}

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.1.1.0/24"
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_batch_compute_environment" "test" {
  compute_environment_name = %[1]q
  service_role             = aws_iam_role.test.arn
  type                     = "MANAGED"

  compute_resources {
    instance_role      = aws_iam_instance_profile.ecs_instance_role.arn
    instance_type      = ["c5", "m5", "r5"]
    max_vcpus          = 1
    min_vcpus          = 0
    security_group_ids = [aws_security_group.test.id]
    subnets            = [aws_subnet.test.id]
    type               = "EC2"
  }

  depends_on = [aws_iam_role_policy_attachment.test]
}
`, rName)
}

func testAccJobQueueConfig_StateUpgradeComputeEnvironments_initial(rName string) string {
	return acctest.ConfigCompose(
		testAccJobQueueConfig_legacyBase(rName),
		fmt.Sprintf(`
resource "aws_batch_job_queue" "test" {
  compute_environments = [aws_batch_compute_environment.test.arn]
  name                 = %[1]q
  priority             = 1
  state                = "ENABLED"
}
`, rName))
}

func testAccJobQueueConfig_StateUpgradeComputeEnvironments_updated(rName string) string {
	return acctest.ConfigCompose(
		testAccJobQueueConfig_base(rName),
		fmt.Sprintf(`
resource "aws_batch_job_queue" "test" {
  compute_environment_order {
    compute_environment = aws_batch_compute_environment.test.arn
    order               = 0
  }
  name     = %[1]q
  priority = 1
  state    = "ENABLED"
}
`, rName))
}
