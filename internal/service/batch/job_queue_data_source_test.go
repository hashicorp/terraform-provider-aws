// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBatchJobQueueDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix("tf_acc_test_")
	resourceName := "aws_batch_job_queue.test"
	dataSourceName := "data.aws_batch_job_queue.by_name"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccJobQueueDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "compute_environment_order.#", resourceName, "compute_environments.#"),
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
	rName := sdkacctest.RandomWithPrefix("tf_acc_test_")
	resourceName := "aws_batch_job_queue.test"
	dataSourceName := "data.aws_batch_job_queue.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccJobQueueDataSourceConfig_schedulingPolicy(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "compute_environment_order.#", resourceName, "compute_environments.#"),
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

func testAccJobQueueDataSourceConfigBase(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "ecs_instance_role" {
  name = "ecs_%[1]s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.${data.aws_partition.current.dns_suffix}"
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
  name = "ecs_%[1]s"
  role = aws_iam_role.ecs_instance_role.name
}

resource "aws_iam_role" "aws_batch_service_role" {
  name = "batch_%[1]s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Effect": "Allow",
      "Principal": {
        "Service": "batch.${data.aws_partition.current.dns_suffix}"
      }
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "aws_batch_service_role" {
  role       = aws_iam_role.aws_batch_service_role.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AWSBatchServiceRole"
}

resource "aws_security_group" "sample" {
  name = "%[1]s"
}

resource "aws_vpc" "sample" {
  cidr_block = "10.1.0.0/16"
}

resource "aws_subnet" "sample" {
  vpc_id     = aws_vpc.sample.id
  cidr_block = "10.1.1.0/24"
}

resource "aws_batch_compute_environment" "sample" {
  compute_environment_name = "%[1]s"

  compute_resources {
    instance_role = aws_iam_instance_profile.ecs_instance_role.arn

    instance_type = [
      "c4.large",
    ]

    max_vcpus = 16
    min_vcpus = 0

    security_group_ids = [
      aws_security_group.sample.id,
    ]

    subnets = [
      aws_subnet.sample.id,
    ]

    type = "EC2"
  }

  service_role = aws_iam_role.aws_batch_service_role.arn
  type         = "MANAGED"
  depends_on   = [aws_iam_role_policy_attachment.aws_batch_service_role]
}
`, rName)
}

func testAccJobQueueDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccJobQueueDataSourceConfigBase(rName),
		fmt.Sprintf(`
resource "aws_batch_job_queue" "test" {
  name                 = "%[1]s"
  state                = "ENABLED"
  priority             = 1
  compute_environments = [aws_batch_compute_environment.sample.arn]
}

resource "aws_batch_job_queue" "wrong" {
  name                 = "%[1]s_wrong"
  state                = "ENABLED"
  priority             = 2
  compute_environments = [aws_batch_compute_environment.sample.arn]
}

data "aws_batch_job_queue" "by_name" {
  name = aws_batch_job_queue.test.name
}
`, rName))
}

func testAccJobQueueDataSourceConfig_schedulingPolicy(rName string) string {
	return acctest.ConfigCompose(
		testAccJobQueueDataSourceConfigBase(rName),
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
  compute_environments  = [aws_batch_compute_environment.sample.arn]
}

data "aws_batch_job_queue" "test" {
  name = aws_batch_job_queue.test.name
}
`, rName))
}
