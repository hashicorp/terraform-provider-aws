// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/apigateway"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccAPIGatewayRestAPIDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandString(8)
	dataSourceName := "data.aws_api_gateway_rest_api.test"
	resourceName := "aws_api_gateway_rest_api.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "root_resource_id", resourceName, "root_resource_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags", resourceName, "tags"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "policy", resourceName, "policy"),
					resource.TestCheckResourceAttrPair(dataSourceName, "api_key_source", resourceName, "api_key_source"),
					resource.TestCheckResourceAttrPair(dataSourceName, "minimum_compression_size", resourceName, "minimum_compression_size"),
					resource.TestCheckResourceAttrPair(dataSourceName, "binary_media_types", resourceName, "binary_media_types"),
					resource.TestCheckResourceAttrPair(dataSourceName, "endpoint_configuration", resourceName, "endpoint_configuration"),
					resource.TestCheckResourceAttrPair(dataSourceName, "execution_arn", resourceName, "execution_arn"),
				),
			},
		},
	})
}

func TestAccAPIGatewayRestAPIDataSource_Endpoint_vpcEndpointIDs(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandString(8)
	dataSourceName := "data.aws_api_gateway_rest_api.test"
	resourceName := "aws_api_gateway_rest_api.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccRestAPIDataSourceConfig_Endpoint_vpcEndpointIDs(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "root_resource_id", resourceName, "root_resource_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags", resourceName, "tags"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "policy", resourceName, "policy"),
					resource.TestCheckResourceAttrPair(dataSourceName, "api_key_source", resourceName, "api_key_source"),
					resource.TestCheckResourceAttrPair(dataSourceName, "minimum_compression_size", resourceName, "minimum_compression_size"),
					resource.TestCheckResourceAttrPair(dataSourceName, "binary_media_types", resourceName, "binary_media_types"),
					resource.TestCheckResourceAttrPair(dataSourceName, "endpoint_configuration.#", resourceName, "endpoint_configuration.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "endpoint_configuration.0.vpc_endpoint_ids.#", resourceName, "endpoint_configuration.0.vpc_endpoint_ids.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "execution_arn", resourceName, "execution_arn"),
				),
			},
		},
	})
}

func testAccRestAPIDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccRestAPIConfig_name(rName),
		`
data "aws_api_gateway_rest_api" "test" {
  name = aws_api_gateway_rest_api.test.name
}
`,
	)
}

func testAccRestAPIDataSourceConfig_Endpoint_vpcEndpointIDs(rName string) string {
	return acctest.ConfigCompose(
		testAccRestAPIConfig_vpcEndpointIDs1(rName),
		`
data "aws_api_gateway_rest_api" "test" {
  name = aws_api_gateway_rest_api.test.name
}
`,
	)
}
