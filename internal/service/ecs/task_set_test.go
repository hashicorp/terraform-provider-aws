// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfecs "github.com/hashicorp/terraform-provider-aws/internal/service/ecs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccECSTaskSet_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_task_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTaskSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTaskSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskSetExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "ecs", regexache.MustCompile(fmt.Sprintf("task-set/%[1]s/%[1]s/ecs-svc/.+", rName))),
					resource.TestCheckResourceAttr(resourceName, "service_registries.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "load_balancer.#", acctest.Ct0),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_task_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTaskSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTaskSetConfig_externalID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "service_registries.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "load_balancer.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrExternalID, "TEST_ID"),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_task_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTaskSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTaskSetConfig_scale(rName, 0.0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "scale.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scale.0.unit", string(awstypes.ScaleUnitPercent)),
					resource.TestCheckResourceAttr(resourceName, "scale.0.value", acctest.Ct0),
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
					testAccCheckTaskSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "scale.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "scale.0.unit", string(awstypes.ScaleUnitPercent)),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_task_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTaskSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTaskSetConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskSetExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfecs.ResourceTaskSet(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccECSTaskSet_withCapacityProviderStrategy(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_task_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTaskSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTaskSetConfig_capacityProviderStrategy(rName, 1, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskSetExists(ctx, resourceName),
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
					testAccCheckTaskSetExists(ctx, resourceName),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_task_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTaskSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTaskSetConfig_alb(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "load_balancer.#", acctest.Ct1),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_task_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTaskSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTaskSetConfig_launchTypeFargate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "launch_type", "FARGATE"),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.assign_public_ip", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.security_groups.#", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.subnets.#", acctest.Ct2),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_task_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTaskSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTaskSetConfig_launchTypeFargateAndPlatformVersion(rName, "1.3.0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskSetExists(ctx, resourceName),
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
					testAccCheckTaskSetExists(ctx, resourceName),
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
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_task_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTaskSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTaskSetConfig_serviceRegistries(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "service_registries.#", acctest.Ct1),
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

func TestAccECSTaskSet_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_task_set.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTaskSetDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTaskSetConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
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
				Config: testAccTaskSetConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccTaskSetConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTaskSetExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckTaskSetExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ECSClient(ctx)

		_, err := tfecs.FindTaskSetNoTagsByThreePartKey(ctx, conn, rs.Primary.Attributes["task_set_id"], rs.Primary.Attributes["service"], rs.Primary.Attributes["cluster"])

		return err
	}
}

func testAccCheckTaskSetDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ECSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ecs_task_set" {
				continue
			}

			_, err := tfecs.FindTaskSetNoTagsByThreePartKey(ctx, conn, rs.Primary.Attributes["task_set_id"], rs.Primary.Attributes["service"], rs.Primary.Attributes["cluster"])

			if tfresource.NotFound(err) {
				return nil
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ECS Task Set %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccTaskSetConfig_base(rName string) string {
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
	return acctest.ConfigCompose(testAccTaskSetConfig_base(rName), `
resource "aws_ecs_task_set" "test" {
  service         = aws_ecs_service.test.id
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
}
`)
}

func testAccTaskSetConfig_externalID(rName string) string {
	return acctest.ConfigCompose(testAccTaskSetConfig_base(rName), `
resource "aws_ecs_task_set" "test" {
  service         = aws_ecs_service.test.id
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  external_id     = "TEST_ID"
}
`)
}

func testAccTaskSetConfig_scale(rName string, scale float64) string {
	return acctest.ConfigCompose(testAccTaskSetConfig_base(rName), fmt.Sprintf(`
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
	return acctest.ConfigCompose(testAccCapacityProviderConfig_base(rName), testAccTaskSetConfig_base(rName), fmt.Sprintf(`
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

func testAccTaskSetConfig_alb(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
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
  subnets  = aws_subnet.test[*].id
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
	return acctest.ConfigCompose(testAccTaskSetConfig_base(rName), fmt.Sprintf(`
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
	return acctest.ConfigCompose(testAccTaskSetConfig_base(rName), fmt.Sprintf(`
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
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  ingress {
    protocol    = "-1"
    from_port   = 0
    to_port     = 0
    cidr_blocks = [aws_vpc.test.cidr_block]
  }

  tags = {
    Name = %[1]q
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
    subnets         = aws_subnet.test[*].id
  }
}
`, rName))
}

func testAccTaskSetConfig_launchTypeFargate(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
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

  tags = {
    Name = %[1]q
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
    security_groups  = aws_security_group.test[*].id
    subnets          = aws_subnet.test[*].id
    assign_public_ip = false
  }
}
`, rName))
}

func testAccTaskSetConfig_launchTypeFargateAndPlatformVersion(rName, platformVersion string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
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

  tags = {
    Name = %[1]q
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
    security_groups  = aws_security_group.test[*].id
    subnets          = aws_subnet.test[*].id
    assign_public_ip = false
  }
}
`, rName, platformVersion))
}
