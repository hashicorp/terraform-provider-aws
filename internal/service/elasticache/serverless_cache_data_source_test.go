// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticache_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccElastiCacheServerlessCacheDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_serverless_cache.test"
	dataSourceName := "data.aws_elasticache_serverless_cache.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ElastiCacheEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServerlessCacheDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrCreateTime, resourceName, names.AttrCreateTime),
					resource.TestCheckResourceAttrPair(dataSourceName, "daily_snapshot_time", resourceName, "daily_snapshot_time"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrEngine, resourceName, names.AttrEngine),
					resource.TestCheckResourceAttrPair(dataSourceName, "full_engine_version", resourceName, "full_engine_version"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrKMSKeyID, resourceName, names.AttrKMSKeyID),
					resource.TestCheckResourceAttrPair(dataSourceName, "major_engine_version", resourceName, "major_engine_version"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "security_group_ids.#", resourceName, "security_group_ids.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "snapshot_retention_limit", resourceName, "snapshot_retention_limit"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrStatus, resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrPair(dataSourceName, "subnet_ids.#", resourceName, "subnet_ids.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "user_group_id", resourceName, "user_group_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cache_usage_limits.#", resourceName, "cache_usage_limits.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "endpoint.address", resourceName, "endpoint.0.address"),
					resource.TestCheckResourceAttrPair(dataSourceName, "endpoint.port", resourceName, "endpoint.0.port"),
					resource.TestCheckResourceAttrPair(dataSourceName, "reader_endpoint.address", resourceName, "reader_endpoint.0.address"),
					resource.TestCheckResourceAttrPair(dataSourceName, "reader_endpoint.port", resourceName, "reader_endpoint.0.port"),
				),
			},
		},
	})
}

func testAccServerlessCacheDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_serverless_cache" "test" {
  engine = "redis"
  name   = %[1]q
}

data "aws_elasticache_serverless_cache" "test" {
  name = aws_elasticache_serverless_cache.test.name
}
`, rName)
}
