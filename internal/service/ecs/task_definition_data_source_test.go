package ecs_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/ecs"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccECSTaskDefinitionDataSource_ecsTaskDefinition(t *testing.T) {
	resourceName := "data.aws_ecs_task_definition.mongo"
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(5))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskDefinitionDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", "aws_ecs_task_definition.mongo", "arn"),
					resource.TestCheckResourceAttr(resourceName, "family", rName),
					resource.TestCheckResourceAttr(resourceName, "network_mode", "bridge"),
					resource.TestMatchResourceAttr(resourceName, "revision", regexp.MustCompile("^[1-9][0-9]*$")),
					resource.TestCheckResourceAttr(resourceName, "status", "ACTIVE"),
					resource.TestCheckResourceAttrPair(resourceName, "task_role_arn", "aws_iam_role.mongo_role", "arn"),
				),
			},
		},
	})
}

func testAccTaskDefinitionDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "mongo_role" {
  name = "%[1]s"

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}

resource "aws_ecs_task_definition" "mongo" {
  family        = "%[1]s"
  task_role_arn = aws_iam_role.mongo_role.arn
  network_mode  = "bridge"

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

data "aws_ecs_task_definition" "mongo" {
  task_definition = aws_ecs_task_definition.mongo.family
}
`, rName)
}
