// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfapigateway "github.com/hashicorp/terraform-provider-aws/internal/service/apigateway"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccAccount_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_account.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAccountDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAccountConfig_role0(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_role_arn", "aws_iam_role.test.0", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "throttle_settings.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "api_key_version"),
					resource.TestCheckResourceAttrSet(resourceName, "features.#"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAccountConfig_role1(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_role_arn", "aws_iam_role.test.1", names.AttrARN),
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

func testAccCheckAccountDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).APIGatewayClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_api_gateway_stage" {
				continue
			}

			account, err := tfapigateway.FindAccount(ctx, conn)
			if err != nil {
				return err
			}

			if account.CloudwatchRoleArn == nil {
				// Settings have been reset
				continue
			}

			return fmt.Errorf("API Gateway Stage %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

const testAccAccountConfig_empty = `
resource "aws_api_gateway_account" "test" {}
`

func testAccAccountConfig_base(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  count = 2

  name = "%[1]s-${count.index}"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "apigateway.amazonaws.com"
      }
    }]
  })

  managed_policy_arns = ["arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonAPIGatewayPushToCloudWatchLogs"]
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
