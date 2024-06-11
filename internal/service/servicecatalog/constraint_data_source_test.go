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

func TestAccServiceCatalogConstraintDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_servicecatalog_constraint.test"
	dataSourceName := "data.aws_servicecatalog_constraint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckConstraintDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccConstraintDataSourceConfig_basic(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckConstraintExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "accept_language", dataSourceName, "accept_language"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrDescription, dataSourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrOwner, dataSourceName, names.AttrOwner),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrParameters, dataSourceName, names.AttrParameters),
					resource.TestCheckResourceAttrPair(resourceName, "portfolio_id", dataSourceName, "portfolio_id"),
					resource.TestCheckResourceAttrPair(resourceName, "product_id", dataSourceName, "product_id"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrStatus, dataSourceName, names.AttrStatus),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrType, dataSourceName, names.AttrType),
				),
			},
		},
	})
}

func testAccConstraintDataSourceConfig_basic(rName, description string) string {
	return acctest.ConfigCompose(testAccConstraintConfig_basic(rName, description), `
data "aws_servicecatalog_constraint" "test" {
  id = aws_servicecatalog_constraint.test.id
}
`)
}
