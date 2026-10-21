// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIPAMPoolsDataSource_basic(t *testing.T) { // nosemgrep:ci.vpc-in-test-name
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_vpc_ipam_pools.test"
	dataSourceNameTwo := "data.aws_vpc_ipam_pools.testtwo"
	resourceName := "aws_vpc_ipam_pool.testthree"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMPoolsDataSourceConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "ipam_pools.#", 0),
				),
			},
			{
				Config: testAccIPAMPoolsDataSourceConfig_basicTwoPools,
				Check: resource.ComposeAggregateTestCheckFunc(
					// DS 1 finds all 3 pools
					acctest.CheckResourceAttrGreaterThanValue(dataSourceName, "ipam_pools.#", 2),

					// DS 2 filters on 1 specific pool to validate attributes
					resource.TestCheckResourceAttr(dataSourceNameTwo, "ipam_pools.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceNameTwo, "ipam_pools.0.address_family", resourceName, "address_family"),
					resource.TestCheckResourceAttrPair(dataSourceNameTwo, "ipam_pools.0.allocation_default_netmask_length", resourceName, "allocation_default_netmask_length"),
					resource.TestCheckResourceAttrPair(dataSourceNameTwo, "ipam_pools.0.allocation_max_netmask_length", resourceName, "allocation_max_netmask_length"),
					resource.TestCheckResourceAttrPair(dataSourceNameTwo, "ipam_pools.0.allocation_min_netmask_length", resourceName, "allocation_min_netmask_length"),
					resource.TestCheckResourceAttrPair(dataSourceNameTwo, "ipam_pools.0.allocation_resource_tags.%", resourceName, "allocation_resource_tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceNameTwo, "ipam_pools.0.arn", resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceNameTwo, "ipam_pools.0.auto_import", resourceName, "auto_import"),
					resource.TestCheckResourceAttrPair(dataSourceNameTwo, "ipam_pools.0.description", resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(dataSourceNameTwo, "ipam_pools.0.aws_service", resourceName, "aws_service"),
					resource.TestCheckResourceAttrPair(dataSourceNameTwo, "ipam_pools.0.id", resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceNameTwo, "ipam_pools.0.ipam_scope_id", resourceName, "ipam_scope_id"),
					resource.TestCheckResourceAttrPair(dataSourceNameTwo, "ipam_pools.0.ipam_scope_type", resourceName, "ipam_scope_type"),
					resource.TestCheckResourceAttrPair(dataSourceNameTwo, "ipam_pools.0.locale", resourceName, "locale"),
					resource.TestCheckResourceAttrPair(dataSourceNameTwo, "ipam_pools.0.pool_depth", resourceName, "pool_depth"),
					resource.TestCheckResourceAttrPair(dataSourceNameTwo, "ipam_pools.0.publicly_advertisable", resourceName, "publicly_advertisable"),
					resource.TestCheckResourceAttrPair(dataSourceNameTwo, "ipam_pools.0.source_ipam_pool_id", resourceName, "source_ipam_pool_id"),
					resource.TestCheckResourceAttr(dataSourceNameTwo, "ipam_pools.0.tags.tagtest", "3"),
				),
			},
		},
	})
}

// https://github.com/hashicorp/terraform-provider-aws/issues/40314.
func TestAccIPAMPoolsDataSource_identicalTags(t *testing.T) { // nosemgrep:ci.vpc-in-test-name
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_vpc_ipam_pools.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckIPAMPoolDestroy(ctx, t),
			testAccCheckIPAMDestroy(ctx, t),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMPoolsDataSourceConfig_identicalTags(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "ipam_pools.#", "2"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "ipam_pools.*.id", "aws_vpc_ipam_pool.test1", names.AttrID),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "ipam_pools.*.id", "aws_vpc_ipam_pool.test2", names.AttrID),
				),
			},
		},
	})
}

func TestAccIPAMPoolsDataSource_empty(t *testing.T) { // nosemgrep:ci.vpc-in-test-name
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_vpc_ipam_pools.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccIPAMPoolsDataSourceConfig_empty,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "ipam_pools.#", "0"),
				),
			},
		},
	})
}

var testAccIPAMPoolsDataSourceConfig_basic = acctest.ConfigCompose(testAccIPAMPoolConfig_base, `
resource "aws_vpc_ipam_pool" "test" {
  address_family                    = "ipv4"
  ipam_scope_id                     = aws_vpc_ipam.test.private_default_scope_id
  auto_import                       = true
  allocation_default_netmask_length = 32
  allocation_max_netmask_length     = 32
  allocation_min_netmask_length     = 32
  allocation_resource_tags = {
    test = "1"
  }
  description = "test"
}

data "aws_vpc_ipam_pools" "test" {
  depends_on = [
    aws_vpc_ipam_pool.test
  ]
}
`)

var testAccIPAMPoolsDataSourceConfig_basicTwoPools = acctest.ConfigCompose(testAccIPAMPoolConfig_base, `
resource "aws_vpc_ipam_pool" "test" {
  address_family                    = "ipv4"
  ipam_scope_id                     = aws_vpc_ipam.test.private_default_scope_id
  auto_import                       = true
  allocation_default_netmask_length = 32
  allocation_max_netmask_length     = 32
  allocation_min_netmask_length     = 32
  allocation_resource_tags = {
    test = "1"
  }
  description = "test"
}

resource "aws_vpc_ipam_pool" "testtwo" {
  address_family = "ipv4"
  ipam_scope_id  = aws_vpc_ipam.test.private_default_scope_id
  allocation_resource_tags = {
    test = "2"
  }
  description = "testtwo"
}

resource "aws_vpc_ipam_pool" "testthree" {
  address_family                    = "ipv4"
  ipam_scope_id                     = aws_vpc_ipam.test.private_default_scope_id
  allocation_default_netmask_length = 32
  allocation_max_netmask_length     = 32
  allocation_min_netmask_length     = 32
  auto_import                       = true
  allocation_resource_tags = {
    test = "3"
  }
  description = "testthree"
  tags = {
    tagtest = 3
  }
}

data "aws_vpc_ipam_pools" "test" {
  depends_on = [
    aws_vpc_ipam_pool.test,
    aws_vpc_ipam_pool.testtwo,
    aws_vpc_ipam_pool.testthree
  ]
}

data "aws_vpc_ipam_pools" "testtwo" {
  filter {
    name   = "description"
    values = ["*three*"]
  }

  depends_on = [
    aws_vpc_ipam_pool.test,
    aws_vpc_ipam_pool.testtwo,
    aws_vpc_ipam_pool.testthree
  ]
}
`)

const testAccIPAMPoolsDataSourceConfig_empty = `
data "aws_vpc_ipam_pools" "test" {
  filter {
    name   = "description"
    values = ["*none*"]
  }
}
`

func testAccIPAMPoolsDataSourceConfig_identicalTags(rName string) string {
	return acctest.ConfigCompose(testAccIPAMPoolConfig_base, fmt.Sprintf(`
resource "aws_vpc_ipam_pool" "test1" {
  address_family = "ipv4"
  ipam_scope_id  = aws_vpc_ipam.test.private_default_scope_id
  description    = "%[1]s-1"
}

resource "aws_vpc_ipam_pool" "test2" {
  address_family = "ipv4"
  ipam_scope_id  = aws_vpc_ipam.test.private_default_scope_id
  description    = "%[1]s-2"
}

data "aws_vpc_ipam_pools" "test" {
  filter {
    name   = "description"
    values = ["%[1]s-*"]
  }

  depends_on = [
    aws_vpc_ipam_pool.test1,
    aws_vpc_ipam_pool.test2,
  ]
}
`, rName))
}
