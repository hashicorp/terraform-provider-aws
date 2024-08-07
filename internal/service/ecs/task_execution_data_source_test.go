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

func TestAccECSTaskExecutionDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_ecs_task_execution.test"
	clusterName := "aws_ecs_cluster.test"
	taskDefinitionName := "aws_ecs_task_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ECSEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskExecutionDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "cluster", clusterName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, "task_definition", taskDefinitionName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, "client_token", "some_token"),
					resource.TestCheckResourceAttr(dataSourceName, "desired_count", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "launch_type", "FARGATE"),
					resource.TestCheckResourceAttr(dataSourceName, "network_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "task_arns.#", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccECSTaskExecutionDataSource_overrides(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_ecs_task_execution.test"
	clusterName := "aws_ecs_cluster.test"
	taskDefinitionName := "aws_ecs_task_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ECSEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskExecutionDataSourceConfig_overrides(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "cluster", clusterName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, "task_definition", taskDefinitionName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, "desired_count", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "launch_type", "FARGATE"),
					resource.TestCheckResourceAttr(dataSourceName, "network_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "task_arns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "overrides.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "overrides.0.container_overrides.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "overrides.0.container_overrides.0.environment.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "overrides.0.container_overrides.0.environment.0.key", acctest.CtKey1),
					resource.TestCheckResourceAttr(dataSourceName, "overrides.0.container_overrides.0.environment.0.value", acctest.CtValue1),
				),
			},
		},
	})
}

func TestAccECSTaskExecutionDataSource_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_ecs_task_execution.test"
	clusterName := "aws_ecs_cluster.test"
	taskDefinitionName := "aws_ecs_task_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ECSEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskExecutionDataSourceConfig_tags(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "cluster", clusterName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, "task_definition", taskDefinitionName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, "desired_count", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "launch_type", "FARGATE"),
					resource.TestCheckResourceAttr(dataSourceName, "network_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "task_arns.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
		},
	})
}

func testAccTaskExecutionDataSourceConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  ingress {
    protocol    = "6"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = [aws_vpc.test.cidr_block]
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_cluster_capacity_providers" "test" {
  cluster_name       = aws_ecs_cluster.test.name
  capacity_providers = ["FARGATE"]
}

resource "aws_ecs_task_definition" "test" {
  family                   = %[1]q
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "256"
  memory                   = "512"

  container_definitions = jsonencode([
    {
      name      = "sleep"
      image     = "busybox"
      cpu       = 10
      command   = ["sleep", "10"]
      memory    = 10
      essential = true
      portMappings = [
        {
          protocol      = "tcp"
          containerPort = 8000
        }
      ]
    }
  ])
}
`, rName)
}

func testAccTaskExecutionDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
		testAccTaskExecutionDataSourceConfig_base(rName),
		`
data "aws_ecs_task_execution" "test" {
  depends_on = [aws_ecs_cluster_capacity_providers.test]

  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  client_token    = "some_token"
  desired_count   = 1
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = aws_subnet.test[*].id
    security_groups  = [aws_security_group.test.id]
    assign_public_ip = false
  }
}
`)
}

func testAccTaskExecutionDataSourceConfig_tags(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
		testAccTaskExecutionDataSourceConfig_base(rName),
		fmt.Sprintf(`
data "aws_ecs_task_execution" "test" {
  depends_on = [aws_ecs_cluster_capacity_providers.test]

  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = aws_subnet.test[*].id
    security_groups  = [aws_security_group.test.id]
    assign_public_ip = false
  }

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccTaskExecutionDataSourceConfig_overrides(rName, envKey1, envValue1 string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
		testAccTaskExecutionDataSourceConfig_base(rName),
		fmt.Sprintf(`
data "aws_ecs_task_execution" "test" {
  depends_on = [aws_ecs_cluster_capacity_providers.test]

  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1
  launch_type     = "FARGATE"

  network_configuration {
    subnets          = aws_subnet.test[*].id
    security_groups  = [aws_security_group.test.id]
    assign_public_ip = false
  }

  overrides {
    container_overrides {
      name = "sleep"

      environment {
        key   = %[1]q
        value = %[2]q
      }
    }
    cpu    = "256"
    memory = "512"
  }
}
`, envKey1, envValue1))
}
