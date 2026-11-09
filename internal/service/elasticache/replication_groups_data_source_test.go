// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package elasticache_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccElastiCacheReplicationGroupsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"
	dataSourceName := "data.aws_elasticache_replication_groups.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupsDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanOrEqualValue(dataSourceName, "replication_group_ids.#", 1),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "replication_group_ids.*", resourceName, "replication_group_id"),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroupsDataSource_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName1 := "aws_elasticache_replication_group.test1"
	resourceName2 := "aws_elasticache_replication_group.test2"
	dataSourceName := "data.aws_elasticache_replication_groups.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupsDataSourceConfig_multiple(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanOrEqualValue(dataSourceName, "replication_group_ids.#", 2),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "replication_group_ids.*", resourceName1, "replication_group_id"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "replication_group_ids.*", resourceName2, "replication_group_id"),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroupsDataSource_empty(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	dataSourceName := "data.aws_elasticache_replication_groups.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupsDataSourceConfig_empty(),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanOrEqualValue(dataSourceName, "replication_group_ids.#", 0),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrID),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroupsDataSource_valkey(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"
	dataSourceName := "data.aws_elasticache_replication_groups.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupsDataSourceConfig_valkey(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanOrEqualValue(dataSourceName, "replication_group_ids.#", 1),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "replication_group_ids.*", resourceName, "replication_group_id"),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroupsDataSource_valkeyMultiple(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName1 := "aws_elasticache_replication_group.test1"
	resourceName2 := "aws_elasticache_replication_group.test2"
	dataSourceName := "data.aws_elasticache_replication_groups.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupsDataSourceConfig_valkeyMultiple(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanOrEqualValue(dataSourceName, "replication_group_ids.#", 2),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "replication_group_ids.*", resourceName1, "replication_group_id"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "replication_group_ids.*", resourceName2, "replication_group_id"),
				),
			},
		},
	})
}

func testAccReplicationGroupsDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id = %[1]q
  description          = "test description"
  node_type            = "cache.t3.small"
  num_cache_clusters   = 1
  port                 = 6379
}

data "aws_elasticache_replication_groups" "test" {
  depends_on = [aws_elasticache_replication_group.test]
}
`, rName)
}

func testAccReplicationGroupsDataSourceConfig_multiple(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test1" {
  replication_group_id = "%[1]s-1"
  description          = "test description"
  node_type            = "cache.t3.small"
  num_cache_clusters   = 1
  port                 = 6379
}

resource "aws_elasticache_replication_group" "test2" {
  replication_group_id = "%[1]s-2"
  description          = "test description"
  node_type            = "cache.t3.small"
  num_cache_clusters   = 1
  port                 = 6379
}

data "aws_elasticache_replication_groups" "test" {
  depends_on = [aws_elasticache_replication_group.test1, aws_elasticache_replication_group.test2]
}
`, rName)
}

func testAccReplicationGroupsDataSourceConfig_empty() string {
	return `
data "aws_elasticache_replication_groups" "test" {}
`
}

func testAccReplicationGroupsDataSourceConfig_valkey(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id = %[1]q
  description          = "test description"
  engine               = "valkey"
  node_type            = "cache.t3.small"
  num_cache_clusters   = 1
  port                 = 6379
}

data "aws_elasticache_replication_groups" "test" {
  depends_on = [aws_elasticache_replication_group.test]
}
`, rName)
}

func testAccReplicationGroupsDataSourceConfig_valkeyMultiple(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test1" {
  replication_group_id = "%[1]s-1"
  description          = "test description"
  engine               = "valkey"
  node_type            = "cache.t3.small"
  num_cache_clusters   = 1
  port                 = 6379
}

resource "aws_elasticache_replication_group" "test2" {
  replication_group_id = "%[1]s-2"
  description          = "test description"
  engine               = "valkey"
  node_type            = "cache.t3.small"
  num_cache_clusters   = 1
  port                 = 6379
}

data "aws_elasticache_replication_groups" "test" {
  depends_on = [aws_elasticache_replication_group.test1, aws_elasticache_replication_group.test2]
}
`, rName)
}
