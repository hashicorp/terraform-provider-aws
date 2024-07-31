// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lexmodels_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
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
		ErrorCheck:               acctest.ErrorCheck(t, names.LexModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccSlotTypeConfig_basic(rName),
					testAccSlotTypeDataSourceConfig_basic(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "checksum", resourceName, "checksum"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrCreatedDate, resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(dataSourceName, "enumeration_value.#", resourceName, "enumeration_value.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrLastUpdatedDate, resourceName, names.AttrLastUpdatedDate),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "value_selection_strategy", resourceName, "value_selection_strategy"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrVersion, resourceName, names.AttrVersion),
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
		ErrorCheck:               acctest.ErrorCheck(t, names.LexModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccSlotTypeConfig_withVersion(rName),
					testAccSlotTypeDataSourceConfig_withVersion(),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "checksum", resourceName, "checksum"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrCreatedDate, resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrLastUpdatedDate, resourceName, names.AttrLastUpdatedDate),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "value_selection_strategy", resourceName, "value_selection_strategy"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrVersion, resourceName, names.AttrVersion),
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
