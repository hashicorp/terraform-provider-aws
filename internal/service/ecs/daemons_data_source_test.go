// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccECSDaemonsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_ecs_daemons.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDaemonsDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "daemons.#", "1"),
				),
			},
		},
	})
}

func TestAccECSDaemonsDataSource_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_ecs_daemons.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDaemonsDataSourceConfig_multiple(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "daemons.#", "2"),
				),
			},
		},
	})
}

func TestAccECSDaemonsDataSource_empty(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_ecs_daemons.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDaemonsDataSourceConfig_empty(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "daemons.#", "0"),
				),
			},
		},
	})
}

func testAccDaemonsDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccDaemonConfig_basic(rName),
		`
data "aws_ecs_daemons" "test" {
  cluster = aws_ecs_cluster.test.arn

  depends_on = [aws_ecs_daemon.test]
}
`)
}

func testAccDaemonsDataSourceConfig_multiple(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 1),
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_capacity_provider" "test" {
  name    = %[1]q
  cluster = aws_ecs_cluster.test.name

  managed_instances_provider {
    infrastructure_role_arn = aws_iam_role.infra.arn

    instance_launch_template {
      ec2_instance_profile_arn = aws_iam_instance_profile.test.arn

      network_configuration {
        subnets         = [aws_subnet.test[0].id]
        security_groups = [aws_security_group.test.id]
      }
    }
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_iam_role" "infra" {
  name = "%[1]s-infra"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = { Service = "ecs.${data.aws_partition.current.dns_suffix}" }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "infra" {
  role       = aws_iam_role.infra.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonECSInfrastructureRolePolicyForManagedInstances"
}

resource "aws_iam_role" "instance" {
  name = "%[1]s-instance"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = { Service = "ec2.${data.aws_partition.current.dns_suffix}" }
    }]
  })
}

resource "aws_iam_role_policy_attachment" "instance" {
  role       = aws_iam_role.instance.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonEC2ContainerServiceforEC2Role"
}

resource "aws_iam_instance_profile" "test" {
  name = %[1]q
  role = aws_iam_role.instance.name
}

resource "aws_ecs_daemon_task_definition" "test" {
  family = %[1]q

  container_definition {
    name      = "test"
    image     = "nginx:latest"
    cpu       = 256
    memory    = 512
    essential = true
  }

  cpu    = "256"
  memory = "512"
}

resource "aws_ecs_daemon" "test1" {
  name                   = "%[1]s-1"
  cluster                = aws_ecs_cluster.test.arn
  daemon_task_definition = aws_ecs_daemon_task_definition.test.arn
  capacity_provider_arns = [aws_ecs_capacity_provider.test.arn]
}

resource "aws_ecs_daemon" "test2" {
  name                   = "%[1]s-2"
  cluster                = aws_ecs_cluster.test.arn
  daemon_task_definition = aws_ecs_daemon_task_definition.test.arn
  capacity_provider_arns = [aws_ecs_capacity_provider.test.arn]
}

data "aws_ecs_daemons" "test" {
  cluster = aws_ecs_cluster.test.arn

  depends_on = [aws_ecs_daemon.test1, aws_ecs_daemon.test2]
}
`, rName))
}

func testAccDaemonsDataSourceConfig_empty(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

data "aws_ecs_daemons" "test" {
  cluster = aws_ecs_cluster.test.arn
}
`, rName)
}
