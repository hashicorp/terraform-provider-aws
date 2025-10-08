// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package datazone_test

import (
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccDataZoneDomainDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	dataSourceName := "data.aws_datazone_domain.test"
	resourceName := "aws_datazone_domain.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
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
				),
			},
		},
	})
}

func TestAccDataZoneDomainDataSource_name(t *testing.T) {
	ctx := acctest.Context(t)

	dataSourceName := "data.aws_datazone_domain.test"
	resourceName := "aws_datazone_domain.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
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
