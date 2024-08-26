// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAPIGatewayAuthorizersDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_api_gateway_authorizers.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAuthorizersDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "ids.#", acctest.Ct2),
				),
			},
		},
	})
}

func testAccAuthorizersDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccAuthorizerConfig_base(rName), fmt.Sprintf(`
resource "aws_api_gateway_authorizer" "test" {
  count = 2

  name                   = "%[1]s-${count.index}"
  rest_api_id            = aws_api_gateway_rest_api.test.id
  authorizer_uri         = aws_lambda_function.test.invoke_arn
  authorizer_credentials = aws_iam_role.test.arn
}

data "aws_api_gateway_authorizers" "test" {
  rest_api_id = aws_api_gateway_rest_api.test.id

  depends_on = [aws_api_gateway_authorizer.test[0], aws_api_gateway_authorizer.test[1]]
}
`, rName))
}
