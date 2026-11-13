// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package elasticache_test

import (
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfknownvalue "github.com/hashicorp/terraform-provider-aws/internal/acctest/knownvalue"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccElastiCacheServiceUpdateActionsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	dataSourceName := "data.aws_elasticache_service_update_actions.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ElastiCacheEndpointID)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceUpdateActionsDataSourceConfig_basic(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("cache_cluster_id"), knownvalue.Null()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("replication_group_id"), knownvalue.Null()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("service_update_status"), knownvalue.Null()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("update_actions"), knownvalue.NotNull()),
				},
			},
		},
	})
}

func TestAccElastiCacheServiceUpdateActionsDataSource_replicationGroupID(t *testing.T) {
	ctx := acctest.Context(t)

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_elasticache_service_update_actions.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ElastiCacheEndpointID)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"time": {
				Source:            "hashicorp/time",
				VersionConstraint: "0.14.0",
			},
		},
		CheckDestroy: acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceUpdateActionsDataSourceConfig_replicationGroupID(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("cache_cluster_id"), knownvalue.Null()),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New("replication_group_id"), "aws_elasticache_replication_group.test", tfjsonpath.New("replication_group_id"), compare.ValuesSame()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("service_update_status"), knownvalue.Null()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("update_actions"), tfknownvalue.SetNotEmpty()),
				},
			},
		},
	})
}

func TestAccElastiCacheServiceUpdateActionsDataSource_clusterID(t *testing.T) {
	ctx := acctest.Context(t)

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_elasticache_service_update_actions.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ElastiCacheEndpointID)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"time": {
				Source:            "hashicorp/time",
				VersionConstraint: "0.14.0",
			},
		},
		CheckDestroy: acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceUpdateActionsDataSourceConfig_clusterID(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New("cache_cluster_id"), "aws_elasticache_cluster.test", tfjsonpath.New("cluster_id"), compare.ValuesSame()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("replication_group_id"), knownvalue.Null()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("service_update_status"), knownvalue.Null()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("update_actions"), tfknownvalue.SetNotEmpty()),
				},
			},
		},
	})
}

func TestAccElastiCacheServiceUpdateActionsDataSource_serviceUpdateStatus(t *testing.T) {
	ctx := acctest.Context(t)

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_elasticache_service_update_actions.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ElastiCacheEndpointID)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"time": {
				Source:            "hashicorp/time",
				VersionConstraint: "0.14.0",
			},
		},
		CheckDestroy: acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceUpdateActionsDataSourceConfig_serviceUpdateStatus(rName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("cache_cluster_id"), knownvalue.Null()),
					statecheck.CompareValuePairs(dataSourceName, tfjsonpath.New("replication_group_id"), "aws_elasticache_replication_group.test", tfjsonpath.New("replication_group_id"), compare.ValuesSame()),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("service_update_status"), knownvalue.SetExact([]knownvalue.Check{
						tfknownvalue.StringExact(awstypes.ServiceUpdateStatusAvailable),
					})),
					statecheck.ExpectKnownValue(dataSourceName, tfjsonpath.New("update_actions"), tfknownvalue.SetNotEmpty()),
				},
			},
		},
	})
}

func testAccServiceUpdateActionsDataSourceConfig_basic() string {
	return `
data "aws_elasticache_service_update_actions" "test" {}
`
}

func testAccServiceUpdateActionsDataSourceConfig_replicationGroupID(rName string) string {
	return acctest.ConfigCompose(
		testAccReplicationGroupConfig_basic_engine(rName, "valkey"), `
data "aws_elasticache_service_update_actions" "test" {
  replication_group_id = aws_elasticache_replication_group.test.replication_group_id

  depends_on = [time_sleep.wait]
}

# It takes approximately 10 minutes for Update Actions to be registered after the Replication Group becomes available.
resource "time_sleep" "wait" {
  create_duration = "10m"

  depends_on = [aws_elasticache_replication_group.test]
}
`)
}

func testAccServiceUpdateActionsDataSourceConfig_clusterID(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_engineRedis(rName), `
data "aws_elasticache_service_update_actions" "test" {
  cache_cluster_id = aws_elasticache_cluster.test.cluster_id

  depends_on = [time_sleep.wait]
}

# It takes approximately 10 minutes for Update Actions to be registered after the Cache Cluster becomes available.
resource "time_sleep" "wait" {
  create_duration = "10m"

  depends_on = [aws_elasticache_cluster.test]
}
`)
}

func testAccServiceUpdateActionsDataSourceConfig_serviceUpdateStatus(rName string) string {
	return acctest.ConfigCompose(
		testAccReplicationGroupConfig_basic_engine(rName, "valkey"), `
data "aws_elasticache_service_update_actions" "test" {
  replication_group_id = aws_elasticache_replication_group.test.replication_group_id

  service_update_status = ["available"]

  depends_on = [time_sleep.wait]
}

# It takes approximately 10 minutes for Update Actions to be registered after the Replication Group becomes available.
resource "time_sleep" "wait" {
  create_duration = "10m"

  depends_on = [aws_elasticache_replication_group.test]
}
`)
}
