// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecs_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccECSTaskDefinitionDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ecs_task_definition.test"
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(5))

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskDefinitionDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, "aws_ecs_task_definition.test", names.AttrARN),
					resource.TestCheckResourceAttrSet(dataSourceName, "container_definitions"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrExecutionRoleARN, "aws_iam_role.execution", names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrFamily, rName),
					resource.TestCheckResourceAttr(dataSourceName, "network_mode", "bridge"),
					resource.TestMatchResourceAttr(dataSourceName, "revision", regexache.MustCompile("^[1-9][0-9]*$")),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrStatus, "ACTIVE"),
					resource.TestCheckResourceAttrPair(dataSourceName, "task_role_arn", "aws_iam_role.test", names.AttrARN),
				),
			},
		},
	})
}

func TestAccECSTaskDefinitionDataSource_ec2(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ecs_task_definition.test"
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(5))

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskDefinitionDataSourceConfig_ec2(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, "aws_ecs_task_definition.test", names.AttrARN),
					resource.TestCheckResourceAttrSet(dataSourceName, "container_definitions"),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrFamily, rName),
					resource.TestCheckResourceAttr(dataSourceName, "ipc_mode", "host"),
					resource.TestCheckResourceAttr(dataSourceName, "network_mode", "awsvpc"),
					resource.TestCheckResourceAttr(dataSourceName, "pid_mode", "host"),
					resource.TestCheckResourceAttr(dataSourceName, "placement_constraints.#", "1"),
					resource.TestCheckResourceAttrSet(dataSourceName, "placement_constraints.0.expression"),
					resource.TestCheckResourceAttr(dataSourceName, "placement_constraints.0.type", "memberOf"),
					resource.TestCheckResourceAttr(dataSourceName, "requires_compatibilities.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "requires_compatibilities.0", "EC2"),
					resource.TestMatchResourceAttr(dataSourceName, "revision", regexache.MustCompile("^[1-9][0-9]*$")),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrStatus, "ACTIVE"),
					resource.TestCheckResourceAttr(dataSourceName, "volume.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "volume.0.name", "service-storage"),
					resource.TestCheckResourceAttr(dataSourceName, "volume.0.docker_volume_configuration.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "volume.0.docker_volume_configuration.0.scope", "shared"),
					resource.TestCheckResourceAttr(dataSourceName, "volume.0.docker_volume_configuration.0.autoprovision", acctest.CtTrue),
					resource.TestCheckResourceAttr(dataSourceName, "volume.0.docker_volume_configuration.0.driver", "local"),
					resource.TestCheckResourceAttr(dataSourceName, "volume.0.docker_volume_configuration.0.driver_opts.type", "nfs"),
					resource.TestCheckResourceAttr(dataSourceName, "volume.0.docker_volume_configuration.0.driver_opts.device", "test-efs.internal:/"),
					resource.TestCheckResourceAttr(dataSourceName, "volume.0.docker_volume_configuration.0.driver_opts.o", "addr=test-efs.internal,rsize=1048576,wsize=1048576,hard,timeo=600,retrans=2,noresvport"),
				),
			},
		},
	})
}

func TestAccECSTaskDefinitionDataSource_fargate(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ecs_task_definition.test"
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(5))

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskDefinitionDataSourceConfig_fargate(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, "aws_ecs_task_definition.test", names.AttrARN),
					resource.TestCheckResourceAttrSet(dataSourceName, "container_definitions"),
					resource.TestCheckResourceAttr(dataSourceName, "cpu", "1024"),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrFamily, rName),
					resource.TestCheckResourceAttr(dataSourceName, "enable_fault_injection", acctest.CtTrue),
					resource.TestCheckResourceAttr(dataSourceName, "ephemeral_storage.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "ephemeral_storage.0.size_in_gib", "30"),
					resource.TestCheckResourceAttr(dataSourceName, "memory", "8192"),
					resource.TestCheckResourceAttr(dataSourceName, "network_mode", "awsvpc"),
					resource.TestCheckResourceAttr(dataSourceName, "requires_compatibilities.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "requires_compatibilities.0", "FARGATE"),
					resource.TestMatchResourceAttr(dataSourceName, "revision", regexache.MustCompile("^[1-9][0-9]*$")),
					resource.TestCheckResourceAttr(dataSourceName, "runtime_platform.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "runtime_platform.0.cpu_architecture", "X86_64"),
					resource.TestCheckResourceAttr(dataSourceName, "runtime_platform.0.operating_system_family", "WINDOWS_SERVER_2022_CORE"),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrStatus, "ACTIVE"),
				),
			},
		},
	})
}

func TestAccECSTaskDefinitionDataSource_proxyConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ecs_task_definition.test"
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(5))

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskDefinitionDataSourceConfig_proxyConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, "aws_ecs_task_definition.test", names.AttrARN),
					resource.TestCheckResourceAttrSet(dataSourceName, "container_definitions"),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrFamily, rName),
					resource.TestCheckResourceAttr(dataSourceName, "network_mode", "awsvpc"),
					resource.TestCheckResourceAttr(dataSourceName, "proxy_configuration.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "proxy_configuration.0.type", "APPMESH"),
					resource.TestCheckResourceAttr(dataSourceName, "proxy_configuration.0.container_name", "web"),
					resource.TestCheckResourceAttr(dataSourceName, "proxy_configuration.0.properties.AppPorts", "80"),
					resource.TestCheckResourceAttr(dataSourceName, "proxy_configuration.0.properties.EgressIgnoredIPs", "169.254.170.2,169.254.169.254"),
					resource.TestCheckResourceAttr(dataSourceName, "proxy_configuration.0.properties.EgressIgnoredPorts", "5500"),
					resource.TestCheckResourceAttr(dataSourceName, "proxy_configuration.0.properties.IgnoredGID", "999"),
					resource.TestCheckResourceAttr(dataSourceName, "proxy_configuration.0.properties.IgnoredUID", "1337"),
					resource.TestCheckResourceAttr(dataSourceName, "proxy_configuration.0.properties.ProxyEgressPort", "15001"),
					resource.TestCheckResourceAttr(dataSourceName, "proxy_configuration.0.properties.ProxyIngressPort", "15000"),
					resource.TestMatchResourceAttr(dataSourceName, "revision", regexache.MustCompile("^[1-9][0-9]*$")),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrStatus, "ACTIVE"),
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
  execution_role_arn = aws_iam_role.execution.arn
  family             = %[1]q
  network_mode       = "bridge"
  task_role_arn      = aws_iam_role.test.arn

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

func testAccTaskDefinitionDataSourceConfig_ec2(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_ecs_task_definition" "test" {
  family                   = %[1]q
  ipc_mode                 = "host"
  network_mode             = "awsvpc"
  pid_mode                 = "host"
  requires_compatibilities = ["EC2"]

  container_definitions = <<TASK_DEFINITION
[
  {
    "name": "sleep",
    "image": "busybox",
    "cpu": 256,
    "command": ["sleep", "360"],
    "memory": 2048,
    "essential": true,
    "portMappings": [{"protocol": "tcp", "containerPort": 8000}]
  }
]
TASK_DEFINITION

  placement_constraints {
    expression = "attribute:ecs.availability-zone in [${data.aws_availability_zones.available.names[0]}, ${data.aws_availability_zones.available.names[1]}]"
    type       = "memberOf"
  }

  volume {
    name = "service-storage"

    docker_volume_configuration {
      scope         = "shared"
      autoprovision = true
      driver        = "local"

      driver_opts = {
        "type"   = "nfs"
        "device" = "test-efs.internal:/"
        "o"      = "addr=test-efs.internal,rsize=1048576,wsize=1048576,hard,timeo=600,retrans=2,noresvport"
      }
    }
  }
}

data "aws_ecs_task_definition" "test" {
  task_definition = aws_ecs_task_definition.test.family
}
`, rName)
}

func testAccTaskDefinitionDataSourceConfig_fargate(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_task_definition" "test" {
  cpu                      = "1024"
  enable_fault_injection   = true
  family                   = %[1]q
  memory                   = "8192"
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]

  container_definitions = <<TASK_DEFINITION
[
  {
    "name": "sleep",
    "image": "busybox",
    "cpu": 256,
    "command": ["sleep", "360"],
    "memory": 2048,
    "essential": true,
    "portMappings": [{"protocol": "tcp", "containerPort": 8000}]
  }
]
TASK_DEFINITION

  ephemeral_storage {
    size_in_gib = 30
  }

  runtime_platform {
    cpu_architecture        = "X86_64"
    operating_system_family = "WINDOWS_SERVER_2022_CORE"
  }
}

data "aws_ecs_task_definition" "test" {
  task_definition = aws_ecs_task_definition.test.family
}
`, rName)
}

func testAccTaskDefinitionDataSourceConfig_proxyConfiguration(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_task_definition" "test" {
  family       = %[1]q
  network_mode = "awsvpc"

  proxy_configuration {
    type           = "APPMESH"
    container_name = "web"
    properties = {
      AppPorts           = "80"
      EgressIgnoredPorts = "5500"
      EgressIgnoredIPs   = "169.254.170.2,169.254.169.254"
      IgnoredGID         = "999"
      IgnoredUID         = "1337"
      ProxyIngressPort   = "15000"
      ProxyEgressPort    = "15001"
    }
  }

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "nginx:latest",
    "memory": 128,
    "name": "web"
  }
]
DEFINITION
}

data "aws_ecs_task_definition" "test" {
  task_definition = aws_ecs_task_definition.test.family
}
`, rName)
}
