// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticache_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccElastiCacheReservedNodeOffering_Redis_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_elasticache_reserved_cache_node_offering.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccReservedNodeOfferingConfig_Redis_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "cache_node_type", "cache.t4g.small"),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrDuration, "P1Y"),
					resource.TestCheckResourceAttrSet(dataSourceName, "fixed_price"),
					resource.TestCheckResourceAttrSet(dataSourceName, "offering_id"),
					resource.TestCheckResourceAttr(dataSourceName, "offering_type", "No Upfront"),
					resource.TestCheckResourceAttr(dataSourceName, "product_description", "redis"),
				),
			},
		},
	})
}

func TestAccElastiCacheReservedNodeOffering_Valkey_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_elasticache_reserved_cache_node_offering.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccReservedNodeOfferingConfig_Valkey_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "cache_node_type", "cache.t4g.small"),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrDuration, "P1Y"),
					resource.TestCheckResourceAttrSet(dataSourceName, "fixed_price"),
					resource.TestCheckResourceAttrSet(dataSourceName, "offering_id"),
					resource.TestCheckResourceAttr(dataSourceName, "offering_type", "No Upfront"),
					resource.TestCheckResourceAttr(dataSourceName, "product_description", "valkey"),
				),
			},
		},
	})
}

func testAccReservedNodeOfferingConfig_Redis_basic() string {
	return `
data "aws_elasticache_reserved_cache_node_offering" "test" {
  cache_node_type     = "cache.t4g.small"
  duration            = "P1Y"
  offering_type       = "No Upfront"
  product_description = "redis"
}
`
}

func testAccReservedNodeOfferingConfig_Valkey_basic() string {
	return `
data "aws_elasticache_reserved_cache_node_offering" "test" {
  cache_node_type     = "cache.t4g.small"
  duration            = "P1Y"
  offering_type       = "No Upfront"
  product_description = "valkey"
}
`
}
