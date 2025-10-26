// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccECSUpdateServiceAction_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ECS)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		CheckDestroy: acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccUpdateServiceActionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUpdateServiceAction(ctx, rName),
				),
			},
		},
	})
}

func TestAccECSUpdateServiceAction_desiredCount(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ECS)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		CheckDestroy: acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccUpdateServiceActionConfig_desiredCount(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckUpdateServiceActionDesiredCount(ctx, rName, 2),
				),
			},
		},
	})
}

// Test helper functions

func testAccCheckUpdateServiceAction(ctx context.Context, serviceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ECSClient(ctx)

		// Verify the service exists and is stable
		input := &ecs.DescribeServicesInput{
			Cluster:  &serviceName, // Using same name for cluster and service
			Services: []string{serviceName},
		}

		output, err := conn.DescribeServices(ctx, input)
		if err != nil {
			return fmt.Errorf("failed to describe ECS service %s: %w", serviceName, err)
		}

		if len(output.Services) == 0 {
			return fmt.Errorf("ECS service %s not found", serviceName)
		}

		service := output.Services[0]
		if len(service.Deployments) == 0 {
			return fmt.Errorf("ECS service %s has no deployments", serviceName)
		}

		// Check that service has a stable deployment
		primaryDeployment := service.Deployments[0]
		if primaryDeployment.RolloutState != awstypes.DeploymentRolloutStateCompleted {
			return fmt.Errorf("expected service %s to have completed deployment, got: %s", serviceName, primaryDeployment.RolloutState)
		}

		return nil
	}
}

func testAccCheckUpdateServiceActionDesiredCount(ctx context.Context, serviceName string, expectedCount int32) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ECSClient(ctx)

		input := &ecs.DescribeServicesInput{
			Cluster:  &serviceName,
			Services: []string{serviceName},
		}

		output, err := conn.DescribeServices(ctx, input)
		if err != nil {
			return fmt.Errorf("failed to describe ECS service %s: %w", serviceName, err)
		}

		if len(output.Services) == 0 {
			return fmt.Errorf("ECS service %s not found", serviceName)
		}

		service := output.Services[0]
		if service.DesiredCount != expectedCount {
			return fmt.Errorf("expected service %s to have desired count %d, got %d", serviceName, expectedCount, service.DesiredCount)
		}

		if service.RunningCount != expectedCount {
			return fmt.Errorf("expected service %s to have running count %d, got %d", serviceName, expectedCount, service.RunningCount)
		}

		return nil
	}
}

// Configuration functions

func testAccUpdateServiceActionConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccUpdateServiceActionConfig_base(rName),
		`
action "aws_ecs_update_service" "test" {
  config {
    cluster_name        = aws_ecs_cluster.test.name
    service_name        = aws_ecs_service.test.name
    force_new_deployment = true
  }
}

resource "terraform_data" "trigger" {
  input = "trigger"
  lifecycle {
    action_trigger {
      events  = [before_create, before_update]
      actions = [action.aws_ecs_update_service.test]
    }
  }
}
`)
}

func testAccUpdateServiceActionConfig_desiredCount(rName string) string {
	return acctest.ConfigCompose(
		testAccUpdateServiceActionConfig_base(rName),
		`
action "aws_ecs_update_service" "test" {
  config {
    cluster_name = aws_ecs_cluster.test.name
    service_name = aws_ecs_service.test.name
    desired_count = 2
  }
}

resource "terraform_data" "trigger" {
  input = "trigger"
  lifecycle {
    action_trigger {
      events  = [before_create, before_update]
      actions = [action.aws_ecs_update_service.test]
    }
  }
}
`)
}

func testAccUpdateServiceActionConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecs_cluster" "test" {
  name = %[1]q
}

resource "aws_ecs_task_definition" "test" {
  family = %[1]q
  
  container_definitions = jsonencode([
    {
      name  = "test"
      image = "nginx:latest"
      memory = 128
      essential = true
      portMappings = [
        {
          containerPort = 80
          protocol      = "tcp"
        }
      ]
    }
  ])
}

resource "aws_ecs_service" "test" {
  name            = %[1]q
  cluster         = aws_ecs_cluster.test.id
  task_definition = aws_ecs_task_definition.test.arn
  desired_count   = 1

  deployment_maximum_percent         = 200
  deployment_minimum_healthy_percent = 50
}
`, rName)
}
