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

func TestAccAPIGatewayV2AuthorizerDataSource_jwt(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_apigatewayv2_authorizer.test"
	resourceName := "aws_apigatewayv2_authorizer.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIGatewayAuthorizerDataSourceConfig_jwt(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckNoResourceAttr(dataSourceName, "authorizer_credentials_arn"),
					resource.TestCheckNoResourceAttr(dataSourceName, "authorizer_payload_format_version"),
					resource.TestCheckNoResourceAttr(dataSourceName, "authorizer_result_ttl_in_seconds"),
					resource.TestCheckResourceAttrPair(dataSourceName, "authorizer_type", resourceName, "authorizer_type"),
					resource.TestCheckNoResourceAttr(dataSourceName, "authorizer_uri"),
					resource.TestCheckNoResourceAttr(dataSourceName, "enable_simple_responses"),
					resource.TestCheckResourceAttrPair(dataSourceName, "identity_sources.#", resourceName, "identity_sources.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "identity_sources.0", resourceName, "identity_sources.0"),
					resource.TestCheckResourceAttrPair(dataSourceName, "jwt_configuration.#", resourceName, "jwt_configuration.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "jwt_configuration.0.audience", resourceName, "jwt_configuration.0.audience"),
					resource.TestCheckResourceAttrPair(dataSourceName, "jwt_configuration.0.issuer", resourceName, "jwt_configuration.0.issuer"),
				),
			},
		},
	})
}

func TestAccAPIGatewayV2AuthorizerDataSource_request(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_apigatewayv2_authorizer.test"
	resourceName := "aws_apigatewayv2_authorizer.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIGatewayAuthorizerDataSourceConfig_request(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckNoResourceAttr(dataSourceName, "authorizer_credentials_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "authorizer_payload_format_version", resourceName, "authorizer_payload_format_version"),
					resource.TestCheckResourceAttrPair(dataSourceName, "authorizer_result_ttl_in_seconds", resourceName, "authorizer_result_ttl_in_seconds"),
					resource.TestCheckResourceAttrPair(dataSourceName, "authorizer_type", resourceName, "authorizer_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "authorizer_uri", resourceName, "authorizer_uri"),
					resource.TestCheckResourceAttrPair(dataSourceName, "identity_sources.#", resourceName, "identity_sources.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "enable_simple_responses", resourceName, "enable_simple_responses"),
					resource.TestCheckResourceAttrPair(dataSourceName, "jwt_configuration.#", resourceName, "jwt_configuration.#"),
				),
			},
		},
	})
}

func TestAccAPIGatewayV2AuthorizerDataSource_credentials(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_apigatewayv2_authorizer.test"
	resourceName := "aws_apigatewayv2_authorizer.test"
	iamRoleResourceName := "aws_iam_role.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIGatewayAuthorizerDataSourceConfig_credentials(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "authorizer_credentials_arn", iamRoleResourceName, names.AttrARN),
					resource.TestCheckNoResourceAttr(dataSourceName, "authorizer_payload_format_version"),
					resource.TestCheckNoResourceAttr(dataSourceName, "authorizer_result_ttl_in_seconds"),
					resource.TestCheckResourceAttrPair(dataSourceName, "authorizer_type", resourceName, "authorizer_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "authorizer_uri", resourceName, "authorizer_uri"),
					resource.TestCheckResourceAttrPair(dataSourceName, "identity_sources.#", resourceName, "identity_sources.#"),
					resource.TestCheckNoResourceAttr(dataSourceName, "enable_simple_responses"),
					resource.TestCheckResourceAttrPair(dataSourceName, "jwt_configuration.#", resourceName, "jwt_configuration.#"),
				),
			},
		},
	})
}

func testAccAPIGatewayAuthorizerDataSourceConfig_jwt(rName string) string {
	return acctest.ConfigCompose(testAccAuthorizerConfig_jwt(rName), `
data "aws_apigatewayv2_authorizer" "test" {
  api_id        = aws_apigatewayv2_api.test.id
  authorizer_id = aws_apigatewayv2_authorizer.test.id
}

`)
}

func testAccAPIGatewayAuthorizerDataSourceConfig_request(rName string) string {
	return acctest.ConfigCompose(testAccAuthorizerConfig_httpAPILambdaRequest(rName), `
data "aws_apigatewayv2_authorizer" "test" {
  api_id = aws_apigatewayv2_api.test.id
  authorizer_id = aws_apigatewayv2_authorizer.test.id
}

`)
}

func testAccAPIGatewayAuthorizerDataSourceConfig_credentials(rName string) string {
	return acctest.ConfigCompose(testAccAuthorizerConfig_credentials(rName), `
data "aws_apigatewayv2_authorizer" "test" {
  api_id = aws_apigatewayv2_api.test.id
  authorizer_id = aws_apigatewayv2_authorizer.test.id
}

`)
}
