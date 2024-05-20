// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway_test

import (
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAPIGatewaySDKDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_api_gateway_sdk.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSDKDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("aws_api_gateway_stage.test", "rest_api_id", dataSourceName, "rest_api_id"),
					resource.TestCheckResourceAttrPair("aws_api_gateway_stage.test", "stage_name", dataSourceName, "stage_name"),
					resource.TestCheckResourceAttrSet(dataSourceName, "body"),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrContentType),
					resource.TestCheckResourceAttrSet(dataSourceName, "content_disposition"),
				),
			},
		},
	})
}

func testAccSDKDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccStageConfig_base(rName), `
resource "aws_api_gateway_stage" "test" {
  rest_api_id   = aws_api_gateway_rest_api.test.id
  stage_name    = "prod"
  deployment_id = aws_api_gateway_deployment.test.id
}

data "aws_api_gateway_sdk" "test" {
  rest_api_id = aws_api_gateway_stage.test.rest_api_id
  stage_name  = aws_api_gateway_stage.test.stage_name
  sdk_type    = "android"

  parameters = {
    groupId         = "test"
    artifactId      = "test"
    artifactVersion = "test"
    invokerPackage  = "test"
  }
}
`)
}
