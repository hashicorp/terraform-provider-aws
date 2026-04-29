// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecs_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccECSDaemonTaskDefinitionDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ecs_daemon_task_definition.test"
	resourceName := "aws_ecs_daemon_task_definition.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDaemonTaskDefinitionDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, "container_definition.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "container_definition.0.name", "app"),
					resource.TestCheckResourceAttr(dataSourceName, "container_definition.0.image", "nginx:latest"),
					resource.TestCheckResourceAttr(dataSourceName, "container_definition.0.cpu", "256"),
					resource.TestCheckResourceAttr(dataSourceName, "container_definition.0.memory", "512"),
					resource.TestCheckResourceAttr(dataSourceName, "container_definition.0.essential", acctest.CtTrue),
					resource.TestCheckResourceAttr(dataSourceName, "cpu", "256"),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrFamily, rName),
					resource.TestCheckResourceAttr(dataSourceName, "memory", "512"),
					resource.TestCheckResourceAttrSet(dataSourceName, "registered_at"),
					resource.TestCheckResourceAttrSet(dataSourceName, "registered_by"),
					resource.TestMatchResourceAttr(dataSourceName, "revision", regexache.MustCompile("^[1-9][0-9]*$")),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrStatus, "ACTIVE"),
				),
			},
		},
	})
}

func TestAccECSDaemonTaskDefinitionDataSource_withVolumes(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ecs_daemon_task_definition.test"
	resourceName := "aws_ecs_daemon_task_definition.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDaemonTaskDefinitionDataSourceConfig_withVolumes(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, "container_definition.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrFamily, rName),
					resource.TestMatchResourceAttr(dataSourceName, "revision", regexache.MustCompile("^[1-9][0-9]*$")),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrStatus, "ACTIVE"),
					resource.TestCheckResourceAttr(dataSourceName, "volume.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "volume.*", map[string]string{
						names.AttrName: "data-volume",
						"host_path":    "/data",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "volume.*", map[string]string{
						names.AttrName: "logs-volume",
						"host_path":    "/var/log",
					}),
				),
			},
		},
	})
}

func TestAccECSDaemonTaskDefinitionDataSource_withRoles(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ecs_daemon_task_definition.test"
	resourceName := "aws_ecs_daemon_task_definition.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDaemonTaskDefinitionDataSourceConfig_withRoles(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, "container_definition.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrExecutionRoleARN, "aws_iam_role.execution", names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrFamily, rName),
					resource.TestMatchResourceAttr(dataSourceName, "revision", regexache.MustCompile("^[1-9][0-9]*$")),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrStatus, "ACTIVE"),
					resource.TestCheckResourceAttrPair(dataSourceName, "task_role_arn", "aws_iam_role.task", names.AttrARN),
				),
			},
		},
	})
}

func testAccDaemonTaskDefinitionDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_daemon_task_definition" "test" {
  family = %[1]q
  cpu    = "256"
  memory = "512"

  container_definition {
    name      = "app"
    image     = "nginx:latest"
    cpu       = 256
    memory    = 512
    essential = true
  }
}

data "aws_ecs_daemon_task_definition" "test" {
  daemon_task_definition = aws_ecs_daemon_task_definition.test.arn
}
`, rName)
}

func testAccDaemonTaskDefinitionDataSourceConfig_withVolumes(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_daemon_task_definition" "test" {
  family = %[1]q

  container_definition {
    name      = "app"
    image     = "nginx:latest"
    cpu       = 256
    memory    = 512
    essential = true
  }

  volume {
    name      = "data-volume"
    host_path = "/data"
  }

  volume {
    name      = "logs-volume"
    host_path = "/var/log"
  }
}

data "aws_ecs_daemon_task_definition" "test" {
  daemon_task_definition = aws_ecs_daemon_task_definition.test.arn
}
`, rName)
}

func testAccDaemonTaskDefinitionDataSourceConfig_withRoles(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "task" {
  name = "%[1]s-task"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ecs-tasks.amazonaws.com"
      }
    }]
  })
}

resource "aws_iam_role" "execution" {
  name = "%[1]s-execution"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "ecs-tasks.amazonaws.com"
      }
    }]
  })
}

resource "aws_ecs_daemon_task_definition" "test" {
  family             = %[1]q
  execution_role_arn = aws_iam_role.execution.arn
  task_role_arn      = aws_iam_role.task.arn

  container_definition {
    name      = "app"
    image     = "nginx:latest"
    cpu       = 256
    memory    = 512
    essential = true
  }
}

data "aws_ecs_daemon_task_definition" "test" {
  daemon_task_definition = aws_ecs_daemon_task_definition.test.arn
}
`, rName)
}
