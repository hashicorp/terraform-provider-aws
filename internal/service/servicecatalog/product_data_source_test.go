// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package servicecatalog_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccServiceCatalogProductDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_servicecatalog_product.test"
	dataSourceName := "data.aws_servicecatalog_product.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domain := fmt.Sprintf("http://%s", acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProductDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProductDataSourceConfig_basic(rName, "beskrivning", "supportbeskrivning", domain, acctest.DefaultEmailAddress),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrCreatedTime, dataSourceName, names.AttrCreatedTime),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrDescription, dataSourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(resourceName, "distributor", dataSourceName, "distributor"),
					resource.TestCheckResourceAttrPair(resourceName, "has_default_path", dataSourceName, "has_default_path"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrName, dataSourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrOwner, dataSourceName, names.AttrOwner),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrStatus, dataSourceName, names.AttrStatus),
					resource.TestCheckResourceAttrPair(resourceName, "support_description", dataSourceName, "support_description"),
					resource.TestCheckResourceAttrPair(resourceName, "support_email", dataSourceName, "support_email"),
					resource.TestCheckResourceAttrPair(resourceName, "support_url", dataSourceName, "support_url"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrType, dataSourceName, names.AttrType),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{})),
				},
			},
		},
	})
}

func TestAccServiceCatalogProductDataSource_physicalID(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_servicecatalog_product.test"
	dataSourceName := "data.aws_servicecatalog_product.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	domain := fmt.Sprintf("http://%s", acctest.RandomDomainName())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ServiceCatalogServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckProductDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccProductDataSourceConfig_physicalID(rName, domain, acctest.DefaultEmailAddress),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, names.AttrARN, dataSourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrCreatedTime, dataSourceName, names.AttrCreatedTime),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrDescription, dataSourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(resourceName, "distributor", dataSourceName, "distributor"),
					resource.TestCheckResourceAttrPair(resourceName, "has_default_path", dataSourceName, "has_default_path"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrName, dataSourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrOwner, dataSourceName, names.AttrOwner),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrStatus, dataSourceName, names.AttrStatus),
					resource.TestCheckResourceAttrPair(resourceName, "support_description", dataSourceName, "support_description"),
					resource.TestCheckResourceAttrPair(resourceName, "support_email", dataSourceName, "support_email"),
					resource.TestCheckResourceAttrPair(resourceName, "support_url", dataSourceName, "support_url"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrType, dataSourceName, names.AttrType),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New(names.AttrTags), knownvalue.MapExact(map[string]knownvalue.Check{})),
				},
			},
		},
	})
}

func testAccProductDataSourceConfig_basic(rName, description, supportDescription, domain, email string) string {
	return acctest.ConfigCompose(testAccProductConfig_basic(rName, description, supportDescription, domain, email), `
data "aws_servicecatalog_product" "test" {
  id = aws_servicecatalog_product.test.id
}
`)
}

func testAccProductDataSourceConfig_physicalID(rName, domain, email string) string {
	return acctest.ConfigCompose(testAccProductConfig_physicalID(rName, domain, email), `
data "aws_servicecatalog_product" "test" {
  id = aws_servicecatalog_product.test.id
}
`)
}
