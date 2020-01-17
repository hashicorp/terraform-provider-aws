package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccAWSEcsDataSource_ecsTaskDefinition(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_ecs_task_definition.test"
	resourceName := "aws_ecs_task_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsEcsTaskDefinitionDataSourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "family", resourceName, "family"),
					resource.TestCheckResourceAttrPair(dataSourceName, "network_mode", resourceName, "network_mode"),
					resource.TestCheckResourceAttrPair(dataSourceName, "revision", resourceName, "revision"),
					resource.TestCheckResourceAttr(dataSourceName, "status", "ACTIVE"),
					resource.TestCheckResourceAttrPair(dataSourceName, "task_role_arn", resourceName, "task_role_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "execution_role_arn", resourceName, "execution_role_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cpu", resourceName, "cpu"),
					resource.TestCheckResourceAttrPair(dataSourceName, "memory", resourceName, "memory"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags", resourceName, "tags"),
					resource.TestCheckResourceAttrPair(dataSourceName, "volume", resourceName, "volume"),
					resource.TestCheckResourceAttrPair(dataSourceName, "placement_constraints", resourceName, "placement_constraints"),
					resource.TestCheckResourceAttrPair(dataSourceName, "requires_compatibilities", resourceName, "requires_compatibilities"),
					resource.TestCheckResourceAttrPair(dataSourceName, "proxy_configuration", resourceName, "proxy_configuration"),
					resource.TestCheckResourceAttrPair(resourceName, "task_role_arn", "aws_iam_role.test", "arn"),
				),
			},
		},
	})
}

func testAccCheckAwsEcsTaskDefinitionDataSourceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

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

resource "aws_ecs_task_definition" "test" {
  family        = %[1]q
  task_role_arn = "${aws_iam_role.test.arn}"
  network_mode  = "bridge"

  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "environment": [{
      "name": "SECRET",
      "value": "KEY"
    }],
    "essential": true,
    "image": "mongo:latest",
    "memory": 128,
    "memoryReservation": 64,
    "name": "testdb"
  }
]
DEFINITION
}

data "aws_ecs_task_definition" "test" {
  task_definition = "${aws_ecs_task_definition.test.family}"
}
`, rName)
}
