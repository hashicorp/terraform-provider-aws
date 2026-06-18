// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecs_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccECSServiceDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ecs_service.test"
	resourceName := "aws_ecs_service.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "availability_zone_rebalancing", dataSourceName, "availability_zone_rebalancing"),
					resource.TestCheckResourceAttrPair(resourceName, "desired_count", dataSourceName, "desired_count"),
					resource.TestCheckResourceAttrPair(resourceName, "launch_type", dataSourceName, "launch_type"),
					resource.TestCheckResourceAttrPair(resourceName, "scheduling_strategy", dataSourceName, "scheduling_strategy"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrName, dataSourceName, names.AttrServiceName),
					resource.TestCheckResourceAttrPair(resourceName, "task_definition", dataSourceName, "task_definition"),
					resource.TestCheckResourceAttrPair(resourceName, acctest.CtTagsPercent, dataSourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(resourceName, "tags.Name", dataSourceName, "tags.Name"),
				),
			},
		},
	})
}

func TestAccECSServiceDataSource_loadBalancer(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ecs_service.test"
	resourceName := "aws_ecs_service.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)[:16]

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDataSourceConfig_loadBalancer(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "availability_zone_rebalancing", dataSourceName, "availability_zone_rebalancing"),
					resource.TestCheckResourceAttrPair(resourceName, "desired_count", dataSourceName, "desired_count"),
					resource.TestCheckResourceAttrPair(resourceName, "launch_type", dataSourceName, "launch_type"),
					resource.TestCheckResourceAttrPair(resourceName, "load_balancer", dataSourceName, "load_balancer"),
					resource.TestCheckResourceAttrPair(resourceName, "scheduling_strategy", dataSourceName, "scheduling_strategy"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrName, dataSourceName, names.AttrServiceName),
					resource.TestCheckResourceAttrPair(resourceName, "task_definition", dataSourceName, "task_definition"),
					resource.TestCheckResourceAttrPair(resourceName, acctest.CtTagsPercent, dataSourceName, acctest.CtTagsPercent),
				),
			},
		},
	})
}

func TestAccECSServiceDataSource_deploymentConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ecs_service.test"
	resourceName := "aws_ecs_service.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)[:16]

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDataSourceConfig_linearDeployment(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "deployment_configuration.#", dataSourceName, "deployment_configuration.#"),
					resource.TestCheckResourceAttrPair(resourceName, "deployment_configuration.0.strategy", dataSourceName, "deployment_configuration.0.strategy"),
					resource.TestCheckResourceAttr(dataSourceName, "deployment_configuration.0.strategy", "LINEAR"),
					resource.TestCheckResourceAttrPair(resourceName, "deployment_configuration.0.bake_time_in_minutes", dataSourceName, "deployment_configuration.0.bake_time_in_minutes"),
					resource.TestCheckResourceAttrPair(resourceName, "deployment_configuration.0.linear_configuration.#", dataSourceName, "deployment_configuration.0.linear_configuration.#"),
					resource.TestCheckResourceAttrPair(resourceName, "deployment_configuration.0.linear_configuration.0.step_percent", dataSourceName, "deployment_configuration.0.linear_configuration.0.step_percent"),
					resource.TestCheckResourceAttrPair(resourceName, "deployment_configuration.0.linear_configuration.0.step_bake_time_in_minutes", dataSourceName, "deployment_configuration.0.linear_configuration.0.step_bake_time_in_minutes"),
					// Resource has these at top level, data source has them in deployment_configuration
					resource.TestCheckResourceAttrPair(resourceName, "deployment_maximum_percent", dataSourceName, "deployment_configuration.0.maximum_percent"),
					resource.TestCheckResourceAttrPair(resourceName, "deployment_minimum_healthy_percent", dataSourceName, "deployment_configuration.0.minimum_healthy_percent"),
					resource.TestCheckResourceAttrPair(resourceName, "deployment_circuit_breaker.#", dataSourceName, "deployment_configuration.0.deployment_circuit_breaker.#"),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrSet(dataSourceName, "created_by"),
					resource.TestCheckResourceAttrSet(dataSourceName, "pending_count"),
					resource.TestCheckResourceAttrSet(dataSourceName, "running_count"),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrStatus),
				),
			},
		},
	})
}

func TestAccECSServiceDataSource_canaryDeployment(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ecs_service.test"
	resourceName := "aws_ecs_service.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)[:16]

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDataSourceConfig_canaryDeployment(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "deployment_configuration.#", dataSourceName, "deployment_configuration.#"),
					resource.TestCheckResourceAttrPair(resourceName, "deployment_configuration.0.strategy", dataSourceName, "deployment_configuration.0.strategy"),
					resource.TestCheckResourceAttr(dataSourceName, "deployment_configuration.0.strategy", "CANARY"),
					resource.TestCheckResourceAttrPair(resourceName, "deployment_configuration.0.bake_time_in_minutes", dataSourceName, "deployment_configuration.0.bake_time_in_minutes"),
					resource.TestCheckResourceAttrPair(resourceName, "deployment_configuration.0.canary_configuration.#", dataSourceName, "deployment_configuration.0.canary_configuration.#"),
					resource.TestCheckResourceAttrPair(resourceName, "deployment_configuration.0.canary_configuration.0.canary_percent", dataSourceName, "deployment_configuration.0.canary_configuration.0.canary_percent"),
					resource.TestCheckResourceAttrPair(resourceName, "deployment_configuration.0.canary_configuration.0.canary_bake_time_in_minutes", dataSourceName, "deployment_configuration.0.canary_configuration.0.canary_bake_time_in_minutes"),
				),
			},
		},
	})
}

func TestAccECSServiceDataSource_fullConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ecs_service.test"
	resourceName := "aws_ecs_service.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)[:16]

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceDataSourceConfig_fullConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "deployment_controller.#", dataSourceName, "deployment_controller.#"),
					resource.TestCheckResourceAttrPair(resourceName, "enable_ecs_managed_tags", dataSourceName, "enable_ecs_managed_tags"),
					resource.TestCheckResourceAttrPair(resourceName, "platform_version", dataSourceName, "platform_version"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrPropagateTags, dataSourceName, names.AttrPropagateTags),
					// Resource has these at top level, data source has them in deployment_configuration
					resource.TestCheckResourceAttrPair(resourceName, "deployment_maximum_percent", dataSourceName, "deployment_configuration.0.maximum_percent"),
					resource.TestCheckResourceAttrPair(resourceName, "deployment_minimum_healthy_percent", dataSourceName, "deployment_configuration.0.minimum_healthy_percent"),
					resource.TestCheckResourceAttrSet(dataSourceName, "deployments.#"),
					resource.TestCheckResourceAttrSet(dataSourceName, "events.#"),
					resource.TestCheckResourceAttrSet(dataSourceName, "platform_family"),
				),
			},
		},
	})
}

func testAccServiceDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "mongo:latest",
    "memory": 128,
    "memoryReservation": 64,
    "name": "mongodb"
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  name            = "mongodb"
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1
  tags = {
    Name = %[1]q
  }
}

data "aws_ecs_service" "test" {
  service_name = aws_ecs_service.test.name
  cluster_arn  = aws_ecs_cluster.test.arn
}
`, rName)
}

func testAccServiceDataSourceConfig_loadBalancer(rName string) string {
	return acctest.ConfigCompose(
		testAccServiceConfig_blueGreenDeployment_basic(rName, false),
		`
data "aws_ecs_service" "test" {
  service_name = aws_ecs_service.test.name
  cluster_arn  = aws_ecs_cluster.main.arn
}
`)
}

func testAccServiceDataSourceConfig_linearDeployment(rName string) string {
	return acctest.ConfigCompose(
		testAccServiceConfig_linearDeployment_basic(rName, false),
		`
data "aws_ecs_service" "test" {
  service_name = aws_ecs_service.test.name
  cluster_arn  = aws_ecs_cluster.main.arn
}
`)
}

func testAccServiceDataSourceConfig_canaryDeployment(rName string) string {
	return acctest.ConfigCompose(
		testAccServiceConfig_canaryDeployment_basic(rName, false),
		`
data "aws_ecs_service" "test" {
  service_name = aws_ecs_service.test.name
  cluster_arn  = aws_ecs_cluster.main.arn
}
`)
}

func testAccServiceDataSourceConfig_fullConfiguration(rName string) string {
	return acctest.ConfigCompose(testAccServiceConfig_launchTypeFargateBase(rName), fmt.Sprintf(`
resource "aws_ecs_service" "test" {
  name             = %[1]q
  cluster          = aws_ecs_cluster.test.id
  task_definition  = aws_ecs_task_definition.test.arn
  desired_count    = 1
  launch_type      = "FARGATE"
  platform_version = "LATEST"

  enable_ecs_managed_tags = true
  propagate_tags          = "SERVICE"

  deployment_maximum_percent         = 200
  deployment_minimum_healthy_percent = 100

  deployment_controller {
    type = "ECS"
  }

  network_configuration {
    security_groups  = aws_security_group.test[*].id
    subnets          = aws_subnet.test[*].id
    assign_public_ip = true
  }

  tags = {
    Name = %[1]q
  }
}

data "aws_ecs_service" "test" {
  service_name = aws_ecs_service.test.name
  cluster_arn  = aws_ecs_cluster.test.arn
}
`, rName))
}
