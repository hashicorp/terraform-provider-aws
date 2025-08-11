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

func TestAccAPIGatewayV2IntegrationDataSource_id(t *testing.T) {
	ctx := acctest.Context(t)
	dataSource1Name := "data.aws_apigatewayv2_integration.test1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIIntegrationsDataSourceConfig_id(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSource1Name, "ids.#", "2"),
				),
			},
			{
				Config: testAccAPIIntegrationsDataSourceConfig_no_ids(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSource1Name, "ids.#", "0"),
				),
			},
		},
	})
}

func testAccAPIIntegrationsBaseDataSourceConfig() string {
	apiName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "api" {
  name          = %[1]q
  protocol_type = "HTTP"
}
resource "aws_apigatewayv2_integration" "test1" {
  api_id             = aws_apigatewayv2_api.api.id
  integration_type   = "HTTP_PROXY"
  integration_method = "ANY"
  integration_uri    = "https://example.com"
}
resource "aws_apigatewayv2_integration" "test2" {
  api_id             = aws_apigatewayv2_api.api.id
  integration_type   = "HTTP_PROXY"
  integration_method = "ANY"
  integration_uri    = "https://example.com"
}
`, apiName)
}

func testAccAPIIntegrationsDataSourceConfig_id() string {
	return acctest.ConfigCompose(
		testAccAPIIntegrationsBaseDataSourceConfig(),
		`
data "aws_apigatewayv2_integrations" "test1" {
  api_id = aws_apigatewayv2_api.api.id
  depends_on = [
    aws_apigatewayv2_integration.test1,
    aws_apigatewayv2_integration.test2,
  ]
}
`)
}

func testAccAPIIntegrationsNoIntegrationsDataSourceConfig() string {
	apiName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "api" {
  name          = %[1]q
  protocol_type = "HTTP"
}
`, apiName)
}

func testAccAPIIntegrationsDataSourceConfig_no_ids() string {
	return acctest.ConfigCompose(
		testAccAPIIntegrationsNoIntegrationsDataSourceConfig(),
		`
data "aws_apigatewayv2_integrations" "test1" {
  api_id = aws_apigatewayv2_api.api.id
}
`)
}
