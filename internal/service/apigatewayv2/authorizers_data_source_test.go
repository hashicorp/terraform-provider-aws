// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigatewayv2_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAPIGatewayV2AuthorizersDataSource_id(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_apigatewayv2_authorizers.test"
	resourceName := "aws_apigatewayv2_authorizer.test"
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName3 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIAuthorizersDataSourceConfig_id(rName1, rName2, rName3),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "api_id", resourceName, "api_id"),
					resource.TestCheckResourceAttr(dataSourceName, "ids.#", "3"),
				),
			},
		},
	})
}

func testAccAuthorizerConfig_moreAuthorizers(rName1, rName2 string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_authorizer" "test2" {
  api_id          = aws_apigatewayv2_api.test.id
  authorizer_type = "REQUEST"
  authorizer_uri  = aws_lambda_function.test.invoke_arn
  name            = %[1]q
}

resource "aws_apigatewayv2_authorizer" "test3" {
  api_id          = aws_apigatewayv2_api.test.id
  authorizer_type = "REQUEST"
  authorizer_uri  = aws_lambda_function.test.invoke_arn
  name            = %[2]q
}

`, rName1, rName2)
}

func testAccAPIAuthorizersDataSourceConfig_id(rName1, rName2, rName3 string) string {
	return acctest.ConfigCompose(
		testAccAuthorizerConfig_basic(rName1),
		testAccAuthorizerConfig_moreAuthorizers(rName2, rName3), `
data "aws_apigatewayv2_authorizers" "test" {
  api_id = aws_apigatewayv2_api.test.id

  depends_on = [
    aws_apigatewayv2_authorizer.test,
    aws_apigatewayv2_authorizer.test2,
    aws_apigatewayv2_authorizer.test3,
  ]
}

`)
}
