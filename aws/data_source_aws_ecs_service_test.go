package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSEcsDataSource_ecsService(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsEcsServiceDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_ecs_service.default", "service_name", "mongodb"),
					resource.TestCheckResourceAttr("data.aws_ecs_service.default", "desired_count", "1"),
					resource.TestCheckResourceAttr("data.aws_ecs_service.default", "launch_type", "EC2"),
					resource.TestCheckResourceAttrSet("data.aws_ecs_service.default", "arn"),
					resource.TestCheckResourceAttrSet("data.aws_ecs_service.default", "task_definition"),
				),
			},
		},
	})
}

var testAccCheckAwsEcsServiceDataSourceConfig = fmt.Sprintf(`
resource "aws_ecs_cluster" "default" {
  name = "default-%d"
}

resource "aws_ecs_task_definition" "mongo" {
  family = "mongodb"
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

resource "aws_ecs_service" "mongo" {
  name = "mongodb"
  cluster = "${aws_ecs_cluster.default.id}"
  task_definition = "${aws_ecs_task_definition.mongo.arn}"
  desired_count = 1
}

data "aws_ecs_service" "default" {
  service_name = "${aws_ecs_service.mongo.name}"
  cluster_arn = "${aws_ecs_cluster.default.arn}"
}
`, acctest.RandInt())
