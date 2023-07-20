// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/ecs"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccECSTaskDefinitionDataSource_ecsTaskDefinition(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ecs_task_definition.test"
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskDefinitionDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", "aws_ecs_task_definition.test", "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "execution_role_arn", "aws_iam_role.execution", "arn"),
					resource.TestCheckResourceAttr(dataSourceName, "family", rName),
					resource.TestCheckResourceAttr(dataSourceName, "network_mode", "bridge"),
					resource.TestMatchResourceAttr(dataSourceName, "revision", regexp.MustCompile("^[1-9][0-9]*$")),
					resource.TestCheckResourceAttr(dataSourceName, "status", "ACTIVE"),
					resource.TestCheckResourceAttrPair(dataSourceName, "task_role_arn", "aws_iam_role.test", "arn"),
				),
			},
		},
	})
}

func testAccTaskDefinitionDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [{
    "Action": "sts:AssumeRole",
    "Principal": {
      "Service": "ec2.amazonaws.com"
    },
    "Effect": "Allow",
    "Sid": ""
  }]
}
POLICY
}

resource "aws_iam_role" "execution" {
  name = "%[1]s-execution"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [{
    "Action": "sts:AssumeRole",
    "Principal": {
      "Service": "ec2.amazonaws.com"
    },
    "Effect": "Allow",
    "Sid": ""
  }]
}
POLICY
}

resource "aws_ecs_task_definition" "test" {
  family             = %[1]q
  execution_role_arn = aws_iam_role.execution.arn
  task_role_arn      = aws_iam_role.test.arn
  network_mode       = "bridge"

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "environment": [
      {
        "name": "SECRET",
        "value": "KEY"
      }
    ],
    "essential": true,
    "image": "mongo:latest",
    "memory": 128,
    "memoryReservation": 64,
    "name": "mongodb"
  }
]
DEFINITION
}

data "aws_ecs_task_definition" "test" {
  task_definition = aws_ecs_task_definition.test.family
}
`, rName)
}
