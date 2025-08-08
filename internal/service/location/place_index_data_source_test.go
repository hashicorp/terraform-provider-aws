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

func TestAccLocationPlaceIndexDataSource_indexName(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_location_place_index.test"
	resourceName := "aws_location_place_index.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.LocationServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPlaceIndexDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccPlaceIndexDataSourceConfig_indexName(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrCreateTime, resourceName, names.AttrCreateTime),
					resource.TestCheckResourceAttrPair(dataSourceName, "data_source", resourceName, "data_source"),
					resource.TestCheckResourceAttrPair(dataSourceName, "data_source_configuration.#", resourceName, "data_source_configuration.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "data_source_configuration.0.intended_use", resourceName, "data_source_configuration.0.intended_use"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(dataSourceName, "index_arn", resourceName, "index_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "index_name", resourceName, "index_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(dataSourceName, "update_time", resourceName, "update_time"),
				),
			},
		},
	})
}

func testAccPlaceIndexDataSourceConfig_indexName(rName string) string {
	return fmt.Sprintf(`
resource "aws_location_place_index" "test" {
  data_source = "Here"
  index_name  = %[1]q
}

data "aws_location_place_index" "test" {
  index_name = aws_location_place_index.test.index_name
}
`, rName)
}
