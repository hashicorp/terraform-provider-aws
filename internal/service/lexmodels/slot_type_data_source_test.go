// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lexmodels_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccLexModelsSlotTypeDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)
	dataSourceName := "data.aws_lex_slot_type.test"
	resourceName := "aws_lex_slot_type.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, lexmodelbuildingservice.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, lexmodelbuildingservice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccSlotTypeConfig_basic(rName),
					testAccSlotTypeDataSourceConfig_basic(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "checksum", resourceName, "checksum"),
					resource.TestCheckResourceAttrPair(dataSourceName, "created_date", resourceName, "created_date"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "enumeration_value.#", resourceName, "enumeration_value.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "last_updated_date", resourceName, "last_updated_date"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "value_selection_strategy", resourceName, "value_selection_strategy"),
					resource.TestCheckResourceAttrPair(dataSourceName, "version", resourceName, "version"),
				),
			},
		},
	})
}

func TestAccLexModelsSlotTypeDataSource_withVersion(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)
	dataSourceName := "data.aws_lex_slot_type.test"
	resourceName := "aws_lex_slot_type.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, lexmodelbuildingservice.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, lexmodelbuildingservice.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccSlotTypeConfig_withVersion(rName),
					testAccSlotTypeDataSourceConfig_withVersion(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "checksum", resourceName, "checksum"),
					resource.TestCheckResourceAttrPair(dataSourceName, "created_date", resourceName, "created_date"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "last_updated_date", resourceName, "last_updated_date"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "value_selection_strategy", resourceName, "value_selection_strategy"),
					resource.TestCheckResourceAttrPair(dataSourceName, "version", resourceName, "version"),
				),
			},
		},
	})
}

func testAccSlotTypeDataSourceConfig_basic() string {
	return `
data "aws_lex_slot_type" "test" {
  name = aws_lex_slot_type.test.name
}
`
}

func testAccSlotTypeDataSourceConfig_withVersion() string {
	return `
data "aws_lex_slot_type" "test" {
  name    = aws_lex_slot_type.test.name
  version = "1"
}
`
}
