// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigatewayv2_test

import (
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAPIGatewayV2IntegrationsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSource1Name := "data.aws_apigatewayv2_integrations.test"
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIIntegrationsDataSourceConfig_basic(rName1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSource1Name, "ids.#", "3"),
				),
			},
		},
	})
}

func testAccAPIIntegrationsBaseDataSourceConfig_moreIntegrations() string {
	return `
resource "aws_apigatewayv2_integration" "test2" {
  api_id           = aws_apigatewayv2_api.test.id
  integration_type = "MOCK"
}

resource "aws_apigatewayv2_integration" "test3" {
  api_id           = aws_apigatewayv2_api.test.id
  integration_type = "MOCK"
}

`
}

func testAccAPIIntegrationsDataSourceConfig_basic(rName1 string) string {
	return acctest.ConfigCompose(
		testAccIntegrationConfig_basic(rName1),
		testAccAPIIntegrationsBaseDataSourceConfig_moreIntegrations(),
		`
data "aws_apigatewayv2_integrations" "test" {
  api_id = aws_apigatewayv2_api.test.id

  depends_on = [
    aws_apigatewayv2_integration.test,
    aws_apigatewayv2_integration.test2,
	aws_apigatewayv2_integration.test3,
  ]
}

`)
}
