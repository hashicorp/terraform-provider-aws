// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalog_test

import (
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccServiceCatalogPortfolioConstraintsDataSource_Constraint_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_servicecatalog_constraint.test"
	dataSourceName := "data.aws_servicecatalog_portfolio_constraints.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPortfolioConstraintsDataSourceConfig_constraintBasic(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "accept_language", resourceName, "accept_language"),
					resource.TestCheckResourceAttr(dataSourceName, "details.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(dataSourceName, "details.0.constraint_id", resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, "details.0.description", resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(dataSourceName, "details.0.owner", resourceName, names.AttrOwner),
					resource.TestCheckResourceAttrPair(dataSourceName, "details.0.portfolio_id", resourceName, "portfolio_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "details.0.product_id", resourceName, "product_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "details.0.type", resourceName, names.AttrType),
					resource.TestCheckResourceAttrPair(dataSourceName, "portfolio_id", resourceName, "portfolio_id"),
				),
			},
		},
	})
}

func testAccPortfolioConstraintsDataSourceConfig_constraintBasic(rName, description string) string {
	return acctest.ConfigCompose(testAccConstraintConfig_basic(rName, description), `
data "aws_servicecatalog_portfolio_constraints" "test" {
  portfolio_id = aws_servicecatalog_constraint.test.portfolio_id
}
`)
}
