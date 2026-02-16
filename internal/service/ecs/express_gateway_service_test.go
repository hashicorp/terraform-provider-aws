// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ecs_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfecs "github.com/hashicorp/terraform-provider-aws/internal/service/ecs"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccECSExpressGatewayService_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var service awstypes.ECSExpressGatewayService
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_express_gateway_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ECSEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExpressGatewayServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccExpressGatewayServiceConfig_basic(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckExpressGatewayServiceExists(ctx, t, resourceName, &service),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("cpu"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("current_deployment"), knownvalue.Null()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("health_check_path"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("ingress_paths"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("memory"), knownvalue.NotNull()),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrNetworkConfiguration), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("scaling_target"), knownvalue.ListSizeExact(1)),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("service_arn"), tfknownvalue.RegionalARNRegexp("ecs", regexache.MustCompile(`service/.+`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New("service_revision_arn"), tfknownvalue.RegionalARNRegexp("ecs", regexache.MustCompile(`service-revision/.+`))),
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.Null()),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportStateVerifyIdentifierAttribute: "service_arn",
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "service_arn"),
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIgnore: []string{
					"wait_for_steady_state",
					"current_deployment", // not returned immediately in Create response, so will show as diff in Import
				},
			},
		},
	})
}

func TestAccECSExpressGatewayService_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var service awstypes.ECSExpressGatewayService
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_express_gateway_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExpressGatewayServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccExpressGatewayServiceConfig_basic(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExpressGatewayServiceExists(ctx, t, resourceName, &service),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfecs.ResourceExpressGatewayService, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccECSExpressGatewayService_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var service awstypes.ECSExpressGatewayService
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_express_gateway_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ECSEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExpressGatewayServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccExpressGatewayServiceConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExpressGatewayServiceExists(ctx, t, resourceName, &service),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1),
					})),
				},
			},
			{
				ResourceName:                         resourceName,
				ImportStateVerifyIdentifierAttribute: "service_arn",
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, "service_arn"),
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateVerifyIgnore: []string{
					"wait_for_steady_state",
					"current_deployment",
				},
			},
			{
				Config: testAccExpressGatewayServiceConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExpressGatewayServiceExists(ctx, t, resourceName, &service),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey1: knownvalue.StringExact(acctest.CtValue1Updated),
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
			{
				Config: testAccExpressGatewayServiceConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExpressGatewayServiceExists(ctx, t, resourceName, &service),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(resourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{
						acctest.CtKey2: knownvalue.StringExact(acctest.CtValue2),
					})),
				},
			},
		},
	})
}

func TestAccECSExpressGatewayService_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var service1, service2 awstypes.ECSExpressGatewayService
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_express_gateway_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ECSEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExpressGatewayServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccExpressGatewayServiceConfig_basic(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExpressGatewayServiceExists(ctx, t, resourceName, &service1),
					resource.TestCheckResourceAttr(resourceName, "primary_container.0.image", "public.ecr.aws/nginx/nginx:1.28-alpine3.21-slim"),
				),
			},
			{
				Config: testAccExpressGatewayServiceConfig_updated(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExpressGatewayServiceExists(ctx, t, resourceName, &service2),
					resource.TestCheckResourceAttr(resourceName, "primary_container.0.image", "public.ecr.aws/nginx/nginx:latest"),
				),
			},
		},
	})
}

func TestAccECSExpressGatewayService_waitForSteadyState(t *testing.T) {
	acctest.Skip(t, "Times out when running with full suite")
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var service awstypes.ECSExpressGatewayService
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_express_gateway_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ECSEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExpressGatewayServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccExpressGatewayServiceConfig_basic(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExpressGatewayServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "wait_for_steady_state", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "primary_container.0.image", "public.ecr.aws/nginx/nginx:1.28-alpine3.21-slim"),
				),
			},
			{
				Config: testAccExpressGatewayServiceConfig_updated(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExpressGatewayServiceExists(ctx, t, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "primary_container.0.image", "public.ecr.aws/nginx/nginx:latest"),
				),
			},
		},
	})
}

func TestAccECSExpressGatewayService_networkConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var service1 awstypes.ECSExpressGatewayService
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_express_gateway_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ECSEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExpressGatewayServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccExpressGatewayServiceConfig_networkConfiguration(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckExpressGatewayServiceExists(ctx, t, resourceName, &service1),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.subnets.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "network_configuration.0.security_groups.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "network_configuration.0.subnets.0"),
					resource.TestCheckResourceAttrSet(resourceName, "network_configuration.0.subnets.1"),
					resource.TestCheckResourceAttrSet(resourceName, "network_configuration.0.security_groups.0"),
					// Verify additional attributes are set correctly
					resource.TestCheckResourceAttr(resourceName, names.AttrServiceName, rName+"-service"),
					resource.TestCheckResourceAttr(resourceName, "cpu", "512"),
					resource.TestCheckResourceAttr(resourceName, "memory", "1024"),
					resource.TestCheckResourceAttr(resourceName, "health_check_path", "/"),
					resource.TestCheckResourceAttr(resourceName, "primary_container.0.container_port", "80"),
					resource.TestCheckResourceAttr(resourceName, "primary_container.0.environment.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "primary_container.0.environment.0.name", "ENV"),
					resource.TestCheckResourceAttr(resourceName, "primary_container.0.environment.0.value", "test"),
					resource.TestCheckResourceAttr(resourceName, "primary_container.0.aws_logs_configuration.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "primary_container.0.aws_logs_configuration.0.log_group"),
					resource.TestCheckResourceAttr(resourceName, "primary_container.0.aws_logs_configuration.0.log_stream_prefix", "test"),
				),
			},
		},
	})
}

func TestAccECSExpressGatewayService_checkIdempotency(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_ecs_express_gateway_service.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ECSEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExpressGatewayServiceDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccExpressGatewayServiceConfig_basic(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExpressGatewayServiceExists(ctx, t, resourceName, &awstypes.ECSExpressGatewayService{}),
				),
			},
			{
				Config:      testAccExpressGatewayServiceConfig_duplicate(rName),
				ExpectError: regexache.MustCompile("Express Gateway Service .* already exists in cluster"),
			},
		},
	})
}

func testAccCheckExpressGatewayServiceDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ECSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ecs_express_gateway_service" {
				continue
			}

			output, err := tfecs.FindExpressGatewayServiceByARN(ctx, conn, rs.Primary.Attributes["service_arn"])

			if retry.NotFound(err) {
				return nil
			}

			if err != nil {
				return err
			}

			if string(output.Status.StatusCode) == string(awstypes.ExpressGatewayServiceStatusCodeInactive) ||
				string(output.Status.StatusCode) == string(awstypes.ExpressGatewayServiceStatusCodeDraining) {
				return nil
			}

			return fmt.Errorf("ECS Express Gateway Service %s still exists", rs.Primary.Attributes["service_arn"])
		}

		return nil
	}
}

func testAccCheckExpressGatewayServiceExists(ctx context.Context, t *testing.T, n string, v *awstypes.ECSExpressGatewayService) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).ECSClient(ctx)

		output, err := tfecs.FindExpressGatewayServiceByARN(ctx, conn, rs.Primary.Attributes["service_arn"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccExpressGatewayServiceConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_vpc" "default" {
  default = true
}

data "aws_subnets" "default" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.default.id]
  }
}

data "aws_security_group" "default" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.default.id]
  }
  filter {
    name   = "group-name"
    values = ["default"]
  }
}

resource "aws_iam_role" "execution" {
  name               = "%[1]s-execution"
  assume_role_policy = <<POLICY
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ecs-tasks.amazonaws.com"
        ]
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "execution" {
  role       = aws_iam_role.execution.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

resource "aws_iam_role_policy" "execution_logs" {
  name = "CreateLogGroupPolicy"
  role = aws_iam_role.execution.id
  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect   = "Allow"
      Action   = "logs:CreateLogGroup"
      Resource = "*"
    }]
  })
}

resource "aws_iam_role" "infrastructure" {
  name               = "%[1]s-infra"
  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ecs.amazonaws.com"
        ]
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy_attachment" "infrastructure" {
  role       = aws_iam_role.infrastructure.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonECSInfrastructureRoleforExpressGatewayServices"
}
`, rName)
}

func testAccExpressGatewayServiceConfig_basic(rName string, waitForSteadyState bool) string {
	waitForSteadyStateConfig := ""
	if waitForSteadyState {
		waitForSteadyStateConfig = "wait_for_steady_state = true"
	}

	return acctest.ConfigCompose(testAccExpressGatewayServiceConfig_base(rName), fmt.Sprintf(`
resource "aws_ecs_express_gateway_service" "test" {
  execution_role_arn      = aws_iam_role.execution.arn
  infrastructure_role_arn = aws_iam_role.infrastructure.arn
  %[2]s

  primary_container {
    image = "public.ecr.aws/nginx/nginx:1.28-alpine3.21-slim"
  }
}
`, rName, waitForSteadyStateConfig))
}

func testAccExpressGatewayServiceConfig_updated(rName string, waitForSteadyState bool) string {
	waitForSteadyStateConfig := ""
	if waitForSteadyState {
		waitForSteadyStateConfig = "wait_for_steady_state = true"
	}

	return acctest.ConfigCompose(testAccExpressGatewayServiceConfig_base(rName), fmt.Sprintf(`
resource "aws_ecs_express_gateway_service" "test" {
  execution_role_arn      = aws_iam_role.execution.arn
  infrastructure_role_arn = aws_iam_role.infrastructure.arn
  %[2]s

  primary_container {
    image = "public.ecr.aws/nginx/nginx:latest"
  }

  scaling_target {
    min_task_count            = 0
    max_task_count            = 1
    auto_scaling_metric       = "AVERAGE_CPU"
    auto_scaling_target_value = 60
  }
}
`, rName, waitForSteadyStateConfig))
}

func testAccExpressGatewayServiceConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccExpressGatewayServiceConfig_base(rName), fmt.Sprintf(`
resource "aws_ecs_express_gateway_service" "test" {
  execution_role_arn      = aws_iam_role.execution.arn
  infrastructure_role_arn = aws_iam_role.infrastructure.arn

  primary_container {
    image = "public.ecr.aws/nginx/nginx:1.28-alpine3.21-slim"
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccExpressGatewayServiceConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccExpressGatewayServiceConfig_base(rName), fmt.Sprintf(`
resource "aws_ecs_express_gateway_service" "test" {
  execution_role_arn      = aws_iam_role.execution.arn
  infrastructure_role_arn = aws_iam_role.infrastructure.arn

  primary_container {
    image = "public.ecr.aws/nginx/nginx:1.28-alpine3.21-slim"
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccExpressGatewayServiceConfig_networkConfiguration(rName string) string {
	return acctest.ConfigCompose(testAccExpressGatewayServiceConfig_base(rName), fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"
}

resource "aws_vpc" "test" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_hostnames = true
  enable_dns_support   = true
  tags = {
    Name = "%[1]s-vpc"
  }
}

resource "aws_subnet" "test_subnet1" {
  vpc_id                  = aws_vpc.test.id
  cidr_block              = "10.0.1.0/24"
  availability_zone       = data.aws_availability_zones.available.names[0]
  map_public_ip_on_launch = false
  tags = {
    Name = "%[1]s-subnet-1"
  }
}

resource "aws_subnet" "test_subnet2" {
  vpc_id                  = aws_vpc.test.id
  cidr_block              = "10.0.2.0/24"
  availability_zone       = data.aws_availability_zones.available.names[1]
  map_public_ip_on_launch = false
  tags = {
    Name = "%[1]s-subnet-2"
  }
}

resource "aws_security_group" "test" {
  name        = "%[1]s-sg"
  description = "Security group for ECS Express Gateway Service test"
  vpc_id      = aws_vpc.test.id
  tags = {
    Name = "%[1]s-sg"
  }
}

resource "aws_cloudwatch_log_group" "test" {
  name              = "/ecs/%[1]s-log-group"
  retention_in_days = 30
}

resource "aws_iam_role" "task_role" {
  name = "%[1]s-task-role"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "ecs-tasks.amazonaws.com"
        }
        Action = "sts:AssumeRole"
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "task_role" {
  role       = aws_iam_role.task_role.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/CloudWatchFullAccess"
}

resource "aws_ecs_express_gateway_service" "test" {
  execution_role_arn      = aws_iam_role.execution.arn
  infrastructure_role_arn = aws_iam_role.infrastructure.arn
  service_name            = "%[1]s-service"
  cluster                 = "default"
  cpu                     = "512"
  memory                  = "1024"
  task_role_arn           = aws_iam_role.task_role.arn
  health_check_path       = "/"

  primary_container {
    image          = "public.ecr.aws/nginx/nginx:1.28-alpine3.21-slim"
    container_port = 80
    command        = ["CMD-SHELL"]

    environment {
      name  = "ENV"
      value = "test"
    }

    aws_logs_configuration {
      log_group         = aws_cloudwatch_log_group.test.name
      log_stream_prefix = "test"
    }
  }

  network_configuration {
    subnets         = [aws_subnet.test_subnet1.id, aws_subnet.test_subnet2.id]
    security_groups = [aws_security_group.test.id]
  }
}
`, rName))
}

func testAccExpressGatewayServiceConfig_duplicate(rName string) string {
	return acctest.ConfigCompose(testAccExpressGatewayServiceConfig_base(rName), `
resource "aws_ecs_express_gateway_service" "test" {
  execution_role_arn      = aws_iam_role.execution.arn
  infrastructure_role_arn = aws_iam_role.infrastructure.arn

  primary_container {
    image = "public.ecr.aws/nginx/nginx:1.28-alpine3.21-slim"
  }
}

resource "aws_ecs_express_gateway_service" "duplicate" {
  execution_role_arn      = aws_iam_role.execution.arn
  infrastructure_role_arn = aws_iam_role.infrastructure.arn
  service_name            = aws_ecs_express_gateway_service.test.service_name

  primary_container {
    image = "public.ecr.aws/nginx/nginx:1.28-alpine3.21-slim"
  }

  depends_on = [
    aws_ecs_express_gateway_service.test
  ]
}
`)
}
