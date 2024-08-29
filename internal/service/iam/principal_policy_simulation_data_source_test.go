// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIAMPrincipalPolicySimulationDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IAMServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPrincipalPolicySimulationDataSourceConfig_main(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_iam_principal_policy_simulation.allow_simple", "all_allowed", acctest.CtTrue),
					resource.TestCheckResourceAttr("data.aws_iam_principal_policy_simulation.allow_simple", "results.#", acctest.Ct1),
					resource.TestCheckResourceAttr("data.aws_iam_principal_policy_simulation.allow_simple", "results.0.action_name", "ec2:AssociateVpcCidrBlock"),
					resource.TestCheckResourceAttr("data.aws_iam_principal_policy_simulation.allow_simple", "results.0.allowed", acctest.CtTrue),
					resource.TestCheckResourceAttr("data.aws_iam_principal_policy_simulation.allow_simple", "results.0.decision", "allowed"),

					// IAM seems to generate the SourcePolicyId by concatenating
					// together the username, the policy name, and some other
					// hard-coded bits. Not sure if this is constractual, so
					// if this turns out to change in future it may be better
					// to test this in a different way.
					resource.TestCheckResourceAttr("data.aws_iam_principal_policy_simulation.allow_simple", "results.0.matched_statements.0.source_policy_id", fmt.Sprintf("user_%s_%s", rName, rName)),
					resource.TestCheckResourceAttr("data.aws_iam_principal_policy_simulation.allow_simple", "results.0.matched_statements.0.source_policy_type", "IAM Policy"),

					resource.TestCheckResourceAttr("data.aws_iam_principal_policy_simulation.deny_explicit", "all_allowed", acctest.CtFalse),
					resource.TestCheckResourceAttr("data.aws_iam_principal_policy_simulation.deny_explicit", "results.#", acctest.Ct1),
					resource.TestCheckResourceAttr("data.aws_iam_principal_policy_simulation.deny_explicit", "results.0.action_name", "ec2:AttachClassicLinkVpc"),
					resource.TestCheckResourceAttr("data.aws_iam_principal_policy_simulation.deny_explicit", "results.0.allowed", acctest.CtFalse),
					resource.TestCheckResourceAttr("data.aws_iam_principal_policy_simulation.deny_explicit", "results.0.decision", "explicitDeny"),

					resource.TestCheckResourceAttr("data.aws_iam_principal_policy_simulation.deny_implicit", "all_allowed", acctest.CtFalse),
					resource.TestCheckResourceAttr("data.aws_iam_principal_policy_simulation.deny_implicit", "results.#", acctest.Ct1),
					resource.TestCheckResourceAttr("data.aws_iam_principal_policy_simulation.deny_implicit", "results.0.action_name", "ec2:AttachVpnGateway"),
					resource.TestCheckResourceAttr("data.aws_iam_principal_policy_simulation.deny_implicit", "results.0.allowed", acctest.CtFalse),
					resource.TestCheckResourceAttr("data.aws_iam_principal_policy_simulation.deny_implicit", "results.0.decision", "implicitDeny"),

					resource.TestCheckResourceAttr("data.aws_iam_principal_policy_simulation.allow_with_context", "all_allowed", acctest.CtTrue),
					resource.TestCheckResourceAttr("data.aws_iam_principal_policy_simulation.allow_with_context", "results.#", acctest.Ct1),
					resource.TestCheckResourceAttr("data.aws_iam_principal_policy_simulation.allow_with_context", "results.0.action_name", "ec2:AttachInternetGateway"),
					resource.TestCheckResourceAttr("data.aws_iam_principal_policy_simulation.allow_with_context", "results.0.allowed", acctest.CtTrue),
					resource.TestCheckResourceAttr("data.aws_iam_principal_policy_simulation.allow_with_context", "results.0.decision", "allowed"),

					resource.TestCheckResourceAttr("data.aws_iam_principal_policy_simulation.allow_with_wrong_context", "all_allowed", acctest.CtFalse),
					resource.TestCheckResourceAttr("data.aws_iam_principal_policy_simulation.allow_with_wrong_context", "results.#", acctest.Ct1),
					resource.TestCheckResourceAttr("data.aws_iam_principal_policy_simulation.allow_with_wrong_context", "results.0.action_name", "ec2:AttachInternetGateway"),
					resource.TestCheckResourceAttr("data.aws_iam_principal_policy_simulation.allow_with_wrong_context", "results.0.allowed", acctest.CtFalse),
					resource.TestCheckResourceAttr("data.aws_iam_principal_policy_simulation.allow_with_wrong_context", "results.0.decision", "implicitDeny"),

					resource.TestCheckResourceAttr("data.aws_iam_principal_policy_simulation.multiple_mixed", "all_allowed", acctest.CtFalse),
					resource.TestCheckResourceAttr("data.aws_iam_principal_policy_simulation.multiple_mixed", "results.#", acctest.Ct2),

					resource.TestCheckResourceAttr("data.aws_iam_principal_policy_simulation.multiple_allow", "all_allowed", acctest.CtTrue),
					resource.TestCheckResourceAttr("data.aws_iam_principal_policy_simulation.multiple_allow", "results.#", acctest.Ct2),

					func(state *terraform.State) error {
						vpcARN := state.RootModule().Outputs["vpc_arn"].Value.(string)
						return resource.TestCheckResourceAttr("data.aws_iam_principal_policy_simulation.allow_simple", "results.0.resource_arn", vpcARN)(state)
					},
				),
			},
		},
	})
}

func testAccPrincipalPolicySimulationDataSourceConfig_main(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "test" {
  name = %[1]q
}

resource "aws_vpc" "test" {
  cidr_block = "192.168.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_iam_user_policy" "test" {
  name = %[1]q
  user = aws_iam_user.test.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action   = "ec2:AssociateVpcCidrBlock"
        Effect   = "Allow"
        Resource = aws_vpc.test.arn
      },
      {
        Action   = "ec2:AttachClassicLinkVpc"
        Effect   = "Deny"
        Resource = aws_vpc.test.arn
      },
      {
        Action   = "ec2:AttachInternetGateway"
        Effect   = "Allow"
        Resource = aws_vpc.test.arn
        Condition = {
          StringEquals = {
            "ec2:ResourceTag/Foo" = "bar"
          }
        }
      },
    ]
  })
}

data "aws_iam_principal_policy_simulation" "allow_simple" {
  action_names      = ["ec2:AssociateVpcCidrBlock"]
  resource_arns     = [aws_vpc.test.arn]
  policy_source_arn = aws_iam_user.test.arn

  depends_on = [aws_iam_user_policy.test]
}

data "aws_iam_principal_policy_simulation" "deny_explicit" {
  action_names      = ["ec2:AttachClassicLinkVpc"]
  resource_arns     = [aws_vpc.test.arn]
  policy_source_arn = aws_iam_user.test.arn

  depends_on = [aws_iam_user_policy.test]
}

data "aws_iam_principal_policy_simulation" "deny_implicit" {
  # This one is implicit deny because our policy
  # doesn't mention ec2:AttachVpnGateway at all.
  action_names      = ["ec2:AttachVpnGateway"]
  resource_arns     = [aws_vpc.test.arn]
  policy_source_arn = aws_iam_user.test.arn

  depends_on = [aws_iam_user_policy.test]
}

data "aws_iam_principal_policy_simulation" "allow_with_context" {
  action_names      = ["ec2:AttachInternetGateway"]
  resource_arns     = [aws_vpc.test.arn]
  policy_source_arn = aws_iam_user.test.arn

  context {
    key    = "ec2:ResourceTag/Foo"
    type   = "string"
    values = ["bar"]
  }

  depends_on = [aws_iam_user_policy.test]
}

data "aws_iam_principal_policy_simulation" "allow_with_wrong_context" {
  action_names      = ["ec2:AttachInternetGateway"]
  resource_arns     = [aws_vpc.test.arn]
  policy_source_arn = aws_iam_user.test.arn

  context {
    key    = "ec2:ResourceTag/Foo"
    type   = "string"
    values = ["baz"]
  }

  depends_on = [aws_iam_user_policy.test]
}

data "aws_iam_principal_policy_simulation" "multiple_mixed" {
  action_names = [
    "ec2:AssociateVpcCidrBlock",
    "ec2:AttachClassicLinkVpc",
  ]
  resource_arns     = [aws_vpc.test.arn]
  policy_source_arn = aws_iam_user.test.arn

  depends_on = [aws_iam_user_policy.test]
}

data "aws_iam_principal_policy_simulation" "multiple_allow" {
  action_names = [
    "ec2:AssociateVpcCidrBlock",
    "ec2:AttachInternetGateway",
  ]
  resource_arns     = [aws_vpc.test.arn]
  policy_source_arn = aws_iam_user.test.arn

  context {
    key    = "ec2:ResourceTag/Foo"
    type   = "string"
    values = ["bar"]
  }

  depends_on = [aws_iam_user_policy.test]
}

output "vpc_arn" {
  value = aws_vpc.test.arn
}
`, rName)
}
