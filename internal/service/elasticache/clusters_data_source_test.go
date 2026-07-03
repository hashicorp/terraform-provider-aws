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

func TestAccElastiCacheClustersDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elasticache_cluster.test"
	dataSourceName := "data.aws_elasticache_clusters.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClustersDataSourceConfig_basic(rName),
				// The data source is account/region-wide, so assert membership
				// (the created ID is present) rather than exact list equality.
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanOrEqualValue(dataSourceName, "cluster_ids.#", 1),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "cluster_ids.*", resourceName, "cluster_id"),
				),
			},
		},
	})
}

func TestAccElastiCacheClustersDataSource_multiple(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName1 := "aws_elasticache_cluster.test1"
	resourceName2 := "aws_elasticache_cluster.test2"
	dataSourceName := "data.aws_elasticache_clusters.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClustersDataSourceConfig_multiple(rName),
				// Each created identifier must appear in the aggregated list.
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanOrEqualValue(dataSourceName, "cluster_ids.#", 2),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "cluster_ids.*", resourceName1, "cluster_id"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "cluster_ids.*", resourceName2, "cluster_id"),
				),
			},
		},
	})
}

func TestAccElastiCacheClustersDataSource_empty(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	dataSourceName := "data.aws_elasticache_clusters.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClustersDataSourceConfig_empty(),
				// The read succeeds and cluster_ids is a valid list (possibly
				// empty). In a shared account other clusters may exist, so we
				// only assert the attribute is present and a valid list.
				Check: resource.ComposeAggregateTestCheckFunc(
					acctest.CheckResourceAttrGreaterThanOrEqualValue(dataSourceName, "cluster_ids.#", 0),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrID),
				),
			},
		},
	})
}

func testAccClustersDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  cluster_id      = %[1]q
  engine          = "memcached"
  node_type       = "cache.t3.small"
  num_cache_nodes = 1
  port            = 11211
}

data "aws_elasticache_clusters" "test" {
  depends_on = [aws_elasticache_cluster.test]
}
`, rName)
}

func testAccClustersDataSourceConfig_multiple(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test1" {
  cluster_id      = "%[1]s-1"
  engine          = "memcached"
  node_type       = "cache.t3.small"
  num_cache_nodes = 1
  port            = 11211
}

resource "aws_elasticache_cluster" "test2" {
  cluster_id      = "%[1]s-2"
  engine          = "memcached"
  node_type       = "cache.t3.small"
  num_cache_nodes = 1
  port            = 11211
}

data "aws_elasticache_clusters" "test" {
  depends_on = [aws_elasticache_cluster.test1, aws_elasticache_cluster.test2]
}
`, rName)
}

func testAccClustersDataSourceConfig_empty() string {
	return `
data "aws_elasticache_clusters" "test" {}
`
}
