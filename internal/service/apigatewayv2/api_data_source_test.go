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

func TestAccAPIGatewayV2APIDataSource_http(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_apigatewayv2_api.test"
	resourceName := "aws_apigatewayv2_api.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIDataSourceConfig_http(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "api_endpoint", resourceName, "api_endpoint"),
					resource.TestCheckResourceAttrPair(dataSourceName, "api_key_selection_expression", resourceName, "api_key_selection_expression"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "cors_configuration.#", resourceName, "cors_configuration.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cors_configuration.0.allow_credentials", resourceName, "cors_configuration.0.allow_credentials"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cors_configuration.0.allow_headers.#", resourceName, "cors_configuration.0.allow_headers.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cors_configuration.0.allow_methods.#", resourceName, "cors_configuration.0.allow_methods.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cors_configuration.0.allow_origins.#", resourceName, "cors_configuration.0.allow_origins.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cors_configuration.0.expose_headers.#", resourceName, "cors_configuration.0.expose_headers.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cors_configuration.0.max_age", resourceName, "cors_configuration.0.max_age"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(dataSourceName, "disable_execute_api_endpoint", resourceName, "disable_execute_api_endpoint"),
					resource.TestCheckResourceAttrPair(dataSourceName, "execution_arn", resourceName, "execution_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "protocol_type", resourceName, "protocol_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "route_selection_expression", resourceName, "route_selection_expression"),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.Key1", resourceName, "tags.Key1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.Key2", resourceName, "tags.Key2"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrVersion, resourceName, names.AttrVersion),
				),
			},
		},
	})
}

func TestAccAPIGatewayV2APIDataSource_webSocket(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_apigatewayv2_api.test"
	resourceName := "aws_apigatewayv2_api.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIDataSourceConfig_webSocket(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "api_endpoint", resourceName, "api_endpoint"),
					resource.TestCheckResourceAttrPair(dataSourceName, "api_key_selection_expression", resourceName, "api_key_selection_expression"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "cors_configuration.#", resourceName, "cors_configuration.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(dataSourceName, "disable_execute_api_endpoint", resourceName, "disable_execute_api_endpoint"),
					resource.TestCheckResourceAttrPair(dataSourceName, "execution_arn", resourceName, "execution_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "protocol_type", resourceName, "protocol_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "route_selection_expression", resourceName, "route_selection_expression"),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.Key1", resourceName, "tags.Key1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.Key2", resourceName, "tags.Key2"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrVersion, resourceName, names.AttrVersion),
				),
			},
		},
	})
}

func testAccAPIDataSourceConfig_http(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  description   = "test description"
  name          = %[1]q
  protocol_type = "HTTP"
  version       = "v1"

  cors_configuration {
    allow_headers = ["Authorization"]
    allow_methods = ["GET", "put"]
    allow_origins = ["https://www.example.com"]
  }

  tags = {
    Key1 = "Value1h"
    Key2 = "Value2h"
  }
}

data "aws_apigatewayv2_api" "test" {
  api_id = aws_apigatewayv2_api.test.id
}
`, rName)
}

func testAccAPIDataSourceConfig_webSocket(rName string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test" {
  api_key_selection_expression = "$context.authorizer.usageIdentifierKey"
  description                  = "test description"
  name                         = %[1]q
  protocol_type                = "WEBSOCKET"
  route_selection_expression   = "$request.body.service"
  version                      = "v1"

  tags = {
    Key1 = "Value1ws"
    Key2 = "Value2ws"
  }
}

data "aws_apigatewayv2_api" "test" {
  api_id = aws_apigatewayv2_api.test.id
}
`, rName)
}
