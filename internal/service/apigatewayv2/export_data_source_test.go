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

func TestAccAPIGatewayV2ExportDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_apigatewayv2_export.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccExportDataSourceConfig_httpBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "api_id", "aws_apigatewayv2_route.test", "api_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "body"),
				),
			},
		},
	})
}

func TestAccAPIGatewayV2ExportDataSource_stage(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_apigatewayv2_export.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccExportDataSourceConfig_httpStage(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "api_id", "aws_apigatewayv2_route.test", "api_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "stage_name", "aws_apigatewayv2_stage.test", names.AttrName),
					resource.TestCheckResourceAttrSet(dataSourceName, "body"),
				),
			},
		},
	})
}

func testAccExportHTTPDataSourceConfigBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  name          = %[1]q
  protocol_type = "HTTP"
}

resource "aws_apigatewayv2_integration" "test" {
  api_id           = aws_apigatewayv2_api.test.id
  integration_type = "HTTP_PROXY"

  integration_method = "GET"
  integration_uri    = "https://example.com"
}

resource "aws_apigatewayv2_route" "test" {
  api_id    = aws_apigatewayv2_api.test.id
  route_key = "GET /test"
  target    = "integrations/${aws_apigatewayv2_integration.test.id}"
}
`, rName)
}

func testAccExportDataSourceConfig_httpBasic(rName string) string {
	return acctest.ConfigCompose(testAccExportHTTPDataSourceConfigBase(rName), `
data "aws_apigatewayv2_export" "test" {
  api_id        = aws_apigatewayv2_route.test.api_id
  specification = "OAS30"
  output_type   = "JSON"
}
`)
}

func testAccExportDataSourceConfig_httpStage(rName string) string {
	return acctest.ConfigCompose(testAccExportHTTPDataSourceConfigBase(rName), fmt.Sprintf(`
resource "aws_apigatewayv2_stage" "test" {
  api_id        = aws_apigatewayv2_deployment.test.api_id
  name          = %[1]q
  deployment_id = aws_apigatewayv2_deployment.test.id
}

resource "aws_apigatewayv2_deployment" "test" {
  api_id      = aws_apigatewayv2_route.test.api_id
  description = %[1]q
}

data "aws_apigatewayv2_export" "test" {
  api_id        = aws_apigatewayv2_api.test.id
  specification = "OAS30"
  output_type   = "JSON"
  stage_name    = aws_apigatewayv2_stage.test.name
}
`, rName))
}
