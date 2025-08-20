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

func TestAccAPIGatewayV2AuthorizerDataSource_jwt(t *testing.T) {
	ctx := acctest.Context(t)
	dataSource1Name := "data.aws_apigatewayv2_authorizer.test1"
	resourceName := "aws_apigatewayv2_authorizer.authorizer"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIGatewayAuthorizerDataSourceConfig_jwt(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSource1Name, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "authorizer_type", resourceName, "authorizer_type"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "identity_sources.#", resourceName, "identity_sources.#"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "identity_sources.0", resourceName, "identity_sources.0"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "jwt_configuration.0.audience", resourceName, "jwt_configuration.0.audience"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "jwt_configuration.0.issuer", resourceName, "jwt_configuration.0.issuer"),
				),
			},
		},
	})
}

func TestAccAPIGatewayV2AuthorizerDataSource_request(t *testing.T) {
	ctx := acctest.Context(t)
	dataSource1Name := "data.aws_apigatewayv2_authorizer.test1"
	resourceName := "aws_apigatewayv2_authorizer.authorizer"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIGatewayAuthorizerDataSourceConfig_request(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSource1Name, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "authorizer_type", resourceName, "authorizer_type"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "identity_sources.#", resourceName, "identity_sources.#"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "identity_sources.0", resourceName, "identity_sources.0"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "authorizer_credentials_arn", resourceName, "authorizer_credentials_arn"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "authorizer_payload_format_version", resourceName, "authorizer_payload_format_version"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "authorizer_result_ttl_in_seconds", resourceName, "authorizer_result_ttl_in_seconds"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "authorizer_uri", resourceName, "authorizer_uri"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "enable_simple_responses", resourceName, "enable_simple_responses"),
				),
			},
		},
	})
}

func testAccAPIGatewayAuthorizerDataSourceConfig_jwt() string {
	apiName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "api" {
  name          = %[1]q
  protocol_type = "HTTP"
}

resource "aws_cognito_user_pool" "pool" {
  name = "testpool"
}

resource "aws_apigatewayv2_authorizer" "authorizer" {
  api_id           = aws_apigatewayv2_api.api.id
  authorizer_type  = "JWT"
  identity_sources = ["$request.header.Authorization"]
  name             = "example-authorizer"
  jwt_configuration {
    audience = ["https:/audience"]
    issuer   = "https://${aws_cognito_user_pool.pool.endpoint}"
  }
}

data "aws_apigatewayv2_authorizer" "test1" {
  api_id = aws_apigatewayv2_api.api.id
  authorizer_id = aws_apigatewayv2_authorizer.authorizer.id
}

`, apiName)
}

func testAccAPIGatewayAuthorizerDataSourceConfig_request() string {
	apiName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "api" {
  name          = %[1]q
  protocol_type = "HTTP"
}

resource "aws_cognito_user_pool" "pool" {
  name = "testpool"
}

data "aws_caller_identity" "current" {}

resource "aws_apigatewayv2_authorizer" "authorizer" {
  api_id           = aws_apigatewayv2_api.api.id
  authorizer_type  = "REQUEST"
  authorizer_credentials_arn = "arn:aws:iam::${data.aws_caller_identity.current.account_id}:role/apigAwsProxyRole"
  authorizer_payload_format_version = "2.0"
  authorizer_result_ttl_in_seconds = 300
  authorizer_uri = "arn:aws:apigateway:us-east-1:lambda:path/2015-03-31/functions/arn:aws:lambda:us-east-1:123456789012:function:HelloWorld/invocations"
  enable_simple_responses = true
  identity_sources = ["$request.header.Authorization"]
  name             = "example-authorizer"
}

data "aws_apigatewayv2_authorizer" "test1" {
  api_id = aws_apigatewayv2_api.api.id
  authorizer_id = aws_apigatewayv2_authorizer.authorizer.id
}

`, apiName)
}
