// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigatewayv2_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccAPIGatewayV2VPCLinkDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_apigatewayv2_vpc_link.test"
	resourceName := "aws_apigatewayv2_vpc_link.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, apigatewayv2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,

		Steps: []resource.TestStep{
			{
				Config: testAccVPCLinkDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "security_group_ids.#", resourceName, "security_group_ids.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "subnet_ids.#", resourceName, "subnet_ids.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
				),
			},
		},
	})
}

func testAccVPCLinkDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccVPCLinkConfig_basic(rName), `
data "aws_apigatewayv2_vpc_link" "test" {
  vpc_link_id = aws_apigatewayv2_vpc_link.test.id
}
`)
}
