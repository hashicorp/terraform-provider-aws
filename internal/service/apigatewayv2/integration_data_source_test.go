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

func TestAccAPIGatewayV2IntegrationDataSource_basicWebsocket(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_apigatewayv2_integration.test"
	resourceName := "aws_apigatewayv2_integration.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccIntegrationsDataSourceConfig_basicWebsocket(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "api_id", resourceName, "api_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "integration_id", resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckNoResourceAttr(dataSourceName, names.AttrConnectionID),
					resource.TestCheckResourceAttrPair(dataSourceName, "connection_type", resourceName, "connection_type"),
					resource.TestCheckNoResourceAttr(dataSourceName, "content_handling_strategy"),
					resource.TestCheckNoResourceAttr(dataSourceName, "credentials_arn"),
					resource.TestCheckNoResourceAttr(dataSourceName, names.AttrDescription),
					resource.TestCheckNoResourceAttr(dataSourceName, "integration_method"),
					resource.TestCheckResourceAttrPair(dataSourceName, "integration_response_selection_expression", resourceName, "integration_response_selection_expression"),
					resource.TestCheckNoResourceAttr(dataSourceName, "integration_subtype"),
					resource.TestCheckResourceAttrPair(dataSourceName, "integration_type", resourceName, "integration_type"),
					resource.TestCheckNoResourceAttr(dataSourceName, "integration_uri"),
					resource.TestCheckResourceAttrPair(dataSourceName, "passthrough_behavior", resourceName, "passthrough_behavior"),
					resource.TestCheckResourceAttrPair(dataSourceName, "payload_format_version", resourceName, "payload_format_version"),
					resource.TestCheckResourceAttrPair(dataSourceName, "request_parameters.%", resourceName, "request_parameters.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "request_templates.%", resourceName, "request_templates.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "response_parameters.#", resourceName, "response_parameters.#"),
					resource.TestCheckNoResourceAttr(dataSourceName, "template_selection_expression"),
					resource.TestCheckResourceAttrPair(dataSourceName, "timeout_milliseconds", resourceName, "timeout_milliseconds"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tls_config.#", resourceName, "tls_config.#"),
				),
			},
		},
	})
}

func TestAccAPIGatewayV2IntegrationDataSource_basicHTTP(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_apigatewayv2_integration.test"
	resourceName := "aws_apigatewayv2_integration.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccIntegrationsDataSourceConfig_basicHTTP(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckNoResourceAttr(dataSourceName, names.AttrConnectionID),
					resource.TestCheckResourceAttrPair(dataSourceName, "connection_type", resourceName, "connection_type"),
					resource.TestCheckNoResourceAttr(dataSourceName, "content_handling_strategy"),
					resource.TestCheckNoResourceAttr(dataSourceName, "credentials_arn"),
					resource.TestCheckNoResourceAttr(dataSourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(dataSourceName, "integration_method", resourceName, "integration_method"),
					resource.TestCheckNoResourceAttr(dataSourceName, "integration_response_selection_expression"),
					resource.TestCheckNoResourceAttr(dataSourceName, "integration_subtype"),
					resource.TestCheckResourceAttrPair(dataSourceName, "integration_type", resourceName, "integration_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "integration_uri", resourceName, "integration_uri"),
					resource.TestCheckNoResourceAttr(dataSourceName, "passthrough_behavior"),
					resource.TestCheckResourceAttrPair(dataSourceName, "payload_format_version", resourceName, "payload_format_version"),
					resource.TestCheckResourceAttrPair(dataSourceName, "request_parameters.%", resourceName, "request_parameters.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "request_parameters.append:header.header1", resourceName, "request_parameters.append:header.header1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "request_parameters.remove:querystring.qs1", resourceName, "request_parameters.remove:querystring.qs1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "request_templates.%", resourceName, "request_templates.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "response_parameters.#", resourceName, "response_parameters.#"),
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "response_parameters.*", map[string]string{
						names.AttrStatusCode:             "500",
						"mappings.%":                     "2",
						"mappings.append:header.header1": "$context.requestId",
						"mappings.overwrite:statuscode":  "403",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(dataSourceName, "response_parameters.*", map[string]string{
						names.AttrStatusCode:           "404",
						"mappings.%":                   "1",
						"mappings.append:header.error": "$stageVariables.environmentId",
					}),
					resource.TestCheckNoResourceAttr(dataSourceName, "template_selection_expression"),
					resource.TestCheckResourceAttr(resourceName, "timeout_milliseconds", "30000"),
					resource.TestCheckResourceAttr(resourceName, "tls_config.#", "0"),
				),
			},
		},
	})
}

func TestAccAPIGatewayV2IntegrationDataSource_integrationTypeHTTP(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_apigatewayv2_integration.test"
	resourceName := "aws_apigatewayv2_integration.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccIntegrationsDataSourceConfig_integrationTypeHTTP(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(dataSourceName, names.AttrConnectionID),
					resource.TestCheckResourceAttrPair(dataSourceName, "connection_type", resourceName, "connection_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "content_handling_strategy", resourceName, "content_handling_strategy"),
					resource.TestCheckNoResourceAttr(dataSourceName, "credentials_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(dataSourceName, "integration_method", resourceName, "integration_method"),
					resource.TestCheckResourceAttrPair(dataSourceName, "integration_response_selection_expression", resourceName, "integration_response_selection_expression"),
					resource.TestCheckNoResourceAttr(dataSourceName, "integration_subtype"),
					resource.TestCheckResourceAttrPair(dataSourceName, "integration_type", resourceName, "integration_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "integration_uri", resourceName, "integration_uri"),
					resource.TestCheckResourceAttrPair(dataSourceName, "passthrough_behavior", resourceName, "passthrough_behavior"),
					resource.TestCheckResourceAttrPair(dataSourceName, "payload_format_version", resourceName, "payload_format_version"),
					resource.TestCheckResourceAttrPair(dataSourceName, "request_parameters.%", resourceName, "request_parameters.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "request_parameters.integration.request.querystring.stage", resourceName, "request_parameters.integration.request.querystring.stage"),
					resource.TestCheckResourceAttrPair(dataSourceName, "request_templates.%", resourceName, "request_templates.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "response_parameters.#", resourceName, "response_parameters.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "template_selection_expression", resourceName, "template_selection_expression"),
					resource.TestCheckResourceAttrPair(dataSourceName, "timeout_milliseconds", resourceName, "timeout_milliseconds"),
					resource.TestCheckResourceAttr(resourceName, "tls_config.#", "0"),
				),
			},
		},
	})
}

func TestAccAPIGatewayV2IntegrationDataSource_serviceIntegration(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_apigatewayv2_integration.test"
	resourceName := "aws_apigatewayv2_integration.test"
	iamRoleResourceName := "aws_iam_role.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccIntegrationsDataSourceConfig_serviceIntegration(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckNoResourceAttr(dataSourceName, names.AttrConnectionID),
					resource.TestCheckResourceAttrPair(dataSourceName, "connection_type", resourceName, "connection_type"),
					resource.TestCheckNoResourceAttr(dataSourceName, "content_handling_strategy"),
					resource.TestCheckResourceAttrPair(dataSourceName, "credentials_arn", iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckNoResourceAttr(dataSourceName, "integration_method"),
					resource.TestCheckNoResourceAttr(dataSourceName, "integration_response_selection_expression"),
					resource.TestCheckResourceAttrPair(dataSourceName, "integration_subtype", resourceName, "integration_subtype"),
					resource.TestCheckResourceAttrPair(dataSourceName, "integration_type", resourceName, "integration_type"),
					resource.TestCheckNoResourceAttr(dataSourceName, "integration_uri"),
					resource.TestCheckNoResourceAttr(dataSourceName, "passthrough_behavior"),
					resource.TestCheckResourceAttrPair(dataSourceName, "payload_format_version", resourceName, "payload_format_version"),
					resource.TestCheckResourceAttrPair(dataSourceName, "request_parameters.%", resourceName, "request_parameters.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "request_templates.%", resourceName, "request_templates.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "response_parameters.#", resourceName, "response_parameters.#"),
					resource.TestCheckNoResourceAttr(dataSourceName, "template_selection_expression"),
					resource.TestCheckResourceAttrPair(dataSourceName, "timeout_milliseconds", resourceName, "timeout_milliseconds"),
					resource.TestCheckResourceAttr(resourceName, "tls_config.#", "0"),
				),
			},
		},
	})
}

func TestAccAPIGatewayV2IntegrationDataSource_tlsConfig(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_apigatewayv2_integration.test"
	resourceName := "aws_apigatewayv2_integration.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccIntegrationsDataSourceConfig_tlsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrConnectionID, resourceName, names.AttrConnectionID),
					resource.TestCheckResourceAttrPair(dataSourceName, "connection_type", resourceName, "connection_type"),
					resource.TestCheckNoResourceAttr(dataSourceName, "content_handling_strategy"),
					resource.TestCheckNoResourceAttr(dataSourceName, "credentials_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(dataSourceName, "integration_method", resourceName, "integration_method"),
					resource.TestCheckNoResourceAttr(dataSourceName, "integration_response_selection_expression"),
					resource.TestCheckNoResourceAttr(dataSourceName, "integration_subtype"),
					resource.TestCheckResourceAttrPair(dataSourceName, "integration_type", resourceName, "integration_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "integration_uri", resourceName, "integration_uri"),
					resource.TestCheckNoResourceAttr(dataSourceName, "passthrough_behavior"),
					resource.TestCheckResourceAttrPair(dataSourceName, "payload_format_version", resourceName, "payload_format_version"),
					resource.TestCheckResourceAttrPair(dataSourceName, "request_parameters.%", resourceName, "request_parameters.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "request_templates.%", resourceName, "request_templates.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "response_parameters.#", resourceName, "response_parameters.#"),
					resource.TestCheckNoResourceAttr(dataSourceName, "template_selection_expression"),
					resource.TestCheckResourceAttrPair(dataSourceName, "timeout_milliseconds", resourceName, "timeout_milliseconds"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tls_config.#", resourceName, "tls_config.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tls_config.0.server_name_to_verify", resourceName, "tls_config.0.server_name_to_verify"),
				),
			},
		},
	})
}

func testAccIntegrationsDataSourceConfig_basicWebsocket(rName string) string {
	return acctest.ConfigCompose(
		testAccIntegrationConfig_basic(rName),
		`
data "aws_apigatewayv2_integration" "test" {
  api_id         = aws_apigatewayv2_api.test.id
  integration_id = aws_apigatewayv2_integration.test.id
  depends_on = [
    aws_apigatewayv2_integration.test,
  ]
}
`)
}

func testAccIntegrationsDataSourceConfig_basicHTTP(rName string) string {
	return acctest.ConfigCompose(
		testAccIntegrationConfig_dataMappingHTTP(rName),
		`
data "aws_apigatewayv2_integration" "test" {
  api_id         = aws_apigatewayv2_api.test.id
  integration_id = aws_apigatewayv2_integration.test.id
  depends_on = [
    aws_apigatewayv2_integration.test,
  ]
}
`)
}

func testAccIntegrationsDataSourceConfig_integrationTypeHTTP(rName string) string {
	return acctest.ConfigCompose(
		testAccIntegrationConfig_typeHTTP(rName),
		`
data "aws_apigatewayv2_integration" "test" {
  api_id         = aws_apigatewayv2_api.test.id
  integration_id = aws_apigatewayv2_integration.test.id
  depends_on = [
    aws_apigatewayv2_integration.test,
  ]
}
`)
}

func testAccIntegrationsDataSourceConfig_serviceIntegration(rName string, queueIndex int) string {
	return acctest.ConfigCompose(
		testAccIntegrationConfig_sqs(rName, queueIndex),
		`
data "aws_apigatewayv2_integration" "test" {
  api_id         = aws_apigatewayv2_api.test.id
  integration_id = aws_apigatewayv2_integration.test.id
  depends_on = [
    aws_apigatewayv2_integration.test,
  ]
}
`)
}

func testAccIntegrationsDataSourceConfig_tlsConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccIntegrationConfig_vpcLinkHTTP(rName),
		`
data "aws_apigatewayv2_integration" "test" {
  api_id         = aws_apigatewayv2_api.test.id
  integration_id = aws_apigatewayv2_integration.test.id
  depends_on = [
    aws_apigatewayv2_integration.test,
  ]
}
`)
}
