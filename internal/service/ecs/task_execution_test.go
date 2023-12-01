package ecs_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ecs"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccECSTaskExecutionResource_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_task_execution.test"
	clusterName := "aws_ecs_cluster.test"
	taskDefinitionName := "aws_ecs_task_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(t, ecs.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskExecutionResourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "cluster", clusterName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "task_definition", taskDefinitionName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "desired_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_type", "FARGATE"),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "task_arns.#", "1"),
				),
			},
		},
	})
}

func TestAccECSTaskExecutionResource_tags(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_task_execution.test"
	clusterName := "aws_ecs_cluster.test"
	taskDefinitionName := "aws_ecs_task_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(t, ecs.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, ecs.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskExecutionResourceConfig_tags(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "cluster", clusterName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "task_definition", taskDefinitionName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "desired_count", "1"),
					resource.TestCheckResourceAttr(resourceName, "launch_type", "FARGATE"),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "task_arns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
		},
	})
}

func testAccTaskExecutionResourceConfig_base(rName string) string {
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

func testAccTaskExecutionResourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
		testAccTaskExecutionResourceConfig_base(rName),
		`
resource "aws_ecs_task_execution" "test" {
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
}
`)
}

func testAccTaskExecutionResourceConfig_tags(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
		testAccTaskExecutionResourceConfig_base(rName),
		fmt.Sprintf(`
resource "aws_ecs_task_execution" "test" {
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
