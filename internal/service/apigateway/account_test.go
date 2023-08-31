// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/apigateway"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccAPIGatewayAccount_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_account.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccAccountConfig_role0(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_role_arn", "aws_iam_role.test.0", "arn"),
					resource.TestCheckResourceAttr(resourceName, "throttle_settings.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"cloudwatch_role_arn"},
			},
			{
				Config: testAccAccountConfig_role1(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_role_arn", "aws_iam_role.test.1", "arn"),
					resource.TestCheckResourceAttr(resourceName, "throttle_settings.#", "1"),
				),
			},
			{
				Config: testAccAccountConfig_empty,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_role_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "throttle_settings.#", "1"),
				),
			},
		},
	})
}

const testAccAccountConfig_empty = `
resource "aws_api_gateway_account" "test" {}
`

func testAccAccountConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  count = 2

  name = "%[1]s-${count.index}"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Sid": "",
    "Effect": "Allow",
    "Principal": {
      "Service": "apigateway.amazonaws.com"
    },
    "Action": "sts:AssumeRole"
  }]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  count = 2

  name = "%[1]s-${count.index}"
  role = aws_iam_role.test[count.index].id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": [
      "logs:CreateLogGroup",
      "logs:CreateLogStream",
      "logs:DescribeLogGroups",
      "logs:DescribeLogStreams",
      "logs:PutLogEvents",
      "logs:GetLogEvents",
      "logs:FilterLogEvents"
    ],
    "Resource": "*"
  }]
}
EOF
}
`, rName)
}

func testAccAccountConfig_role0(rName string) string {
	return acctest.ConfigCompose(testAccAccountConfig_base(rName), `
resource "aws_api_gateway_account" "test" {
  cloudwatch_role_arn = aws_iam_role.test[0].arn
}
`)
}

func testAccAccountConfig_role1(rName string) string {
	return acctest.ConfigCompose(testAccAccountConfig_base(rName), `
resource "aws_api_gateway_account" "test" {
  cloudwatch_role_arn = aws_iam_role.test[1].arn
}
`)
}
