// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambda_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tflambda "github.com/hashicorp/terraform-provider-aws/internal/service/lambda"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLambdaCapacityProvider_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var capacityprovider awstypes.CapacityProvider
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lambda_capacity_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccCapacityProviderPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapacityProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityProviderConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCapacityProviderExists(ctx, t, resourceName, &capacityprovider),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "lambda", regexache.MustCompile(`capacity-provider:.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnet_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", "1"),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccCheckCapacityProviderImportStateID(resourceName),
				ImportStateVerifyIdentifierAttribute: names.AttrName,
			},
		},
	})
}

func TestAccLambdaCapacityProvider_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var capacityprovider awstypes.CapacityProvider
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lambda_capacity_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccCapacityProviderPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapacityProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityProviderConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCapacityProviderExists(ctx, t, resourceName, &capacityprovider),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tflambda.ResourceCapacityProvider, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccLambdaCapacityProvider_instanceRequirements(t *testing.T) {
	ctx := acctest.Context(t)

	var capacityprovider awstypes.CapacityProvider
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lambda_capacity_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccCapacityProviderPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapacityProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityProviderConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCapacityProviderExists(ctx, t, resourceName, &capacityprovider),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "lambda", regexache.MustCompile(`capacity-provider:.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnet_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "permissions_config.#", "1"),
				),
			},
			{
				Config: testAccCapacityProviderConfig_instanceRequirements(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCapacityProviderExists(ctx, t, resourceName, &capacityprovider),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "lambda", regexache.MustCompile(`capacity-provider:.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnet_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.excluded_instance_types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_requirements.0.architectures.#", "1"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionReplace),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccCheckCapacityProviderImportStateID(resourceName),
				ImportStateVerifyIdentifierAttribute: names.AttrName,
			},
		},
	})
}

func TestAccLambdaCapacityProvider_scalingConfig(t *testing.T) {
	ctx := acctest.Context(t)

	var capacityprovider awstypes.CapacityProvider
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lambda_capacity_provider.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccCapacityProviderPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapacityProviderDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityProviderConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCapacityProviderExists(ctx, t, resourceName, &capacityprovider),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "lambda", regexache.MustCompile(`capacity-provider:.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnet_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", "1"),
				),
			},
			{
				Config: testAccCapacityProviderConfig_scalingConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCapacityProviderExists(ctx, t, resourceName, &capacityprovider),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "lambda", regexache.MustCompile(`capacity-provider:.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.subnet_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "vpc_config.0.security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "capacity_provider_scaling_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "capacity_provider_scaling_config.0.max_vcpu_count", "30"),
					resource.TestCheckResourceAttr(resourceName, "capacity_provider_scaling_config.0.scaling_mode", "Auto"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    testAccCheckCapacityProviderImportStateID(resourceName),
				ImportStateVerifyIdentifierAttribute: names.AttrName,
			},
		},
	})
}

func testAccCheckCapacityProviderImportStateID(n string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		return acctest.AttrImportStateIdFunc(n, names.AttrName)(s)
	}
}

func testAccCheckCapacityProviderDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).LambdaClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lambda_capacity_provider" {
				continue
			}

			_, err := tflambda.FindCapacityProviderByName(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				return nil
			}
			if err != nil {
				return create.Error(names.Lambda, create.ErrActionCheckingDestroyed, tflambda.ResNameCapacityProvider, rs.Primary.ID, err)
			}

			return create.Error(names.Lambda, create.ErrActionCheckingDestroyed, tflambda.ResNameCapacityProvider, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckCapacityProviderExists(ctx context.Context, t *testing.T, name string, capacityprovider *awstypes.CapacityProvider) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Lambda, create.ErrActionCheckingExistence, tflambda.ResNameCapacityProvider, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Lambda, create.ErrActionCheckingExistence, tflambda.ResNameCapacityProvider, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).LambdaClient(ctx)
		resp, err := tflambda.FindCapacityProviderByName(ctx, conn, rs.Primary.Attributes[names.AttrARN])
		if err != nil {
			return create.Error(names.Lambda, create.ErrActionCheckingExistence, tflambda.ResNameCapacityProvider, rs.Primary.ID, err)
		}

		*capacityprovider = *resp

		return nil
	}
}

func testAccCapacityProviderPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).LambdaClient(ctx)

	input := lambda.ListCapacityProvidersInput{}

	_, err := conn.ListCapacityProviders(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCapacityProviderConfig_base(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role_policy" "iam_policy_for_lambda" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "arn:${data.aws_partition.current.partition}:logs:*:*:*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "ec2:CreateNetworkInterface",
        "ec2:DescribeNetworkInterfaces",
        "ec2:DeleteNetworkInterface",
        "ec2:AssignPrivateIpAddresses",
        "ec2:UnassignPrivateIpAddresses",
        "ec2:DescribeSecurityGroups",
        "ec2:DescribeSubnets",
        "ec2:DescribeInstanceTypeOfferings",
        "ec2:RunInstances",
		"ec2:TerminateInstances",
        "ec2:AttachNetworkInterface"
      ],
      "Resource": [
        "*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "SNS:Publish"
      ],
      "Resource": [
        "*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "xray:PutTraceSegments"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF
}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_vpc" "test" {
  cidr_block                       = "10.0.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = 2

  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[1]

  cidr_block      = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  ipv6_cidr_block = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, count.index)

  assign_ipv6_address_on_creation = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name        = %[1]q
  description = "Allow all inbound traffic for lambda test"
  vpc_id      = aws_vpc.test.id

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccCapacityProviderConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccCapacityProviderConfig_base(rName),
		fmt.Sprintf(`
resource "aws_lambda_capacity_provider" "test" {
  name = %[1]q

  vpc_config {
    subnet_ids         = aws_subnet.test[*].id
    security_group_ids = [aws_security_group.test.id]
  }

  permissions_config {
    capacity_provider_operator_role_arn = aws_iam_role.test.arn
  }

  depends_on = [
    aws_iam_role_policy.iam_policy_for_lambda
  ]
}
`, rName))
}

func testAccCapacityProviderConfig_instanceRequirements(rName string) string {
	return acctest.ConfigCompose(
		testAccCapacityProviderConfig_base(rName),
		fmt.Sprintf(`
resource "aws_lambda_capacity_provider" "test" {
  name = %[1]q

  vpc_config {
    subnet_ids         = aws_subnet.test[*].id
    security_group_ids = [aws_security_group.test.id]
  }

  permissions_config {
    capacity_provider_operator_role_arn = aws_iam_role.test.arn
  }

  instance_requirements {
    excluded_instance_types = ["m5.8xlarge"]
    architectures           = ["x86_64"]
  }

  depends_on = [
    aws_iam_role_policy.iam_policy_for_lambda
  ]
}
`, rName))
}

func testAccCapacityProviderConfig_scalingConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccCapacityProviderConfig_base(rName),
		fmt.Sprintf(`
resource "aws_lambda_capacity_provider" "test" {
  name = %[1]q

  vpc_config {
    subnet_ids         = aws_subnet.test[*].id
    security_group_ids = [aws_security_group.test.id]
  }

  permissions_config {
    capacity_provider_operator_role_arn = aws_iam_role.test.arn
  }

  capacity_provider_scaling_config {
    scaling_mode   = "Auto"
    max_vcpu_count = 30
  }

  depends_on = [
    aws_iam_role_policy.iam_policy_for_lambda
  ]
}
`, rName))
}
