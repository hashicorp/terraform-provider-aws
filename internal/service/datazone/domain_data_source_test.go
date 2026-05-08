// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package datazone_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDataZoneDomainDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	dataSourceName := "data.aws_datazone_domain.test"
	resourceName := "aws_datazone_domain.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AttrDomainName),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "root_domain_unit_id", resourceName, "root_domain_unit_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "portal_url", resourceName, "portal_url"),
					resource.TestCheckResourceAttrPair(dataSourceName, "domain_version", resourceName, "domain_version"),
				),
			},
		},
	})
}

func TestAccDataZoneDomainDataSource_name(t *testing.T) {
	ctx := acctest.Context(t)

	dataSourceName := "data.aws_datazone_domain.test"
	resourceName := "aws_datazone_domain.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.AttrDomainName),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDomainDataSourceConfig_name(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "root_domain_unit_id", resourceName, "root_domain_unit_id"),
				),
			},
		},
	})
}

func testAccDomainDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccDomainConfig_basic(rName), `
data "aws_datazone_domain" "test" {
  id = aws_datazone_domain.test.id
}`)
}

func testAccDomainDataSourceConfig_name(rName string) string {
	return acctest.ConfigCompose(testAccDomainConfig_basic(rName), `
data "aws_datazone_domain" "test" {
  name = aws_datazone_domain.test.name
}`)
}
