package ecs_test

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfecs "github.com/hashicorp/terraform-provider-aws/internal/service/ecs"
)

func TestAccECSTaskSet_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_task_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskSetExists(resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "ecs", regexp.MustCompile(fmt.Sprintf("task-set/%[1]s/%[1]s/ecs-svc/.+", rName))),
					resource.TestCheckResourceAttr(resourceName, "service_registries.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "load_balancer.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"stability_status",
					"wait_until_stable",
					"wait_until_stable_timeout",
				},
			},
		},
	})
}

func TestAccECSTaskSet_withExternalId(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_task_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskSetConfig_externalID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "service_registries.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "load_balancer.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "external_id", "TEST_ID"),
				),
			},
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateVerifyIgnore: []string{
					"wait_until_stable",
					"wait_until_stable_timeout",
				},
			},
		},
	})
}

func TestAccECSTaskSet_withScale(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_task_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskSetConfig_scale(rName, 0.0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "scale.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scale.0.unit", ecs.ScaleUnitPercent),
					resource.TestCheckResourceAttr(resourceName, "scale.0.value", "0"),
				),
			},
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateVerifyIgnore: []string{
					"wait_until_stable",
					"wait_until_stable_timeout",
				},
			},
			{
				Config: testAccTaskSetConfig_scale(rName, 100.0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "scale.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "scale.0.unit", ecs.ScaleUnitPercent),
					resource.TestCheckResourceAttr(resourceName, "scale.0.value", "100"),
				),
			},
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateVerifyIgnore: []string{
					"wait_until_stable",
					"wait_until_stable_timeout",
				},
			},
		},
	})
}

func TestAccECSTaskSet_disappears(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_task_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskSetExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfecs.ResourceTaskSet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccECSTaskSet_withCapacityProviderStrategy(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_task_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskSetConfig_capacityProviderStrategy(rName, 1, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskSetExists(resourceName),
				),
			},
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateVerifyIgnore: []string{
					"wait_until_stable",
					"wait_until_stable_timeout",
				},
			},
			{
				Config: testAccTaskSetConfig_capacityProviderStrategy(rName, 10, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskSetExists(resourceName),
				),
			},
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateVerifyIgnore: []string{
					"wait_until_stable",
					"wait_until_stable_timeout",
				},
			},
		},
	})
}

func TestAccECSTaskSet_withMultipleCapacityProviderStrategies(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_task_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskSetConfig_multipleCapacityProviderStrategies(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "capacity_provider_strategy.#", "2"),
				),
			},
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateVerifyIgnore: []string{
					"wait_until_stable",
					"wait_until_stable_timeout",
				},
			},
		},
	})
}

func TestAccECSTaskSet_withAlb(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_task_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskSetConfig_alb(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "load_balancer.#", "1"),
				),
			},
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateVerifyIgnore: []string{
					"wait_until_stable",
					"wait_until_stable_timeout",
				},
			},
		},
	})
}

func TestAccECSTaskSet_withLaunchTypeFargate(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_task_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskSetConfig_launchTypeFargate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "launch_type", "FARGATE"),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.assign_public_ip", "false"),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.security_groups.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.subnets.#", "2"),
					resource.TestCheckResourceAttrSet(resourceName, "platform_version"),
				),
			},
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateVerifyIgnore: []string{
					"wait_until_stable",
					"wait_until_stable_timeout",
				},
			},
		},
	})
}

func TestAccECSTaskSet_withLaunchTypeFargateAndPlatformVersion(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_task_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskSetConfig_launchTypeFargateAndPlatformVersion(rName, "1.3.0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "platform_version", "1.3.0"),
				),
			},
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateVerifyIgnore: []string{
					"wait_until_stable",
					"wait_until_stable_timeout",
				},
			},
			{
				Config: testAccTaskSetConfig_launchTypeFargateAndPlatformVersion(rName, "1.4.0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "platform_version", "1.4.0"),
				),
			},
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateVerifyIgnore: []string{
					"wait_until_stable",
					"wait_until_stable_timeout",
				},
			},
		},
	})
}

func TestAccECSTaskSet_withServiceRegistries(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_task_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskSetConfig_serviceRegistries(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "service_registries.#", "1"),
				),
			},
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateVerifyIgnore: []string{
					"wait_until_stable",
					"wait_until_stable_timeout",
				},
			},
		},
	})
}

func TestAccECSTaskSet_Tags(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_task_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ecs.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTaskSetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTaskSetConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateVerifyIgnore: []string{
					"wait_until_stable",
					"wait_until_stable_timeout",
				},
			},
			{
				Config: testAccTaskSetConfig_tags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccTaskSetConfig_tags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskSetExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

//////////////
// Fixtures //
//////////////

func testAccTaskSetBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family                = %[1]q
  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "mongo:latest",
    "memory": 128,
    "name": "mongodb"
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  name          = %[1]q
  cluster       = aws_ecs_cluster.test.id
  desired_count = 1
  deployment_controller {
    type = "EXTERNAL"
  }
}
`, rName)
}

func testAccTaskSetConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccTaskSetBaseConfig(rName),
		`
resource "aws_ecs_task_set" "test" {
  service         = aws_ecs_service.test.id
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
}
`)
}

func testAccTaskSetConfig_externalID(rName string) string {
	return acctest.ConfigCompose(
		testAccTaskSetBaseConfig(rName),
		`
resource "aws_ecs_task_set" "test" {
  service         = aws_ecs_service.test.id
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  external_id     = "TEST_ID"
}
`)
}

func testAccTaskSetConfig_scale(rName string, scale float64) string {
	return acctest.ConfigCompose(
		testAccTaskSetBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_ecs_task_set" "test" {
  service         = aws_ecs_service.test.id
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  scale {
    value = %[1]f
  }
}
`, scale))
}

func testAccTaskSetConfig_capacityProviderStrategy(rName string, weight, base int) string {
	return acctest.ConfigCompose(
		testAccCapacityProviderBaseConfig(rName),
		testAccTaskSetBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_ecs_capacity_provider" "test" {
  name = %[1]q
  auto_scaling_group_provider {
    auto_scaling_group_arn = aws_autoscaling_group.test.arn
  }
}

resource "aws_ecs_task_set" "test" {
  service         = aws_ecs_service.test.id
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  capacity_provider_strategy {
    capacity_provider = aws_ecs_capacity_provider.test.name
    weight            = %[2]d
    base              = %[3]d
  }
}
`, rName, weight, base))
}

func testAccTaskSetConfig_multipleCapacityProviderStrategies(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.10.0.0/16"
  tags = {
    Name = "tf-acc-ecs-service-with-multiple-capacity-providers"
  }
}

resource "aws_security_group" "test" {
  name        = %[1]q
  description = "Allow all inbound traffic"
  vpc_id      = aws_vpc.test.id
  ingress {
    protocol    = "tcp"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = [aws_vpc.test.cidr_block]
  }
}

resource "aws_subnet" "test" {
  cidr_block = cidrsubnet(aws_vpc.test.cidr_block, 8, 1)
  vpc_id     = aws_vpc.test.id
  tags = {
    Name = "tf-acc-ecs-service-with-multiple-capacity-providers"
  }
}

resource "aws_ecs_cluster" "test" {
  name = %[1]q

  capacity_providers = ["FARGATE_SPOT", "FARGATE"]

  default_capacity_provider_strategy {
    capacity_provider = "FARGATE_SPOT"
    weight            = 1
    base              = 1
  }

  default_capacity_provider_strategy {
    capacity_provider = "FARGATE"
    weight            = 1
  }
}

resource "aws_ecs_service" "test" {
  name          = %[1]q
  cluster       = aws_ecs_cluster.test.id
  desired_count = 1
  deployment_controller {
    type = "EXTERNAL"
  }
}

resource "aws_ecs_task_definition" "test" {
  family                   = %[1]q
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "256"
  memory                   = "512"
  container_definitions    = <<DEFINITION
[
  {
    "cpu": 256,
    "essential": true,
    "image": "mongo:latest",
    "memory": 512,
    "name": "mongodb",
    "networkMode": "awsvpc"
  }
]
DEFINITION
}

resource "aws_ecs_task_set" "test" {
  service         = aws_ecs_service.test.id
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  network_configuration {
    security_groups  = [aws_security_group.test.id]
    subnets          = [aws_subnet.test.id]
    assign_public_ip = false
  }
  capacity_provider_strategy {
    capacity_provider = "FARGATE"
    weight            = 1
  }
  capacity_provider_strategy {
    capacity_provider = "FARGATE_SPOT"
    weight            = 1
  }
}
`, rName)
}

func testAccTaskSetConfig_alb(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.10.0.0/16"
  tags = {
    Name = "terraform-testacc-ecs-service-with-alb"
  }
}

resource "aws_subnet" "test" {
  count             = 2
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  availability_zone = data.aws_availability_zones.available.names[count.index]
  vpc_id            = aws_vpc.test.id
  tags = {
    Name = "tf-acc-ecs-service-with-alb"
  }
}

resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family                = %[1]q
  container_definitions = <<DEFINITION
[
  {
    "cpu": 256,
    "essential": true,
    "image": "ghost:latest",
    "memory": 512,
    "name": "ghost",
    "portMappings": [
      {
        "containerPort": 2368,
        "hostPort": 8080
      }
    ]
  }
]
DEFINITION
}

resource "aws_lb_target_group" "test" {
  name     = aws_lb.test.name
  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id
}

resource "aws_lb" "test" {
  name     = %[1]q
  internal = true
  subnets  = aws_subnet.test.*.id
}

resource "aws_lb_listener" "test" {
  load_balancer_arn = aws_lb.test.id
  port              = "80"
  protocol          = "HTTP"
  default_action {
    target_group_arn = aws_lb_target_group.test.id
    type             = "forward"
  }
}

resource "aws_ecs_service" "test" {
  name          = %[1]q
  cluster       = aws_ecs_cluster.test.id
  desired_count = 1
  deployment_controller {
    type = "EXTERNAL"
  }
}

resource "aws_ecs_task_set" "test" {
  service         = aws_ecs_service.test.id
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  load_balancer {
    target_group_arn = aws_lb_target_group.test.id
    container_name   = "ghost"
    container_port   = "2368"
  }
}
`, rName))
}

func testAccTaskSetConfig_tags1(rName, tag1Key, tag1Value string) string {
	return acctest.ConfigCompose(
		testAccTaskSetBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_ecs_task_set" "test" {
  service         = aws_ecs_service.test.id
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  tags = {
    %[1]q = %[2]q
  }
}
`, tag1Key, tag1Value))
}

func testAccTaskSetConfig_tags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return acctest.ConfigCompose(
		testAccTaskSetBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_ecs_task_set" "test" {
  service         = aws_ecs_service.test.id
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tag1Key, tag1Value, tag2Key, tag2Value))
}

func testAccTaskSetConfig_serviceRegistries(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = "tf-acc-with-svc-reg"
  }
}

resource "aws_subnet" "test" {
  count             = 2
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  availability_zone = data.aws_availability_zones.available.names[count.index]
  vpc_id            = aws_vpc.test.id
  tags = {
    Name = "tf-acc-with-svc-reg"
  }
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id
  ingress {
    protocol    = "-1"
    from_port   = 0
    to_port     = 0
    cidr_blocks = [aws_vpc.test.cidr_block]
  }
}

resource "aws_service_discovery_private_dns_namespace" "test" {
  name        = "%[1]s.terraform.local"
  description = "test"
  vpc         = aws_vpc.test.id
}

resource "aws_service_discovery_service" "test" {
  name = %[1]q
  dns_config {
    namespace_id = aws_service_discovery_private_dns_namespace.test.id
    dns_records {
      ttl  = 5
      type = "SRV"
    }
  }
}

resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family                = %[1]q
  network_mode          = "awsvpc"
  container_definitions = <<DEFINITION
[
  {
    "cpu": 128,
    "essential": true,
    "image": "mongo:latest",
    "memory": 128,
    "name": "mongodb"
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  name          = %[1]q
  cluster       = aws_ecs_cluster.test.id
  desired_count = 1
  deployment_controller {
    type = "EXTERNAL"
  }
}

resource "aws_ecs_task_set" "test" {
  service         = aws_ecs_service.test.id
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  service_registries {
    port         = 34567
    registry_arn = aws_service_discovery_service.test.arn
  }
  network_configuration {
    security_groups = [aws_security_group.test.id]
    subnets         = aws_subnet.test.*.id
  }
}
`, rName))
}

func testAccTaskSetConfig_launchTypeFargate(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.10.0.0/16"
  tags = {
    Name = "terraform-testacc-ecs-service-with-launch-type-fargate"
  }
}

resource "aws_subnet" "test" {
  count             = 2
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  availability_zone = data.aws_availability_zones.available.names[count.index]
  vpc_id            = aws_vpc.test.id
  tags = {
    Name = "tf-acc-ecs-service-with-launch-type-fargate"
  }
}

resource "aws_security_group" "test" {
  count = 2

  name        = "%[1]s-${count.index}"
  description = "Allow all inbound traffic"
  vpc_id      = aws_vpc.test.id
  ingress {
    protocol    = "6"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = [aws_vpc.test.cidr_block]
  }
}

resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family                   = %[1]q
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "256"
  memory                   = "512"
  container_definitions    = <<DEFINITION
[
  {
    "cpu": 256,
    "essential": true,
    "image": "mongo:latest",
    "memory": 512,
    "name": "mongodb",
    "networkMode": "awsvpc"
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  name          = %[1]q
  cluster       = aws_ecs_cluster.test.id
  desired_count = 1
  deployment_controller {
    type = "EXTERNAL"
  }
}

resource "aws_ecs_task_set" "test" {
  service         = aws_ecs_service.test.id
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  launch_type     = "FARGATE"
  network_configuration {
    security_groups  = aws_security_group.test.*.id
    subnets          = aws_subnet.test.*.id
    assign_public_ip = false
  }
}
`, rName))
}

func testAccTaskSetConfig_launchTypeFargateAndPlatformVersion(rName, platformVersion string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.10.0.0/16"
  tags = {
    Name = "terraform-testacc-ecs-service-with-launch-type-fargate-and-platform-version"
  }
}

resource "aws_subnet" "test" {
  count             = 2
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  availability_zone = data.aws_availability_zones.available.names[count.index]
  vpc_id            = aws_vpc.test.id
  tags = {
    Name = "tf-acc-ecs-service-with-launch-type-fargate-and-platform-version"
  }
}

resource "aws_security_group" "test" {
  count = 2

  name        = "%[1]s-${count.index}"
  description = "Allow all inbound traffic"
  vpc_id      = aws_vpc.test.id
  ingress {
    protocol    = "6"
    from_port   = 80
    to_port     = 8000
    cidr_blocks = [aws_vpc.test.cidr_block]
  }
}

resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family                   = %[1]q
  network_mode             = "awsvpc"
  requires_compatibilities = ["FARGATE"]
  cpu                      = "256"
  memory                   = "512"
  container_definitions    = <<DEFINITION
[
  {
    "cpu": 256,
    "essential": true,
    "image": "mongo:latest",
    "memory": 512,
    "name": "mongodb",
    "networkMode": "awsvpc"
  }
]
DEFINITION
}

resource "aws_ecs_service" "test" {
  name          = %[1]q
  cluster       = aws_ecs_cluster.test.id
  desired_count = 1
  deployment_controller {
    type = "EXTERNAL"
  }
}

resource "aws_ecs_task_set" "test" {
  service          = aws_ecs_service.test.id
  cluster          = aws_ecs_cluster.test.id
  task_definition  = aws_ecs_task_definition.test.arn
  launch_type      = "FARGATE"
  platform_version = %[2]q
  network_configuration {
    security_groups  = aws_security_group.test.*.id
    subnets          = aws_subnet.test.*.id
    assign_public_ip = false
  }
}
`, rName, platformVersion))
}

func testAccCheckTaskSetExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ECSConn

		taskSetId, service, cluster, err := tfecs.TaskSetParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		input := &ecs.DescribeTaskSetsInput{
			TaskSets: aws.StringSlice([]string{taskSetId}),
			Cluster:  aws.String(cluster),
			Service:  aws.String(service),
		}

		output, err := conn.DescribeTaskSets(input)

		if err != nil {
			return err
		}

		if output == nil || len(output.TaskSets) == 0 {
			return fmt.Errorf("ECS TaskSet (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckTaskSetDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).ECSConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ecs_task_set" {
			continue
		}

		taskSetId, service, cluster, err := tfecs.TaskSetParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		input := &ecs.DescribeTaskSetsInput{
			TaskSets: aws.StringSlice([]string{taskSetId}),
			Cluster:  aws.String(cluster),
			Service:  aws.String(service),
		}

		output, err := conn.DescribeTaskSets(input)

		if tfawserr.ErrCodeEquals(err, ecs.ErrCodeClusterNotFoundException, ecs.ErrCodeServiceNotFoundException, ecs.ErrCodeTaskSetNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}

		if output != nil && len(output.TaskSets) == 1 {
			return fmt.Errorf("ECS TaskSet (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}
