// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ce_test

import (
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/costexplorer/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccCECostCategoryDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var output awstypes.CostCategory
	resourceName := "aws_ce_cost_category.test"
	dataSourceName := "data.aws_ce_cost_category.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckPayerAccount(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ErrorCheck:               acctest.ErrorCheck(t, names.CEServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccCostCategoryDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCostCategoryExists(ctx, t, resourceName, &output),
					resource.TestCheckResourceAttrPair(dataSourceName, "cost_category_arn", resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDefaultValue, resourceName, names.AttrDefaultValue),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "rule_version", resourceName, "rule_version"),
					resource.TestCheckResourceAttrPair(dataSourceName, "rule.%", resourceName, "rule.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
				),
			},
		},
	})
}

func testAccCostCategoryDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccCostCategoryConfig_basic(rName), `
data "aws_ce_cost_category" "test" {
  cost_category_arn = aws_ce_cost_category.test.arn
}
`)
}
