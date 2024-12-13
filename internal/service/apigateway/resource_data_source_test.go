// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAPIGatewayResourceDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandString(8)
	resourceName1 := "aws_api_gateway_resource.example_v1"
	dataSourceName1 := "data.aws_api_gateway_resource.example_v1"
	resourceName2 := "aws_api_gateway_resource.example_v1_endpoint"
	dataSourceName2 := "data.aws_api_gateway_resource.example_v1_endpoint"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckAPIGatewayTypeEDGE(t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccResourceDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName1, names.AttrID, dataSourceName1, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName1, "parent_id", dataSourceName1, "parent_id"),
					resource.TestCheckResourceAttrPair(resourceName1, "path_part", dataSourceName1, "path_part"),
					resource.TestCheckResourceAttrPair(resourceName2, names.AttrID, dataSourceName2, names.AttrID),
					resource.TestCheckResourceAttrPair(resourceName2, "parent_id", dataSourceName2, "parent_id"),
					resource.TestCheckResourceAttrPair(resourceName2, "path_part", dataSourceName2, "path_part"),
				),
			},
		},
	})
}

func testAccResourceDataSourceConfig_basic(r string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "example" {
  name = "%s_example"
}

resource "aws_api_gateway_resource" "example_v1" {
  rest_api_id = aws_api_gateway_rest_api.example.id
  parent_id   = aws_api_gateway_rest_api.example.root_resource_id
  path_part   = "v1"
}

resource "aws_api_gateway_resource" "example_v1_endpoint" {
  rest_api_id = aws_api_gateway_rest_api.example.id
  parent_id   = aws_api_gateway_resource.example_v1.id
  path_part   = "endpoint"
}

data "aws_api_gateway_resource" "example_v1" {
  rest_api_id = aws_api_gateway_rest_api.example.id
  path        = "/${aws_api_gateway_resource.example_v1.path_part}"
}

data "aws_api_gateway_resource" "example_v1_endpoint" {
  rest_api_id = aws_api_gateway_rest_api.example.id
  path        = "/${aws_api_gateway_resource.example_v1.path_part}/${aws_api_gateway_resource.example_v1_endpoint.path_part}"
}
`, r)
}
