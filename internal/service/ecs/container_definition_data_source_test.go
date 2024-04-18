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

func TestAccECSContainerDefinitionDataSource_ecsContainerDefinition(t *testing.T) {
	ctx := acctest.Context(t)
	rString := sdkacctest.RandString(8)
	clusterName := fmt.Sprintf("tf_acc_td_ds_cluster_ecs_containter_definition_%s", rString)
	svcName := fmt.Sprintf("tf_acc_svc_td_ds_ecs_containter_definition_%s", rString)
	tdName := fmt.Sprintf("tf_acc_td_ds_ecs_containter_definition_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccContainerDefinitionDataSourceConfig_basic(clusterName, tdName, svcName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_ecs_container_definition.mongo", "image", "mongo:latest"),
					resource.TestCheckResourceAttr("data.aws_ecs_container_definition.mongo", "image_digest", "latest"),
					resource.TestCheckResourceAttr("data.aws_ecs_container_definition.mongo", "memory", "128"),
					resource.TestCheckResourceAttr("data.aws_ecs_container_definition.mongo", "memory_reservation", "64"),
					resource.TestCheckResourceAttr("data.aws_ecs_container_definition.mongo", "cpu", "128"),
					resource.TestCheckResourceAttr("data.aws_ecs_container_definition.mongo", "environment.SECRET", "KEY"),
				),
			},
		},
	})
}

func testAccContainerDefinitionDataSourceConfig_basic(clusterName, tdName, svcName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "default" {
  name = "%s"
}

resource "aws_ecs_task_definition" "mongo" {
  family = "%s"

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

resource "aws_ecs_service" "mongo" {
  name            = "%s"
  cluster         = aws_ecs_cluster.default.id
  task_definition = aws_ecs_task_definition.mongo.arn
  desired_count   = 1
}

data "aws_ecs_container_definition" "mongo" {
  task_definition = aws_ecs_task_definition.mongo.id
  container_name  = "mongodb"
}
`, clusterName, tdName, svcName)
}
