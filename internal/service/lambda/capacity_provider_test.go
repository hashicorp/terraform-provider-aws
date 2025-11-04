// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lambda_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tflambda "github.com/hashicorp/terraform-provider-aws/internal/service/lambda"
	"github.com/hashicorp/terraform-provider-aws/internal/service/lambda/mocks"
	awsmocktypes "github.com/hashicorp/terraform-provider-aws/internal/service/lambda/mocks/types"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLambdaCapacityProvider_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var capacityprovider awsmocktypes.GetCapacityProviderOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_capacity_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LambdaEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapacityProviderDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityProviderConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCapacityProviderExists(ctx, resourceName, &capacityprovider),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "lambda", regexache.MustCompile(`capacityprovider:.+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
				),
			},
			{
				ResourceName:                         resourceName,
				ImportState:                          true,
				ImportStateVerify:                    true,
				ImportStateIdFunc:                    acctest.AttrImportStateIdFunc(resourceName, names.AttrARN),
				ImportStateVerifyIdentifierAttribute: names.AttrARN,
			},
		},
	})
}

func TestAccLambdaCapacityProvider_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var capacityprovider awsmocktypes.GetCapacityProviderOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lambda_capacity_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LambdaEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckCapacityProviderDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccCapacityProviderConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckCapacityProviderExists(ctx, resourceName, &capacityprovider),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tflambda.ResourceCapacityProvider, resourceName),
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

func testAccCheckCapacityProviderDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		//conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaClient(ctx)
		conn := mocks.LambdaClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lambda_capacity_provider" {
				continue
			}

			_, err := tflambda.FindCapacityProviderByARN(ctx, conn, rs.Primary.ID)
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

func testAccCheckCapacityProviderExists(ctx context.Context, name string, capacityprovider *awsmocktypes.GetCapacityProviderOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.Lambda, create.ErrActionCheckingExistence, tflambda.ResNameCapacityProvider, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.Lambda, create.ErrActionCheckingExistence, tflambda.ResNameCapacityProvider, name, errors.New("not set"))
		}

		//conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaClient(ctx)
		conn := mocks.LambdaClient(ctx)
		resp, err := tflambda.FindCapacityProviderByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])
		if err != nil {
			return create.Error(names.Lambda, create.ErrActionCheckingExistence, tflambda.ResNameCapacityProvider, rs.Primary.ID, err)
		}

		*capacityprovider = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	// conn := acctest.Provider.Meta().(*conns.AWSClient).LambdaClient(ctx)
	conn := mocks.LambdaClient(ctx)

	input := awsmocktypes.ListCapacityProvidersInput{}

	_, err := conn.ListCapacityProviders(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCapacityProviderConfig_base(policyName, roleName, sgName string) string {
	return fmt.Sprintf(`
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
        "ec2:UnassignPrivateIpAddresses"
"lambda.*"
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
  name = %[2]q

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
    Name = %[3]q
  }
}

resource "aws_subnet" "test" {
  count = 2

  vpc_id            = aws_vpc.vpc_for_lambda.id
  availability_zone = data.aws_availability_zones.available.names[1]

  cidr_block      = cidrsubnet(aws_vpc.vpc_for_lambda.cidr_block, 8, 1)
  ipv6_cidr_block = cidrsubnet(aws_vpc.vpc_for_lambda.ipv6_cidr_block, 8, 1)

  assign_ipv6_address_on_creation = true

  tags = {
    Name = %[3]q
  }
}

# This is defined here, rather than only in test cases where it's needed is to
# prevent a timeout issue when fully removing Lambda Filesystems
resource "aws_subnet" "test" {
  count = 2

  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[1]

  cidr_block      = cidrsubnet(aws_vpc.test.cidr_block, 8, 2)
  ipv6_cidr_block = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, 2)

  assign_ipv6_address_on_creation = true

  tags = {
    Name = %[3]q
  }
}

resource "aws_security_group" "test" {
  name        = %[3]q
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
    Name = %[3]q
  }
}
`, policyName, roleName, sgName)
}

func testaccCapacityProvider_baseEC2(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptInDefaultExclude(),
		fmt.Sprintf(`
data "aws_iam_policy_document" "test_assume_role_policy" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["ec2.${data.aws_partition.current.dns_suffix}"]
    }
  }
}

resource "aws_iam_role" "test_instance" {
  name               = %[1]q
  assume_role_policy = data.aws_iam_policy_document.test_assume_role_policy.json
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AdministratorAccess"
}

data "aws_iam_policy_document" "test_instance_assume_role_policy" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["ec2.${data.aws_partition.current.dns_suffix}"]
    }
  }
}

resource "aws_iam_role" "test_instance" {
  name               = "%[1]s-instance"
  assume_role_policy = data.aws_iam_policy_document.test_instance_assume_role_policy.json
}

resource "aws_iam_role_policy_attachment" "test_instance" {
  role       = aws_iam_role.test_instance.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonEC2ContainerServiceforEC2Role"
}

resource "aws_iam_instance_profile" "test" {
  name = %[1]q
  role = aws_iam_role.test_instance.name
}
`, rName))
}

func testAccCapacityProviderConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccCapacityProviderConfig_base(rName, rName, rName),
		testaccCapacityProvider_baseEC2(rName),
		fmt.Sprintf(`
resource "aws_lambda_capacity_provider" "test" {
  name = %[1]q

  vpc_config {
    subnet_ids = aws_subnet.subnet_for_lambda.*.id
  }

  permissions_config {
    instance_profile_arn           = aws_iam_instance_profile.test.arn
    capacity_provider_operator_arn = aws_iam_role.test.arn
  }
}
`, rName))
}
