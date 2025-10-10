// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package location_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLocationRouteCalculatorDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_location_route_calculator.test"
	resourceName := "aws_location_route_calculator.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LocationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckRouteCalculatorDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccRouteCalculatorDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteCalculatorExists(ctx, dataSourceName),
					resource.TestCheckResourceAttrPair(dataSourceName, "calculator_arn", resourceName, "calculator_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "calculator_name", resourceName, "calculator_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrCreateTime, resourceName, names.AttrCreateTime),
					resource.TestCheckResourceAttrPair(dataSourceName, "data_source", resourceName, "data_source"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(dataSourceName, "update_time", resourceName, "update_time"),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
				),
			},
		},
	})
}

func testAccRouteCalculatorDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_location_route_calculator" "test" {
  calculator_name = %[1]q
  data_source     = "Here"
}

data "aws_location_route_calculator" "test" {
  calculator_name = aws_location_route_calculator.test.calculator_name
}
`, rName)
}
