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

func TestAccAPIGatewayV2StagesDataSource_id(t *testing.T) {
	ctx := acctest.Context(t)
	dataSource1Name := "data.aws_apigatewayv2_stages.test1"
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName3 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIStagesDataSourceConfig_id(rName1, rName2, rName3),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSource1Name, "names.#", "3"),
				),
			},
		},
	})
}

func TestAccAPIGatewayV2StagesDataSource_tags(t *testing.T) {
	ctx := acctest.Context(t)
	dataSource1Name := "data.aws_apigatewayv2_stages.test1"
	dataSource2Name := "data.aws_apigatewayv2_stages.test2"
	dataSource3Name := "data.aws_apigatewayv2_stages.test3"
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName3 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayV2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIStagesDataSourceConfig_tags(rName1, rName2, rName3),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSource1Name, "names.#", "1"),
					resource.TestCheckResourceAttr(dataSource2Name, "names.#", "1"),
					resource.TestCheckResourceAttr(dataSource3Name, "names.#", "1"),
					resource.TestCheckResourceAttr(dataSource1Name, "names.0", rName1),
					resource.TestCheckResourceAttr(dataSource2Name, "names.0", rName2),
					resource.TestCheckResourceAttr(dataSource3Name, "names.0", rName3),
				),
			},
		},
	})
}

func testAccAPIStagesBaseDataSourceConfig(rName1, rName2, rName3 string) string {
	apiName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "api" {
  name          = %[1]q
  protocol_type = "HTTP"
}
  
resource "aws_apigatewayv2_stage" "test1" {
  api_id = aws_apigatewayv2_api.api.id
  name   = %[2]q

  tags = {
    Name = %[2]q
  }
}

resource "aws_apigatewayv2_stage" "test2" {
  api_id = aws_apigatewayv2_api.api.id
  name   = %[3]q

  tags = {
    Name = %[3]q
  }
}

resource "aws_apigatewayv2_stage" "test3" {
  api_id = aws_apigatewayv2_api.api.id
  name   = %[4]q

  tags = {
    Name = %[4]q
  }
}

`, apiName, rName1, rName2, rName3)
}

func testAccAPIStagesDataSourceConfig_id(rName1, rName2, rName3 string) string {
	return acctest.ConfigCompose(
		testAccAPIStagesBaseDataSourceConfig(rName1, rName2, rName3),
		`
data "aws_apigatewayv2_stages" "test1" {
  api_id = aws_apigatewayv2_api.api.id

  depends_on = [
    aws_apigatewayv2_stage.test1,
    aws_apigatewayv2_stage.test2,
    aws_apigatewayv2_stage.test3,
  ]
}

`)
}

func testAccAPIStagesDataSourceConfig_tags(rName1, rName2, rName3 string) string {
	return acctest.ConfigCompose(
		testAccAPIStagesBaseDataSourceConfig(rName1, rName2, rName3),
		`
data "aws_apigatewayv2_stages" "test1" {
  # Force dependency on resources.
  api_id = aws_apigatewayv2_api.api.id
  tags = {
    Name = element([aws_apigatewayv2_stage.test1.name, aws_apigatewayv2_stage.test2.name, aws_apigatewayv2_stage.test3.name], 0)
  }
}

data "aws_apigatewayv2_stages" "test2" {
  # Force dependency on resources.
  api_id = aws_apigatewayv2_api.api.id
  tags = {
    Name = element([aws_apigatewayv2_stage.test1.name, aws_apigatewayv2_stage.test2.name, aws_apigatewayv2_stage.test3.name], 1)
  }
}

data "aws_apigatewayv2_stages" "test3" {
  # Force dependency on resources.
  api_id = aws_apigatewayv2_api.api.id
  tags = {
    Name = element([aws_apigatewayv2_stage.test1.name, aws_apigatewayv2_stage.test2.name, aws_apigatewayv2_stage.test3.name], 2)
  }
}

`)
}
