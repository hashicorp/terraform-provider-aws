// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tfecs "github.com/hashicorp/terraform-provider-aws/internal/service/ecs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccECSExpressGatewayService_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var service awstypes.ECSExpressGatewayService
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_express_gateway_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ECSEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"time": {
				Source: "hashicorp/time",
			},
		},
		CheckDestroy: testAccCheckExpressGatewayServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccExpressGatewayServiceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckExpressGatewayServiceExists(ctx, resourceName, &service),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrServiceARN),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrServiceName),
					resource.TestCheckResourceAttrSet(resourceName, "created_at"),
					resource.TestCheckResourceAttr(resourceName, "wait_for_steady_state", "false"),
					resource.TestCheckResourceAttr(resourceName, "primary_container.0.image", "public.ecr.aws/nginx/nginx:1.28-alpine3.21-slim"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrServiceARN, "ecs", regexache.MustCompile(`service/.+$`)),
				),
			},
			{
				ResourceName: resourceName,
				ImportState:  true,
				// ImportStateVerify: true,
				// ImportStateVerifyIgnore: []string{
				// 	"wait_for_steady_state",
				// 	"current_deployment", // is not returned in Create response, so will show as diff in Import.
				// },
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_express_gateway_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"time": {
				Source: "hashicorp/time",
			},
		},
		CheckDestroy: testAccCheckExpressGatewayServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccExpressGatewayServiceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExpressGatewayServiceExists(ctx, resourceName, &service),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tfecs.ResourceExpressGatewayService, resourceName),
				),
				ExpectNonEmptyPlan: true,
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_express_gateway_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ECSEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"time": {
				Source: "hashicorp/time",
			},
		},
		CheckDestroy: testAccCheckExpressGatewayServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccExpressGatewayServiceConfig_tags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExpressGatewayServiceExists(ctx, resourceName, &service),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
				ExpectNonEmptyPlan: true,
			},
			// {
			// 	ResourceName: resourceName,
			// 	ImportState:  true,

			// 	// ImportStateVerify: true,
			// 	// ImportStateVerifyIgnore: []string{
			// 	// 	"wait_for_steady_state",
			// 	// 	"current_deployment", // is not returned in Create response, so will show as diff in Import.
			// 	// },
			// },
		},
	})
}

func TestAccECSExpressGatewayService_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var service1, service2 awstypes.ECSExpressGatewayService
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_express_gateway_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ECSEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"time": {
				Source: "hashicorp/time",
			},
		},
		CheckDestroy: testAccCheckExpressGatewayServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccExpressGatewayServiceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExpressGatewayServiceExists(ctx, resourceName, &service1),
				),
			},
			{
				Config: testAccExpressGatewayServiceConfig_updated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExpressGatewayServiceExists(ctx, resourceName, &service2),
					resource.TestCheckResourceAttr(resourceName, "primary_container.0.image", "public.ecr.aws/nginx/nginx:latest"),
				),
			},
		},
	})
}

// func TestAccECSExpressGatewayService_waitForSteadyState(t *testing.T) {
// 	ctx := acctest.Context(t)
// 	if testing.Short() {
// 		t.Skip("skipping long-running test in short mode")
// 	}

// 	var service awstypes.ECSExpressGatewayService
// 	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
// 	resourceName := "aws_ecs_express_gateway_service.test"

// 	resource.ParallelTest(t, resource.TestCase{
// 		PreCheck: func() {
// 			acctest.PreCheck(ctx, t)
// 			acctest.PreCheckPartitionHasService(t, names.ECSEndpointID)
// 		},
// 		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
// 		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
// 		ExternalProviders: map[string]resource.ExternalProvider{
// 			"time": {
// 				Source: "hashicorp/time",
// 			},
// 		},
// 		CheckDestroy: testAccCheckExpressGatewayServiceDestroy(ctx),
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccExpressGatewayServiceConfig_waitForSteadyState(rName, true),
// 				Check: resource.ComposeTestCheckFunc(
// 					testAccCheckExpressGatewayServiceExists(ctx, resourceName, &service),
// 					resource.TestCheckResourceAttr(resourceName, "wait_for_steady_state", acctest.CtTrue),
// 				),
// 			},
// 			{
// 				ResourceName: resourceName,
// 				ImportState:  true,
// 				// ImportStateVerify: true,
// 				// ImportStateVerifyIgnore: []string{
// 				// 	"wait_for_steady_state",
// 				// 	"current_deployment", // is not returned in Create response, so will show as diff in Import.
// 				// },
// 			},
// 		},
// 	})
// }

func TestAccECSExpressGatewayService_checkIdempotency(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_ecs_express_gateway_service.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ECSEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ECSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"time": {
				Source: "hashicorp/time",
			},
		},
		CheckDestroy: testAccCheckExpressGatewayServiceDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccExpressGatewayServiceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckExpressGatewayServiceExists(ctx, resourceName, &awstypes.ECSExpressGatewayService{}),
				),
			},
			{
				Config:      testAccExpressGatewayServiceConfig_duplicate(rName),
				ExpectError: regexache.MustCompile("Express Gateway Service .* already exists in cluster"),
			},
		},
	})
}

func testAccCheckExpressGatewayServiceDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ECSClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ecs_express_gateway_service" {
				continue
			}

			output, err := tfecs.FindExpressGatewayServiceByARN(ctx, conn, rs.Primary.ID)
			if tfresource.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.ECS, create.ErrActionCheckingDestroyed, tfecs.ResNameExpressGatewayService, rs.Primary.ID, err)
			}

			if string(output.Status.StatusCode) == "INACTIVE" {
				return nil
			}

			return create.Error(names.ECS, create.ErrActionCheckingDestroyed, tfecs.ResNameExpressGatewayService, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckExpressGatewayServiceExists(ctx context.Context, name string, service *awstypes.ECSExpressGatewayService) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.ECS, create.ErrActionCheckingExistence, tfecs.ResNameExpressGatewayService, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.ECS, create.ErrActionCheckingExistence, tfecs.ResNameExpressGatewayService, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ECSClient(ctx)

		resp, err := tfecs.FindExpressGatewayServiceByARN(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.ECS, create.ErrActionCheckingExistence, tfecs.ResNameExpressGatewayService, rs.Primary.ID, err)
		}

		*service = *resp

		return nil
	}
}

// func testAccCheckExpressGatewayServiceNotRecreated(before, after *awstypes.ECSExpressGatewayService) resource.TestCheckFunc {
// 	return func(s *terraform.State) error {
// 		if before, after := aws.ToString(before.ServiceArn), aws.ToString(after.ServiceArn); before != after {
// 			return create.Error(names.ECS, create.ErrActionCheckingNotRecreated, tfecs.ResNameExpressGatewayService, before, errors.New("recreated"))
// 		}

// 		return nil
// 	}
// }

func testAccExpressGatewayServiceConfig_base(rName string) string {
	return fmt.Sprintf(`
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

data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_iam_role" "execution" {
  name = "%[1]s-execution"
  assume_role_policy = <<POLICY
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ecs-tasks-gamma.amazonaws.com",
          "ecs-tasks.amazonaws.com",
					"ecs-slr.aws.internal"
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
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
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
  name = "%[1]s-infra"
  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ecs-service-principal.ecs.aws.internal",
          "ecs.amazonaws.com",
					"ecs-slr.aws.internal"
        ]
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy" "infrastructure" {
  name = "EcsExpressServicesInfraPolicy"
  role = aws_iam_role.infrastructure.id
  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "CreateSLRForAutoscaling",
      "Effect": "Allow",
      "Action": "iam:CreateServiceLinkedRole",
      "Resource": "arn:aws:iam::*:role/aws-service-role/ecs.application-autoscaling.amazonaws.com/AWSServiceRoleForApplicationAutoScaling_ECSService",
      "Condition": {
        "StringEquals": {
          "iam:AWSServiceName": "ecs.application-autoscaling.amazonaws.com"
        }
      }
    },
    {
      "Sid": "ELBOperations",
      "Effect": "Allow",
      "Action": [
        "elasticloadbalancing:CreateListener",
        "elasticloadbalancing:CreateLoadBalancer",
        "elasticloadbalancing:CreateRule",
        "elasticloadbalancing:CreateTargetGroup",
        "elasticloadbalancing:ModifyListener",
        "elasticloadbalancing:ModifyRule",
        "elasticloadbalancing:AddListenerCertificates",
        "elasticloadbalancing:RemoveListenerCertificates",
        "elasticloadbalancing:RegisterTargets",
        "elasticloadbalancing:DeregisterTargets",
        "elasticloadbalancing:DeleteTargetGroup",
        "elasticloadbalancing:DeleteLoadBalancer",
        "elasticloadbalancing:DeleteRule",
        "elasticloadbalancing:DeleteListener"
      ],
      "Resource": [
        "arn:aws:elasticloadbalancing:*:*:loadbalancer/app/*/*",
        "arn:aws:elasticloadbalancing:*:*:listener/app/*/*/*",
        "arn:aws:elasticloadbalancing:*:*:listener-rule/app/*/*/*/*",
        "arn:aws:elasticloadbalancing:*:*:targetgroup/*/*"
      ],
      "Condition": {
        "StringEquals": {
          "aws:ResourceTag/AmazonECSManaged": "true"
        }
      }
    },
    {
      "Sid": "TagOnCreateELBResources",
      "Effect": "Allow",
      "Action": "elasticloadbalancing:AddTags",
      "Resource": [
        "arn:aws:elasticloadbalancing:*:*:loadbalancer/app/*/*",
        "arn:aws:elasticloadbalancing:*:*:listener/app/*/*/*",
        "arn:aws:elasticloadbalancing:*:*:listener-rule/app/*/*/*/*",
        "arn:aws:elasticloadbalancing:*:*:targetgroup/*/*"
      ],
      "Condition": {
        "StringEquals": {
          "elasticloadbalancing:CreateAction": [
            "CreateLoadBalancer",
            "CreateListener",
            "CreateRule",
            "CreateTargetGroup"
          ]
        }
      }
    },
    {
      "Sid": "BlanketAllowCreateSecurityGroupsInVPCs",
      "Effect": "Allow",
      "Action": "ec2:CreateSecurityGroup",
      "Resource": "arn:aws:ec2:*:*:vpc/*"
    },
    {
      "Sid": "CreateSecurityGroupResourcesWithTags",
      "Effect": "Allow",
      "Action": [
        "ec2:CreateSecurityGroup",
        "ec2:AuthorizeSecurityGroupEgress",
        "ec2:AuthorizeSecurityGroupIngress"
      ],
      "Resource": [
        "arn:aws:ec2:*:*:security-group/*",
        "arn:aws:ec2:*:*:security-group-rule/*",
        "arn:aws:ec2:*:*:vpc/*"
      ],
      "Condition": {
        "StringEquals": {
          "aws:RequestTag/AmazonECSManaged": "true"
        }
      }
    },
    {
      "Sid": "ModifySecurityGroupOperations",
      "Effect": "Allow",
      "Action": [
        "ec2:AuthorizeSecurityGroupEgress",
        "ec2:AuthorizeSecurityGroupIngress",
        "ec2:DeleteSecurityGroup",
        "ec2:RevokeSecurityGroupEgress",
        "ec2:RevokeSecurityGroupIngress"
      ],
      "Resource": [
        "arn:aws:ec2:*:*:security-group/*",
        "arn:aws:ec2:*:*:vpc/*"
      ],
      "Condition": {
        "StringEquals": {
          "aws:ResourceTag/AmazonECSManaged": "true"
        }
      }
    },
    {
      "Sid": "TagOnCreateEC2Resources",
      "Effect": "Allow",
      "Action": "ec2:CreateTags",
      "Resource": [
        "arn:aws:ec2:*:*:security-group/*",
        "arn:aws:ec2:*:*:security-group-rule/*"
      ],
      "Condition": {
        "StringEquals": {
          "ec2:CreateAction": [
            "CreateSecurityGroup",
            "AuthorizeSecurityGroupIngress",
            "AuthorizeSecurityGroupEgress"
          ]
        }
      }
    },
    {
      "Sid": "CertificateOperations",
      "Effect": "Allow",
      "Action": [
        "acm:RequestCertificate",
        "acm:AddTagsToCertificate",
        "acm:DeleteCertificate",
        "acm:DescribeCertificate"
      ],
      "Resource": "arn:aws:acm:*:*:certificate/*",
      "Condition": {
        "StringEquals": {
          "aws:ResourceTag/AmazonECSManaged": "true"
        }
      }
    },
    {
      "Sid": "ApplicationAutoscalingCreateOperations",
      "Effect": "Allow",
      "Action": [
        "application-autoscaling:RegisterScalableTarget",
        "application-autoscaling:TagResource",
        "application-autoscaling:DeregisterScalableTarget"
      ],
      "Resource": "arn:aws:application-autoscaling:*:*:scalable-target/*",
      "Condition": {
        "StringEquals": {
          "aws:ResourceTag/AmazonECSManaged": "true"
        }
      }
    },
    {
      "Sid": "ApplicationAutoscalingPolicyOperations",
      "Effect": "Allow",
      "Action": [
        "application-autoscaling:PutScalingPolicy",
        "application-autoscaling:DeleteScalingPolicy"
      ],
      "Resource": "arn:aws:application-autoscaling:*:*:scalable-target/*",
      "Condition": {
        "StringEquals": {
          "application-autoscaling:service-namespace": "ecs"
        }
      }
    },
    {
      "Sid": "ApplicationAutoscalingReadOperations",
      "Effect": "Allow",
      "Action": [
        "application-autoscaling:DescribeScalableTargets",
        "application-autoscaling:DescribeScalingPolicies",
        "application-autoscaling:DescribeScalingActivities"
      ],
      "Resource": "arn:aws:application-autoscaling:*:*:scalable-target/*"
    },
    {
      "Sid": "CloudWatchAlarmCreateOperations",
      "Effect": "Allow",
      "Action": "cloudwatch:PutMetricAlarm",
      "Resource": "arn:aws:cloudwatch:*:*:alarm:*",
      "Condition": {
        "StringEquals": {
          "aws:RequestTag/AmazonECSManaged": "true"
        }
      }
    },
    {
      "Sid": "TagOnCreateCloudWatchAlarms",
      "Effect": "Allow",
      "Action": "cloudwatch:TagResource",
      "Resource": "arn:aws:cloudwatch:*:*:alarm:*",
      "Condition": {
        "StringEquals": {
          "cloudwatch:CreateAction": "PutMetricAlarm"
        }
      }
    },
    {
      "Sid": "CloudWatchAlarmOperations",
      "Effect": "Allow",
      "Action": [
        "cloudwatch:DeleteAlarms",
        "cloudwatch:DescribeAlarms"
      ],
      "Resource": "arn:aws:cloudwatch:*:*:alarm:*",
      "Condition": {
        "StringEquals": {
          "aws:ResourceTag/AmazonECSManaged": "true"
        }
      }
    },
    {
      "Sid": "ELBReadOperations",
      "Effect": "Allow",
      "Action": [
        "elasticloadbalancing:DescribeLoadBalancers",
        "elasticloadbalancing:DescribeTargetGroups",
        "elasticloadbalancing:DescribeListeners",
        "elasticloadbalancing:DescribeRules"
      ],
      "Resource": "*"
    },
    {
      "Sid": "VPCReadOperations",
      "Effect": "Allow",
      "Action": [
        "ec2:DescribeSecurityGroups",
        "ec2:DescribeSubnets",
        "ec2:DescribeRouteTables",
        "ec2:DescribeVpcs"
      ],
      "Resource": "*"
    },
    {
      "Sid": "CloudWatchLogsCreateOperations",
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogGroup",
        "logs:TagResource"
      ],
      "Resource": "arn:aws:logs:*:*:log-group:*",
      "Condition": {
        "StringEquals": {
          "aws:RequestTag/AmazonECSManaged": "true"
        }
      }
    },
    {
      "Sid": "CloudWatchLogsReadOperations",
      "Effect": "Allow",
      "Action": "logs:DescribeLogGroups",
      "Resource": "*"
    },
    {
      "Sid": "ECSDeleteOperation",
      "Effect": "Allow",
      "Action": "ecs:DeleteService",
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "time_sleep" "wait_for_iam" {
  depends_on = [aws_iam_role_policy.infrastructure]
  create_duration = "10s"
}
`, rName)
}

func testAccExpressGatewayServiceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccExpressGatewayServiceConfig_base(rName), fmt.Sprintf(`
resource "aws_ecs_express_gateway_service" "test" {
  execution_role_arn      = aws_iam_role.execution.arn
  infrastructure_role_arn = aws_iam_role.infrastructure.arn

  primary_container {
    image = "public.ecr.aws/nginx/nginx:1.28-alpine3.21-slim"
  }

  depends_on = [
    time_sleep.wait_for_iam
  ]
}
`))
}

func testAccExpressGatewayServiceConfig_updated(rName string) string {
	return acctest.ConfigCompose(testAccExpressGatewayServiceConfig_base(rName), fmt.Sprintf(`
resource "aws_ecs_express_gateway_service" "test" {
  execution_role_arn      = aws_iam_role.execution.arn
  infrastructure_role_arn = aws_iam_role.infrastructure.arn

  primary_container {
    image = "public.ecr.aws/nginx/nginx:latest"
  }

  depends_on = [
    time_sleep.wait_for_iam
  ]
}
`))
}

func testAccExpressGatewayServiceConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccExpressGatewayServiceConfig_base(rName), fmt.Sprintf(`
resource "aws_ecs_express_gateway_service" "test" {
  execution_role_arn      = aws_iam_role.execution.arn
  infrastructure_role_arn = aws_iam_role.infrastructure.arn

  primary_container {
    image = "public.ecr.aws/nginx/nginx:1.28-alpine3.21-slim"
  }

  network_configuration {
    subnets         = data.aws_subnets.default.ids
    security_groups = [data.aws_security_group.default.id]
  }

  tags = {
    %[2]q = %[3]q
  }

  depends_on = [
    time_sleep.wait_for_iam
  ]
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

  network_configuration {
    subnets         = data.aws_subnets.default.ids
    security_groups = [data.aws_security_group.default.id]
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

  depends_on = [
    time_sleep.wait_for_iam
  ]
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccExpressGatewayServiceConfig_waitForSteadyState(rName string, waitForSteadyState bool) string {
	return acctest.ConfigCompose(testAccExpressGatewayServiceConfig_base(rName), fmt.Sprintf(`
resource "aws_ecs_express_gateway_service" "test" {
  execution_role_arn      = aws_iam_role.execution.arn
  infrastructure_role_arn = aws_iam_role.infrastructure.arn
  wait_for_steady_state   = %[2]t

  primary_container {
    image = "public.ecr.aws/nginx/nginx:1.28-alpine3.21-slim"
  }

  network_configuration {
    subnets         = data.aws_subnets.default.ids
    security_groups = [data.aws_security_group.default.id]
  }

  depends_on = [
    time_sleep.wait_for_iam
  ]
}
`, rName, waitForSteadyState))
}

func testAccExpressGatewayServiceConfig_duplicate(rName string) string {
	return acctest.ConfigCompose(testAccExpressGatewayServiceConfig_base(rName), fmt.Sprintf(`
resource "aws_ecs_express_gateway_service" "test" {
  execution_role_arn      = aws_iam_role.execution.arn
  infrastructure_role_arn = aws_iam_role.infrastructure.arn

  primary_container {
    image = "public.ecr.aws/nginx/nginx:1.28-alpine3.21-slim"
  }

  depends_on = [
    time_sleep.wait_for_iam
  ]
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
`))
}
