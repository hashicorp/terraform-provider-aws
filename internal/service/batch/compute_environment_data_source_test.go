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

func TestAccBatchComputeEnvironmentDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix("tf_acc_test_")
	resourceName := "aws_batch_compute_environment.test"
	dataSourceName := "data.aws_batch_compute_environment.by_name"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeEnvironmentDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "compute_environment_name", resourceName, "compute_environment_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ecs_cluster_arn", resourceName, "ecs_cluster_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrServiceRole, resourceName, names.AttrServiceRole),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrState, resourceName, names.AttrState),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrType, resourceName, names.AttrType),
					resource.TestCheckResourceAttr(dataSourceName, "update_policy.#", acctest.Ct0),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{})),
				},
			},
		},
	})
}

func TestAccBatchComputeEnvironmentDataSource_basicUpdatePolicy(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix("tf_acc_test_")
	resourceName := "aws_batch_compute_environment.test"
	dataSourceName := "data.aws_batch_compute_environment.by_name"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.BatchServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccComputeEnvironmentDataSourceConfig_updatePolicy(rName, 30, false),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, "update_policy.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "update_policy.0.%", acctest.Ct2),
					resource.TestCheckResourceAttr(dataSourceName, "update_policy.0.terminate_jobs_on_update", acctest.CtFalse),
					resource.TestCheckResourceAttr(dataSourceName, "update_policy.0.job_execution_timeout_minutes", "30"),
				),
			},
		},
	})
}

func testAccComputeEnvironmentDataSourceConfig_basic(rName string) string {
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

resource "aws_batch_compute_environment" "test" {
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

  type = "MANAGED"
}

data "aws_batch_compute_environment" "by_name" {
  compute_environment_name = aws_batch_compute_environment.test.compute_environment_name
}
`, rName)
}

func testAccComputeEnvironmentDataSourceConfig_updatePolicy(rName string, timeout int, terminate bool) string {
	return acctest.ConfigCompose(testAccComputeEnvironmentConfig_baseDefaultSLR(rName), fmt.Sprintf(`
resource "aws_batch_compute_environment" "test" {
  compute_environment_name = %[1]q

  compute_resources {
    allocation_strategy = "BEST_FIT_PROGRESSIVE"
    instance_role       = aws_iam_instance_profile.ecs_instance.arn
    instance_type       = ["optimal"]
    max_vcpus           = 4
    min_vcpus           = 0
    security_group_ids = [
      aws_security_group.test.id
    ]
    subnets = [
      aws_subnet.test.id
    ]
    type = "EC2"
  }
  update_policy {
    job_execution_timeout_minutes = %[2]d
    terminate_jobs_on_update      = %[3]t
  }

  type = "MANAGED"
}

data "aws_batch_compute_environment" "by_name" {
  compute_environment_name = aws_batch_compute_environment.test.compute_environment_name
}
`, rName, timeout, terminate))
}
