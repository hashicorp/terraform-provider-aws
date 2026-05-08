// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package servicecatalog_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccServiceCatalogPortfolioDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_servicecatalog_portfolio.test"
	resourceName := "aws_servicecatalog_portfolio.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckPortfolioDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccPortfolioDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrCreatedTime, dataSourceName, names.AttrCreatedTime),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrDescription, dataSourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrName, dataSourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrProviderName, dataSourceName, names.AttrProviderName),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{})),
				},
			},
		},
	})
}

func testAccPortfolioDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccPortfolioConfig_basic(rName), `
data "aws_servicecatalog_portfolio" "test" {
  id = aws_servicecatalog_portfolio.test.id
}
`)
}
