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

func TestAccAPIGatewayAPIKeysDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_api_key.test"
	dataSourceName := "data.aws_api_gateway_api_keys.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPIKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAPIKeysDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanOrEqualValue(dataSourceName, "items.#", 1),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "items.*.created_date", resourceName, names.AttrCreatedDate),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "items.*.description", resourceName, names.AttrDescription),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "items.*.enabled", resourceName, names.AttrEnabled),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "items.*.id", resourceName, names.AttrID),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "items.*.last_updated_date", resourceName, names.AttrLastUpdatedDate),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "items.*.name", resourceName, names.AttrName),
					resource.TestCheckNoResourceAttr(dataSourceName, "items.*.value"),
				),
			},
		},
	})
}

func TestAccAPIGatewayAPIKeysDataSource_includeValues(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_api_gateway_api_key.test"
	dataSourceName := "data.aws_api_gateway_api_keys.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPIKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAPIKeysDataSourceConfig_includeValues(rName),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanOrEqualValue(dataSourceName, "items.#", 1),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "items.*.id", resourceName, names.AttrID),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "items.*.name", resourceName, names.AttrName),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "items.*.value", resourceName, names.AttrValue),
				),
			},
		},
	})
}

func TestAccAPIGatewayAPIKeysDataSource_manyKeys(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	dataSourceName := "data.aws_api_gateway_api_keys.test"
	keyCount := 3

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.APIGatewayServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAPIKeyDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAPIKeysDataSourceConfig_manyKeys(rName, keyCount),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanOrEqualValue(dataSourceName, "items.#", keyCount),
				),
			},
		},
	})
}

func testAccAPIKeysDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_api_key" "test" {
  name = %[1]q
}

data "aws_api_gateway_api_keys" "test" {
  depends_on = [aws_api_gateway_api_key.test]
}
`, rName)
}

func testAccAPIKeysDataSourceConfig_includeValues(rName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_api_key" "test" {
  name = %[1]q
}

data "aws_api_gateway_api_keys" "test" {
  depends_on = [aws_api_gateway_api_key.test]

  include_values = true
}
`, rName)
}

func testAccAPIKeysDataSourceConfig_manyKeys(rName string, count int) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_api_key" "test" {
  count = %[2]d
  name  = "%[1]s-${count.index}"
}

data "aws_api_gateway_api_keys" "test" {
  depends_on = [aws_api_gateway_api_key.test]

  include_values = true
}
`, rName, count)
}
