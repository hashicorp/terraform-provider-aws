// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticache_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccElastiCacheClustersDataSource_filter(t *testing.T) {
	ctx := acctest.Context(t)
	var cacheCluster types.CacheCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_elasticache_clusters.test"
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccClustersDataSourceConfig_filter(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, t, resourceName, &cacheCluster),
					resource.TestCheckResourceAttr(dataSourceName, "cluster_arns.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cluster_arns.0", resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, "cluster_identifiers.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cluster_identifiers.0", resourceName, names.AttrClusterIdentifier),
				),
			},
		},
	})
}

func testAccClustersDataSourceConfig_filter(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  cluster_id      = %[1]q
  engine          = "memcached"
  node_type       = "cache.t3.small"
  num_cache_nodes = 1
  port            = 11211
}

data "aws_elasticache_clusters" "test" {
  filter {
    name   = "cache-cluster-id"
    values = [aws_elasticache_cluster.test.cluster_identifier]
  }

  depends_on = [aws_elasticache_cluster.test]
}
`, rName)
}
