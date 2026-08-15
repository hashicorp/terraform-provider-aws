// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lambdacore_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/lambdacore"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lambdacore/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tflambdacore "github.com/hashicorp/terraform-provider-aws/internal/service/lambdacore"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLambdaCoreNetworkConnector_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v lambdacore.GetNetworkConnectorOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lambdacore_network_connector.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkConnectorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkConnectorConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkConnectorExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrState, string(awstypes.NetworkConnectorStateActive)),
					resource.TestCheckResourceAttr(resourceName, "configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.vpc_egress_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.vpc_egress_configuration.0.subnet_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "configuration.0.vpc_egress_configuration.0.security_group_ids.#", "1"),
					acctest.MatchResourceAttrRegionalARN(ctx, resourceName, names.AttrARN, "lambda", regexache.MustCompile(`network-connector:.+$`)),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
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

func TestAccLambdaCoreNetworkConnector_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v lambdacore.GetNetworkConnectorOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lambdacore_network_connector.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LambdaCoreServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckNetworkConnectorDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkConnectorConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckNetworkConnectorExists(ctx, t, resourceName, &v),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tflambdacore.ResourceNetworkConnector, resourceName),
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

func testAccCheckNetworkConnectorDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).LambdaCoreClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lambdacore_network_connector" {
				continue
			}

			_, err := tflambdacore.FindNetworkConnectorByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Lambda Core Network Connector %s still exists", rs.Primary.Attributes[names.AttrARN])
		}

		return nil
	}
}

func testAccCheckNetworkConnectorExists(ctx context.Context, t *testing.T, n string, v *lambdacore.GetNetworkConnectorOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).LambdaCoreClient(ctx)

		output, err := tflambdacore.FindNetworkConnectorByARN(ctx, conn, rs.Primary.Attributes[names.AttrARN])
		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).LambdaCoreClient(ctx)

	input := lambdacore.ListNetworkConnectorsInput{}

	_, err := conn.ListNetworkConnectors(ctx, &input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

// A network connector needs VPC subnets, a security group, and an operator role
// that the connector service assumes to manage ENIs in the VPC.
func testAccNetworkConnectorConfig_base(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

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

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "network-connectors.lambda.amazonaws.com"
      }
    }]
  })
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "CreateENI"
        Effect = "Allow"
        Action = "ec2:CreateNetworkInterface"
        Resource = [
          "arn:${data.aws_partition.current.partition}:ec2:*:*:network-interface/*",
          "arn:${data.aws_partition.current.partition}:ec2:*:*:subnet/*",
          "arn:${data.aws_partition.current.partition}:ec2:*:*:security-group/*",
        ]
      },
      {
        Sid      = "TagENI"
        Effect   = "Allow"
        Action   = "ec2:CreateTags"
        Resource = "arn:${data.aws_partition.current.partition}:ec2:*:*:network-interface/*"
        Condition = {
          StringEquals = {
            "ec2:ManagedResourceOperator" = "network-connectors.lambda.amazonaws.com"
          }
        }
      },
    ]
  })
}

data "aws_partition" "current" {}
`, rName))
}

func testAccNetworkConnectorConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccNetworkConnectorConfig_base(rName), fmt.Sprintf(`
resource "aws_lambdacore_network_connector" "test" {
  name          = %[1]q
  operator_role = aws_iam_role.test.arn

  configuration {
    vpc_egress_configuration {
      subnet_ids         = aws_subnet.test[*].id
      security_group_ids = [aws_security_group.test.id]
    }
  }

  depends_on = [aws_iam_role_policy.test]
}
`, rName))
}
