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

func TestAccAPIGatewayV2APIsDataSource_name(t *testing.T) {
	ctx := acctest.Context(t)
	dataSource1Name := "data.aws_apigatewayv2_apis.test1"
	dataSource2Name := "data.aws_apigatewayv2_apis.test2"
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIsDataSourceConfig_name(rName1, rName2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSource1Name, "ids.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSource2Name, "ids.#", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccAPIGatewayV2APIsDataSource_protocolType(t *testing.T) {
	ctx := acctest.Context(t)
	dataSource1Name := "data.aws_apigatewayv2_apis.test1"
	dataSource2Name := "data.aws_apigatewayv2_apis.test2"
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIsDataSourceConfig_protocolType(rName1, rName2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSource1Name, "ids.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSource2Name, "ids.#", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccAPIGatewayV2APIsDataSource_tags(t *testing.T) {
	ctx := acctest.Context(t)
	dataSource1Name := "data.aws_apigatewayv2_apis.test1"
	dataSource2Name := "data.aws_apigatewayv2_apis.test2"
	dataSource3Name := "data.aws_apigatewayv2_apis.test3"
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIsDataSourceConfig_tags(rName1, rName2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSource1Name, "ids.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSource2Name, "ids.#", acctest.Ct2),
					resource.TestCheckResourceAttr(dataSource3Name, "ids.#", acctest.Ct0),
				),
			},
		},
	})
}

func testAccAPIsBaseDataSourceConfig(rName1, rName2 string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test1" {
  name          = %[1]q
  protocol_type = "HTTP"

  tags = {
    Name = %[1]q
  }
}

resource "aws_apigatewayv2_api" "test2" {
  name          = %[2]q
  protocol_type = "HTTP"

  tags = {
    Name = %[2]q
  }
}

resource "aws_apigatewayv2_api" "test3" {
  name                       = %[2]q
  protocol_type              = "WEBSOCKET"
  route_selection_expression = "$request.body.action"

  tags = {
    Name = %[2]q
  }
}
`, rName1, rName2)
}

func testAccAPIsDataSourceConfig_name(rName1, rName2 string) string {
	return acctest.ConfigCompose(
		testAccAPIsBaseDataSourceConfig(rName1, rName2),
		`
data "aws_apigatewayv2_apis" "test1" {
  # Force dependency on resources.
  name = element([aws_apigatewayv2_api.test1.name, aws_apigatewayv2_api.test2.name, aws_apigatewayv2_api.test3.name], 0)
}

data "aws_apigatewayv2_apis" "test2" {
  # Force dependency on resources.
  name = element([aws_apigatewayv2_api.test1.name, aws_apigatewayv2_api.test2.name, aws_apigatewayv2_api.test3.name], 1)
}
`)
}

func testAccAPIsDataSourceConfig_protocolType(rName1, rName2 string) string {
	return acctest.ConfigCompose(
		testAccAPIsBaseDataSourceConfig(rName1, rName2),
		fmt.Sprintf(`
data "aws_apigatewayv2_apis" "test1" {
  name = %[1]q

  protocol_type = element([aws_apigatewayv2_api.test1.protocol_type, aws_apigatewayv2_api.test2.protocol_type, aws_apigatewayv2_api.test3.protocol_type], 0)
}

data "aws_apigatewayv2_apis" "test2" {
  name = %[2]q

  protocol_type = element([aws_apigatewayv2_api.test1.protocol_type, aws_apigatewayv2_api.test2.protocol_type, aws_apigatewayv2_api.test3.protocol_type], 3)
}
`, rName1, rName2))
}

func testAccAPIsDataSourceConfig_tags(rName1, rName2 string) string {
	return acctest.ConfigCompose(
		testAccAPIsBaseDataSourceConfig(rName1, rName2),
		`
data "aws_apigatewayv2_apis" "test1" {
  # Force dependency on resources.
  tags = {
    Name = element([aws_apigatewayv2_api.test1.name, aws_apigatewayv2_api.test2.name, aws_apigatewayv2_api.test3.name], 0)
  }
}

data "aws_apigatewayv2_apis" "test2" {
  # Force dependency on resources.
  tags = {
    Name = element([aws_apigatewayv2_api.test1.name, aws_apigatewayv2_api.test2.name, aws_apigatewayv2_api.test3.name], 1)
  }
}

data "aws_apigatewayv2_apis" "test3" {
  # Force dependency on resources.
  tags = {
    Name = element([aws_apigatewayv2_api.test1.name, aws_apigatewayv2_api.test2.name, aws_apigatewayv2_api.test3.name], 2)
    Key2 = "Value2"
  }
}
`)
}
