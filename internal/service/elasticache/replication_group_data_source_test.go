// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticache_test

import (
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccElastiCacheReplicationGroupDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"
	dataSourceName := "data.aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "auth_token_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "automatic_failover_enabled", resourceName, "automatic_failover_enabled"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cluster_mode", resourceName, "cluster_mode"),
					resource.TestCheckResourceAttrPair(dataSourceName, "multi_az_enabled", resourceName, "multi_az_enabled"),
					resource.TestCheckResourceAttrPair(dataSourceName, "member_clusters.#", resourceName, "member_clusters.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "node_type", resourceName, "node_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "num_cache_clusters", resourceName, "num_cache_clusters"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrPort, resourceName, names.AttrPort),
					resource.TestCheckResourceAttrPair(dataSourceName, "primary_endpoint_address", resourceName, "primary_endpoint_address"),
					resource.TestCheckResourceAttrPair(dataSourceName, "reader_endpoint_address", resourceName, "reader_endpoint_address"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(dataSourceName, "replication_group_id", resourceName, "replication_group_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "snapshot_window", resourceName, "snapshot_window"),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroupDataSource_clusterMode(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"
	dataSourceName := "data.aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupDataSourceConfig_clusterMode(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "auth_token_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(dataSourceName, "automatic_failover_enabled", resourceName, "automatic_failover_enabled"),
					resource.TestCheckResourceAttrPair(dataSourceName, "multi_az_enabled", resourceName, "multi_az_enabled"),
					resource.TestCheckResourceAttrPair(dataSourceName, "configuration_endpoint_address", resourceName, "configuration_endpoint_address"),
					resource.TestCheckResourceAttrPair(dataSourceName, "node_type", resourceName, "node_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "num_node_groups", resourceName, "num_node_groups"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrPort, resourceName, names.AttrPort),
					resource.TestCheckResourceAttrPair(dataSourceName, "replicas_per_node_group", resourceName, "replicas_per_node_group"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDescription, resourceName, names.AttrDescription),
					resource.TestCheckResourceAttrPair(dataSourceName, "replication_group_id", resourceName, "replication_group_id"),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroupDataSource_multiAZ(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"
	dataSourceName := "data.aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupDataSourceConfig_multiAZ(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "automatic_failover_enabled", resourceName, "automatic_failover_enabled"),
					resource.TestCheckResourceAttrPair(dataSourceName, "multi_az_enabled", resourceName, "multi_az_enabled"),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroupDataSource_Engine_Redis_LogDeliveryConfigurations(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_dataSourceEngineRedisLogDeliveryConfigurations(rName, false, true, string(awstypes.DestinationTypeCloudWatchLogs), string(awstypes.LogFormatText), true, string(awstypes.DestinationTypeKinesisFirehose), string(awstypes.LogFormatJson)),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "log_delivery_configuration.0.destination", rName),
					resource.TestCheckResourceAttr(dataSourceName, "log_delivery_configuration.0.destination_type", "cloudwatch-logs"),
					resource.TestCheckResourceAttr(dataSourceName, "log_delivery_configuration.0.log_format", "text"),
					resource.TestCheckResourceAttr(dataSourceName, "log_delivery_configuration.0.log_type", "slow-log"),
					resource.TestCheckResourceAttr(dataSourceName, "log_delivery_configuration.1.destination", rName),
					resource.TestCheckResourceAttr(dataSourceName, "log_delivery_configuration.1.destination_type", "kinesis-firehose"),
					resource.TestCheckResourceAttr(dataSourceName, "log_delivery_configuration.1.log_format", names.AttrJSON),
					resource.TestCheckResourceAttr(dataSourceName, "log_delivery_configuration.1.log_type", "engine-log"),
				),
			},
		},
	})
}

func testAccReplicationGroupDataSourceConfig_basic(rName string) string {
	return acctest.ConfigAvailableAZsNoOptIn() + fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id        = %[1]q
  description                 = "test description"
  node_type                   = "cache.t3.small"
  num_cache_clusters          = 2
  port                        = 6379
  preferred_cache_cluster_azs = [data.aws_availability_zones.available.names[0], data.aws_availability_zones.available.names[1]]
  automatic_failover_enabled  = true
  snapshot_window             = "01:00-02:00"
}

data "aws_elasticache_replication_group" "test" {
  replication_group_id = aws_elasticache_replication_group.test.replication_group_id
}
`, rName)
}

func testAccReplicationGroupDataSourceConfig_clusterMode(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id       = %[1]q
  description                = "test description"
  node_type                  = "cache.t3.small"
  port                       = 6379
  automatic_failover_enabled = true

  replicas_per_node_group = 1
  num_node_groups         = 2
}

data "aws_elasticache_replication_group" "test" {
  replication_group_id = aws_elasticache_replication_group.test.replication_group_id
}
`, rName)
}

func testAccReplicationGroupDataSourceConfig_multiAZ(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id       = %[1]q
  description                = "test description"
  node_type                  = "cache.t3.small"
  num_cache_clusters         = 2
  automatic_failover_enabled = true
  multi_az_enabled           = true
}

data "aws_elasticache_replication_group" "test" {
  replication_group_id = aws_elasticache_replication_group.test.replication_group_id
}
`, rName)
}
