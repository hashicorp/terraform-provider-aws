// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticache_test

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfelasticache "github.com/hashicorp/terraform-provider-aws/internal/service/elasticache"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(names.ElastiCacheServiceID, testAccErrorCheckSkip)
}

func testAccErrorCheckSkip(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"is not suppored in this region",
	)
}

func TestAccElastiCacheCluster_Engine_memcached(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var ec awstypes.CacheCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_engineMemcached(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, "cache_nodes.0.id", "0001"),
					resource.TestCheckResourceAttrSet(resourceName, "configuration_endpoint"),
					resource.TestCheckResourceAttrSet(resourceName, "cluster_address"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, "memcached"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "11211"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
				},
			},
		},
	})
}

func TestAccElastiCacheCluster_Engine_redis(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var ec awstypes.CacheCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_engineRedis(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, names.AttrAutoMinorVersionUpgrade, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "cache_nodes.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "cache_nodes.0.id", "0001"),
					resource.TestCheckResourceAttr(resourceName, "cache_nodes.0.outpost_arn", ""),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, "redis"),
					resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexache.MustCompile(`^7\.[[:digit:]]+\.[[:digit:]]+$`)),
					resource.TestCheckResourceAttr(resourceName, "ip_discovery", "ipv4"),
					resource.TestCheckResourceAttr(resourceName, "network_type", "ipv4"),
					resource.TestCheckNoResourceAttr(resourceName, "outpost_mode"),
					resource.TestCheckResourceAttr(resourceName, "preferred_outpost_arn", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
				},
			},
		},
	})
}

func TestAccElastiCacheCluster_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var ec awstypes.CacheCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_engineRedis(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &ec),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfelasticache.ResourceCluster(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccElastiCacheCluster_Engine_redis_v5(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var ec awstypes.CacheCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_engineRedisV5(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, "redis"),
					resource.TestCheckResourceAttr(resourceName, "engine_version_actual", "5.0.6"),
					// Even though it is ignored, the API returns `true` in this case
					resource.TestCheckResourceAttr(resourceName, names.AttrAutoMinorVersionUpgrade, acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
				},
			},
		},
	})
}

func TestAccElastiCacheCluster_Engine_None(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_engineNone(rName),
				// Verify "ExactlyOneOf" in the schema for "engine" and "replication_group_id"
				// throws a plan-time error when neither are configured.
				ExpectError: regexache.MustCompile(`Invalid combination of arguments`),
			},
		},
	})
}

func TestAccElastiCacheCluster_PortRedis_default(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var ec awstypes.CacheCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_redisDefaultPort(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, "aws_elasticache_cluster.test", &ec),
					resource.TestCheckResourceAttr("aws_security_group_rule.test", "to_port", "6379"),
					resource.TestCheckResourceAttr("aws_security_group_rule.test", "from_port", "6379"),
				),
			},
		},
	})
}

func TestAccElastiCacheCluster_ParameterGroupName_default(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var ec awstypes.CacheCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_parameterGroupName(rName, "memcached", "1.4.34", "default.memcached1.4"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, "memcached"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, "1.4.34"),
					resource.TestCheckResourceAttr(resourceName, "engine_version_actual", "1.4.34"),
					resource.TestCheckResourceAttr(resourceName, names.AttrParameterGroupName, "default.memcached1.4"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
				},
			},
		},
	})
}

func TestAccElastiCacheCluster_ipDiscovery(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var ec awstypes.CacheCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_ipDiscovery(rName, "ipv6", "dual_stack"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, "ip_discovery", "ipv6"),
					resource.TestCheckResourceAttr(resourceName, "network_type", "dual_stack"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
				},
			},
		},
	})
}

func TestAccElastiCacheCluster_port(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var ec awstypes.CacheCluster
	port := 11212
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_port(rName, port),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, "cache_nodes.0.id", "0001"),
					resource.TestCheckResourceAttrSet(resourceName, "configuration_endpoint"),
					resource.TestCheckResourceAttrSet(resourceName, "cluster_address"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, "memcached"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, strconv.Itoa(port)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
				},
			},
		},
	})
}

func TestAccElastiCacheCluster_snapshotsWithUpdates(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var ec awstypes.CacheCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_snapshots(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, "aws_elasticache_cluster.test", &ec),
					resource.TestCheckResourceAttr("aws_elasticache_cluster.test", "snapshot_window", "05:00-09:00"),
					resource.TestCheckResourceAttr("aws_elasticache_cluster.test", "snapshot_retention_limit", acctest.Ct3),
				),
			},
			{
				Config: testAccClusterConfig_snapshotsUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, "aws_elasticache_cluster.test", &ec),
					resource.TestCheckResourceAttr("aws_elasticache_cluster.test", "snapshot_window", "07:00-09:00"),
					resource.TestCheckResourceAttr("aws_elasticache_cluster.test", "snapshot_retention_limit", "7"),
				),
			},
		},
	})
}

func TestAccElastiCacheCluster_NumCacheNodes_decrease(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var ec awstypes.CacheCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_numCacheNodes(rName, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, "num_cache_nodes", acctest.Ct3),
				),
			},
			{
				Config: testAccClusterConfig_numCacheNodes(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, "num_cache_nodes", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccElastiCacheCluster_NumCacheNodes_increase(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var ec awstypes.CacheCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_numCacheNodes(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, "num_cache_nodes", acctest.Ct1),
				),
			},
			{
				Config: testAccClusterConfig_numCacheNodes(rName, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, "num_cache_nodes", acctest.Ct3),
				),
			},
		},
	})
}

func TestAccElastiCacheCluster_NumCacheNodes_increaseWithPreferredAvailabilityZones(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var ec awstypes.CacheCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_numCacheNodesPreferredAvailabilityZones(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, "num_cache_nodes", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "preferred_availability_zones.#", acctest.Ct1),
				),
			},
			{
				Config: testAccClusterConfig_numCacheNodesPreferredAvailabilityZones(rName, 3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, "num_cache_nodes", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "preferred_availability_zones.#", acctest.Ct3),
				),
			},
		},
	})
}

func TestAccElastiCacheCluster_vpc(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var csg awstypes.CacheSubnetGroup
	var ec awstypes.CacheCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_inVPC(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetGroupExists(ctx, "aws_elasticache_subnet_group.test", &csg),
					testAccCheckClusterExists(ctx, "aws_elasticache_cluster.test", &ec),
					testAccCheckClusterAttributes(&ec),
				),
			},
		},
	})
}

func TestAccElastiCacheCluster_multiAZInVPC(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var csg awstypes.CacheSubnetGroup
	var ec awstypes.CacheCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_multiAZInVPC(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetGroupExists(ctx, "aws_elasticache_subnet_group.test", &csg),
					testAccCheckClusterExists(ctx, "aws_elasticache_cluster.test", &ec),
					resource.TestCheckResourceAttr("aws_elasticache_cluster.test", names.AttrAvailabilityZone, "Multiple"),
				),
			},
		},
	})
}

func TestAccElastiCacheCluster_AZMode_memcached(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var cluster awstypes.CacheCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccClusterConfig_azModeMemcached(rName, "unknown"),
				ExpectError: regexache.MustCompile(`expected az_mode to be one of .*, got unknown`),
			},
			{
				Config:      testAccClusterConfig_azModeMemcached(rName, "cross-az"),
				ExpectError: regexache.MustCompile(`az_mode "cross-az" is not supported with num_cache_nodes = 1`),
			},
			{
				Config: testAccClusterConfig_azModeMemcached(rName, "single-az"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "az_mode", "single-az"),
				),
			},
			{
				Config:      testAccClusterConfig_azModeMemcached(rName, "cross-az"),
				ExpectError: regexache.MustCompile(`az_mode "cross-az" is not supported with num_cache_nodes = 1`),
			},
		},
	})
}

func TestAccElastiCacheCluster_AZMode_redis(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var cluster awstypes.CacheCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccClusterConfig_azModeRedis(rName, "unknown"),
				ExpectError: regexache.MustCompile(`expected az_mode to be one of .*, got unknown`),
			},
			{
				Config:      testAccClusterConfig_azModeRedis(rName, "cross-az"),
				ExpectError: regexache.MustCompile(`az_mode "cross-az" is not supported with num_cache_nodes = 1`),
			},
			{
				Config: testAccClusterConfig_azModeRedis(rName, "single-az"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, "az_mode", "single-az"),
				),
			},
		},
	})
}

func TestAccElastiCacheCluster_EngineVersion_memcached(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var pre, mid, post awstypes.CacheCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_engineVersionMemcached(rName, "1.4.33"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &pre),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, "1.4.33"),
					resource.TestCheckResourceAttr(resourceName, "engine_version_actual", "1.4.33"),
				),
			},
			{
				Config: testAccClusterConfig_engineVersionMemcached(rName, "1.4.24"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &mid),
					testAccCheckClusterRecreated(&pre, &mid),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, "1.4.24"),
					resource.TestCheckResourceAttr(resourceName, "engine_version_actual", "1.4.24"),
				),
			},
			{
				Config: testAccClusterConfig_engineVersionMemcached(rName, "1.4.34"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &post),
					testAccCheckClusterNotRecreated(&mid, &post),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, "1.4.34"),
					resource.TestCheckResourceAttr(resourceName, "engine_version_actual", "1.4.34"),
				),
			},
		},
	})
}

func TestAccElastiCacheCluster_EngineVersion_redis(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v1, v2, v3, v4, v5, v6 awstypes.CacheCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_engineVersionRedis(rName, "4.0.10"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, "4.0.10"),
					resource.TestCheckResourceAttr(resourceName, "engine_version_actual", "4.0.10"),
				),
			},
			{
				Config: testAccClusterConfig_engineVersionRedis(rName, "6.0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v2),
					testAccCheckClusterNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, "6.0"),
					resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexache.MustCompile(`^6\.0\.[[:digit:]]+$`)),
				),
			},
			{
				Config: testAccClusterConfig_engineVersionRedis(rName, "6.2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v3),
					testAccCheckClusterNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, "6.2"),
					resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexache.MustCompile(`^6\.2\.[[:digit:]]+$`)),
				),
			},
			{
				Config: testAccClusterConfig_engineVersionRedis(rName, "5.0.6"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v4),
					testAccCheckClusterRecreated(&v3, &v4),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, "5.0.6"),
					resource.TestCheckResourceAttr(resourceName, "engine_version_actual", "5.0.6"),
				),
			},
			{
				Config: testAccClusterConfig_engineVersionRedis(rName, "6.x"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v5),
					testAccCheckClusterNotRecreated(&v4, &v5),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, "6.x"),
					resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexache.MustCompile(`^6\.[[:digit:]]+\.[[:digit:]]+$`)),
				),
			},
			{
				Config: testAccClusterConfig_engineVersionRedis(rName, "6.0"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &v6),
					testAccCheckClusterRecreated(&v5, &v6),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, "6.0"),
					resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexache.MustCompile(`^6\.0\.[[:digit:]]+$`)),
				),
			},
		},
	})
}

func TestAccElastiCacheCluster_NodeTypeResize_memcached(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var pre, post awstypes.CacheCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_nodeTypeMemcached(rName, "cache.t3.small"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &pre),
					resource.TestCheckResourceAttr(resourceName, "node_type", "cache.t3.small"),
				),
			},
			{
				Config: testAccClusterConfig_nodeTypeMemcached(rName, "cache.t3.medium"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &post),
					testAccCheckClusterRecreated(&pre, &post),
					resource.TestCheckResourceAttr(resourceName, "node_type", "cache.t3.medium"),
				),
			},
		},
	})
}

func TestAccElastiCacheCluster_NodeTypeResize_redis(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var pre, post awstypes.CacheCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_nodeTypeRedis(rName, "cache.t3.small"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &pre),
					resource.TestCheckResourceAttr(resourceName, "node_type", "cache.t3.small"),
				),
			},
			{
				Config: testAccClusterConfig_nodeTypeRedis(rName, "cache.t3.medium"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &post),
					testAccCheckClusterNotRecreated(&pre, &post),
					resource.TestCheckResourceAttr(resourceName, "node_type", "cache.t3.medium"),
				),
			},
		},
	})
}

func TestAccElastiCacheCluster_NumCacheNodes_redis(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccClusterConfig_numCacheNodesRedis(rName, 2),
				ExpectError: regexache.MustCompile(`engine "redis" does not support num_cache_nodes > 1`),
			},
		},
	})
}

func TestAccElastiCacheCluster_ReplicationGroupID_availabilityZone(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var cluster awstypes.CacheCluster
	var replicationGroup awstypes.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	clusterResourceName := "aws_elasticache_cluster.test"
	replicationGroupResourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_replicationGroupIDAvailabilityZone(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, replicationGroupResourceName, &replicationGroup),
					testAccCheckClusterExists(ctx, clusterResourceName, &cluster),
					testAccCheckClusterReplicationGroupIDAttribute(&cluster, &replicationGroup),
				),
			},
		},
	})
}

func TestAccElastiCacheCluster_ReplicationGroupID_transitEncryption(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var cluster awstypes.CacheCluster
	var replicationGroup awstypes.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	clusterResourceName := "aws_elasticache_cluster.test"
	replicationGroupResourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_replicationGroupIDTransitEncryption(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, replicationGroupResourceName, &replicationGroup),
					testAccCheckClusterExists(ctx, clusterResourceName, &cluster),
					testAccCheckClusterReplicationGroupIDAttribute(&cluster, &replicationGroup),
					resource.TestCheckResourceAttr(clusterResourceName, "transit_encryption_enabled", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccElastiCacheCluster_ReplicationGroupID_singleReplica(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var cluster awstypes.CacheCluster
	var replicationGroup awstypes.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	clusterResourceName := "aws_elasticache_cluster.test.0"
	replicationGroupResourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_replicationGroupIDReplica(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, replicationGroupResourceName, &replicationGroup),
					testAccCheckClusterExists(ctx, clusterResourceName, &cluster),
					testAccCheckClusterReplicationGroupIDAttribute(&cluster, &replicationGroup),
					resource.TestCheckResourceAttr(clusterResourceName, names.AttrEngine, "redis"),
					resource.TestCheckResourceAttr(clusterResourceName, "node_type", "cache.t3.medium"),
					resource.TestCheckResourceAttr(clusterResourceName, names.AttrPort, "6379"),
				),
			},
		},
	})
}

func TestAccElastiCacheCluster_ReplicationGroupID_multipleReplica(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var cluster1, cluster2 awstypes.CacheCluster
	var replicationGroup awstypes.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	clusterResourceName1 := "aws_elasticache_cluster.test.0"
	clusterResourceName2 := "aws_elasticache_cluster.test.1"
	replicationGroupResourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_replicationGroupIDReplica(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, replicationGroupResourceName, &replicationGroup),

					testAccCheckClusterExists(ctx, clusterResourceName1, &cluster1),
					testAccCheckClusterReplicationGroupIDAttribute(&cluster1, &replicationGroup),
					resource.TestCheckResourceAttr(clusterResourceName1, names.AttrEngine, "redis"),
					resource.TestCheckResourceAttr(clusterResourceName1, "node_type", "cache.t3.medium"),
					resource.TestCheckResourceAttr(clusterResourceName1, names.AttrPort, "6379"),

					testAccCheckClusterExists(ctx, clusterResourceName2, &cluster2),
					testAccCheckClusterReplicationGroupIDAttribute(&cluster2, &replicationGroup),
					resource.TestCheckResourceAttr(clusterResourceName2, names.AttrEngine, "redis"),
					resource.TestCheckResourceAttr(clusterResourceName2, "node_type", "cache.t3.medium"),
					resource.TestCheckResourceAttr(clusterResourceName2, names.AttrPort, "6379"),
				),
			},
		},
	})
}

func TestAccElastiCacheCluster_Memcached_finalSnapshot(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccClusterConfig_memcachedFinalSnapshot(rName),
				ExpectError: regexache.MustCompile(`engine "memcached" does not support final_snapshot_identifier`),
			},
		},
	})
}

func TestAccElastiCacheCluster_Redis_finalSnapshot(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var cluster awstypes.CacheCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_redisFinalSnapshot(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, names.AttrFinalSnapshotIdentifier, rName),
				),
			},
		},
	})
}

func TestAccElastiCacheCluster_Redis_autoMinorVersionUpgrade(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var cluster awstypes.CacheCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_redisAutoMinorVersionUpgrade(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, names.AttrAutoMinorVersionUpgrade, acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
				},
			},
			{
				Config: testAccClusterConfig_redisAutoMinorVersionUpgrade(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, names.AttrAutoMinorVersionUpgrade, acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccElastiCacheCluster_Engine_Redis_LogDeliveryConfigurations(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var ec awstypes.CacheCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_dataSourceEngineRedisLogDeliveryConfigurations(rName, true, awstypes.DestinationTypeCloudWatchLogs, awstypes.LogFormatText, true, awstypes.DestinationTypeCloudWatchLogs, awstypes.LogFormatText),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, "redis"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.destination", rName),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.destination_type", "cloudwatch-logs"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.log_format", "text"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.log_type", "engine-log"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.1.destination", rName),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.1.destination_type", "cloudwatch-logs"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.1.log_format", "text"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.1.log_type", "slow-log"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately},
			},
			{
				Config: testAccClusterConfig_dataSourceEngineRedisLogDeliveryConfigurations(rName, true, awstypes.DestinationTypeKinesisFirehose, awstypes.LogFormatJson, true, awstypes.DestinationTypeKinesisFirehose, awstypes.LogFormatJson),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, "redis"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.destination", rName),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.destination_type", "kinesis-firehose"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.log_format", names.AttrJSON),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.log_type", "engine-log"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.1.destination", rName),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.1.destination_type", "kinesis-firehose"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.1.log_format", names.AttrJSON),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.1.log_type", "slow-log"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately},
			},
			{
				Config: testAccClusterConfig_dataSourceEngineRedisLogDeliveryConfigurations(rName, true, awstypes.DestinationTypeCloudWatchLogs, awstypes.LogFormatText, true, awstypes.DestinationTypeKinesisFirehose, awstypes.LogFormatJson),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, "redis"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.destination", rName),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.destination_type", "cloudwatch-logs"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.log_format", "text"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.log_type", "slow-log"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.1.destination", rName),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.1.destination_type", "kinesis-firehose"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.1.log_format", names.AttrJSON),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.1.log_type", "engine-log"),
				),
			},
			{
				Config: testAccClusterConfig_dataSourceEngineRedisLogDeliveryConfigurations(rName, true, awstypes.DestinationTypeKinesisFirehose, awstypes.LogFormatJson, true, awstypes.DestinationTypeCloudWatchLogs, awstypes.LogFormatText),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, "redis"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.destination", rName),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.destination_type", "cloudwatch-logs"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.log_format", "text"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.log_type", "engine-log"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.1.destination", rName),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.1.destination_type", "kinesis-firehose"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.1.log_format", names.AttrJSON),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.1.log_type", "slow-log"),
				),
			},
			{
				Config: testAccClusterConfig_dataSourceEngineRedisLogDeliveryConfigurations(rName, false, "", "", false, "", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, "redis"),
					resource.TestCheckNoResourceAttr(resourceName, "log_delivery_configuration.0.destination"),
					resource.TestCheckNoResourceAttr(resourceName, "log_delivery_configuration.0.destination_type"),
					resource.TestCheckNoResourceAttr(resourceName, "log_delivery_configuration.0.log_format"),
					resource.TestCheckNoResourceAttr(resourceName, "log_delivery_configuration.0.log_type"),
					resource.TestCheckNoResourceAttr(resourceName, "log_delivery_configuration.1.destination"),
					resource.TestCheckNoResourceAttr(resourceName, "log_delivery_configuration.1.destination_type"),
					resource.TestCheckNoResourceAttr(resourceName, "log_delivery_configuration.1.log_format"),
					resource.TestCheckNoResourceAttr(resourceName, "log_delivery_configuration.1.log_type"),
				),
			},
			{
				Config: testAccClusterConfig_dataSourceEngineRedisLogDeliveryConfigurations(rName, true, awstypes.DestinationTypeKinesisFirehose, awstypes.LogFormatJson, false, "", ""),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, "redis"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.destination", rName),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.destination_type", "kinesis-firehose"),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.log_format", names.AttrJSON),
					resource.TestCheckResourceAttr(resourceName, "log_delivery_configuration.0.log_type", "slow-log"),
					resource.TestCheckNoResourceAttr(resourceName, "log_delivery_configuration.1.destination"),
					resource.TestCheckNoResourceAttr(resourceName, "log_delivery_configuration.1.destination_type"),
					resource.TestCheckNoResourceAttr(resourceName, "log_delivery_configuration.1.log_format"),
					resource.TestCheckNoResourceAttr(resourceName, "log_delivery_configuration.1.log_type"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately},
			},
		},
	})
}

func TestAccElastiCacheCluster_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var cluster awstypes.CacheCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", acctest.CtValue1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately}, //not in the API
			},
			{
				Config: testAccClusterConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key2", acctest.CtValue2),
				),
			},
			{
				Config: testAccClusterConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key2", acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccElastiCacheCluster_tagWithOtherModification(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var cluster awstypes.CacheCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_versionAndTag(rName, "5.0.4", acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, "5.0.4"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", acctest.CtValue1),
				),
			},
			{
				Config: testAccClusterConfig_versionAndTag(rName, "5.0.6", acctest.CtKey1, acctest.CtValue1Updated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, "5.0.6"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", acctest.CtValue1Updated),
				),
			},
		},
	})
}

func TestAccElastiCacheCluster_TransitEncryption(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	var cluster awstypes.CacheCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccClusterConfig_transitEncryption(rName, "memcached", "1.6.6"),
				ExpectError: regexache.MustCompile(`InvalidParameterCombination: Encryption features are not supported for engine version 1.6.6. Please use engine version 1.6.12`),
			},
			{
				Config:      testAccClusterConfig_transitEncryption(rName, "redis", "6.2"),
				ExpectError: regexache.MustCompile(`InvalidParameterCombination: Encryption feature is not supported for engine REDIS.`),
			},
			{
				Config: testAccClusterConfig_transitEncryption(rName, "memcached", "1.6.12"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &cluster),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, "memcached"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, "1.6.12"),
					resource.TestCheckResourceAttr(resourceName, "transit_encryption_enabled", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccElastiCacheCluster_outpost_memcached(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var ec awstypes.CacheCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_outpost_memcached(rName, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, "cache_nodes.0.id", "0001"),
					resource.TestCheckResourceAttrSet(resourceName, "cache_nodes.0.outpost_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "configuration_endpoint"),
					resource.TestCheckResourceAttrSet(resourceName, "cluster_address"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, "memcached"),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "11211"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
				},
			},
		},
	})
}

func TestAccElastiCacheCluster_outpost_redis(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var ec awstypes.CacheCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_outpost_redis(rName, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &ec),
					resource.TestCheckResourceAttr(resourceName, "cache_nodes.0.id", "0001"),
					resource.TestCheckResourceAttrSet(resourceName, "cache_nodes.0.outpost_arn"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, "redis"),
					resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexache.MustCompile(`^7\.[[:digit:]]+\.[[:digit:]]+$`)),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "6379"),
					resource.TestCheckResourceAttr(resourceName, names.AttrAutoMinorVersionUpgrade, acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
				},
			},
		},
	})
}

func TestAccElastiCacheCluster_outpostID_memcached(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var pre, post awstypes.CacheCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_outpost_memcached(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &pre),
				),
			},
			{
				Config: testAccClusterConfig_outpost_memcached(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &post),
					testAccCheckClusterRecreated(&pre, &post),
				),
			},
		},
	})
}

func TestAccElastiCacheCluster_outpostID_redis(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var pre, post awstypes.CacheCluster
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOutpostsOutposts(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_outpost_redis(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &pre),
				),
			},
			{
				Config: testAccClusterConfig_outpost_redis(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckClusterExists(ctx, resourceName, &post),
					testAccCheckClusterNotRecreated(&pre, &post),
				),
			},
		},
	})
}

func testAccCheckClusterAttributes(v *awstypes.CacheCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if v.NotificationConfiguration == nil {
			return fmt.Errorf("Expected NotificationConfiguration for ElastiCache Cluster (%s)", *v.CacheClusterId)
		}

		if strings.ToLower(*v.NotificationConfiguration.TopicStatus) != "active" {
			return fmt.Errorf("Expected NotificationConfiguration status to be 'active', got (%s)", *v.NotificationConfiguration.TopicStatus)
		}

		return nil
	}
}

func testAccCheckClusterReplicationGroupIDAttribute(cluster *awstypes.CacheCluster, replicationGroup *awstypes.ReplicationGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if cluster.ReplicationGroupId == nil {
			return errors.New("expected cluster ReplicationGroupId to be set")
		}

		if aws.ToString(cluster.ReplicationGroupId) != aws.ToString(replicationGroup.ReplicationGroupId) {
			return errors.New("expected cluster ReplicationGroupId to equal replication group ID")
		}

		return nil
	}
}

func testAccCheckClusterNotRecreated(i, j *awstypes.CacheCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !aws.ToTime(i.CacheClusterCreateTime).Equal(aws.ToTime(j.CacheClusterCreateTime)) {
			return errors.New("ElastiCache Cluster was recreated")
		}

		return nil
	}
}

func testAccCheckClusterRecreated(i, j *awstypes.CacheCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToTime(i.CacheClusterCreateTime).Equal(aws.ToTime(j.CacheClusterCreateTime)) {
			return errors.New("ElastiCache Cluster was not recreated")
		}

		return nil
	}
}

func testAccCheckClusterDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ElastiCacheClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_elasticache_cluster" {
				continue
			}

			_, err := tfelasticache.FindCacheClusterByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ElastiCache Cluster %s still exists", rs.Primary.ID)
		}
		return nil
	}
}

func testAccCheckClusterExists(ctx context.Context, n string, v *awstypes.CacheCluster) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ElastiCache Cluster ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ElastiCacheClient(ctx)

		output, err := tfelasticache.FindCacheClusterByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccClusterConfig_engineMemcached(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  cluster_id      = %[1]q
  engine          = "memcached"
  node_type       = "cache.t3.small"
  num_cache_nodes = 1
}
`, rName)
}

func testAccClusterConfig_engineRedis(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  cluster_id      = %[1]q
  engine          = "redis"
  node_type       = "cache.t3.small"
  num_cache_nodes = 1
}
`, rName)
}

func testAccClusterConfig_engineRedisV5(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  cluster_id      = %[1]q
  engine_version  = "5.0.6"
  engine          = "redis"
  node_type       = "cache.t3.small"
  num_cache_nodes = 1
}
`, rName)
}

func testAccClusterConfig_engineNone(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  cluster_id      = %[1]q
  node_type       = "cache.t3.small"
  num_cache_nodes = 1
}
`, rName)
}

func testAccClusterConfig_outpost_memcached(rName string, outpostID int) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[%[2]d]
}

resource "aws_elasticache_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_elasticache_cluster" "test" {
  cluster_id            = %[1]q
  outpost_mode          = "single-outpost"
  preferred_outpost_arn = data.aws_outposts_outpost.test.arn
  engine                = "memcached"
  node_type             = "cache.r5.large"
  num_cache_nodes       = 1
  subnet_group_name     = aws_elasticache_subnet_group.test.name
}
`, rName, outpostID))
}

func testAccClusterConfig_outpost_redis(rName string, outpostID int) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[%[2]d]
}

resource "aws_elasticache_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_elasticache_cluster" "test" {
  cluster_id            = %[1]q
  outpost_mode          = "single-outpost"
  preferred_outpost_arn = data.aws_outposts_outpost.test.arn
  engine                = "redis"
  node_type             = "cache.r5.large"
  num_cache_nodes       = 1
  subnet_group_name     = aws_elasticache_subnet_group.test.name
}
`, rName, outpostID))
}

func testAccClusterConfig_parameterGroupName(rName, engine, engineVersion, parameterGroupName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  cluster_id           = %[1]q
  engine               = %[2]q
  engine_version       = %[3]q
  node_type            = "cache.t3.small"
  num_cache_nodes      = 1
  parameter_group_name = %[4]q
}
`, rName, engine, engineVersion, parameterGroupName)
}

func testAccClusterConfig_ipDiscovery(rName, ipDiscovery, networkType string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnetsIPv6(rName, 1), fmt.Sprintf(`
resource "aws_elasticache_subnet_group" "test" {
  name        = %[1]q
  description = %[1]q
  subnet_ids  = aws_subnet.test[*].id
}

resource "aws_security_group" "test" {
  name        = %[1]q
  description = %[1]q
  vpc_id      = aws_vpc.test.id

  ingress {
    from_port   = -1
    to_port     = -1
    protocol    = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_elasticache_cluster" "test" {
  cluster_id      = %[1]q
  engine          = "memcached"
  engine_version  = "1.6.6"
  node_type       = "cache.t3.small"
  num_cache_nodes = 1
  ip_discovery    = %[2]q
  network_type    = %[3]q

  subnet_group_name  = aws_elasticache_subnet_group.test.name
  security_group_ids = [aws_security_group.test.id]
  availability_zone  = data.aws_availability_zones.available.names[0]
}
`, rName, ipDiscovery, networkType))
}

func testAccClusterConfig_port(rName string, port int) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  cluster_id      = %[1]q
  engine          = "memcached"
  node_type       = "cache.t3.small"
  num_cache_nodes = 1
  port            = %[2]d
}
`, rName, port)
}

func testAccClusterConfig_snapshots(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  cluster_id               = %[1]q
  engine                   = "redis"
  node_type                = "cache.t3.small"
  num_cache_nodes          = 1
  port                     = 6379
  snapshot_window          = "05:00-09:00"
  snapshot_retention_limit = 3
}
`, rName)
}

func testAccClusterConfig_snapshotsUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  cluster_id               = %[1]q
  engine                   = "redis"
  node_type                = "cache.t3.small"
  num_cache_nodes          = 1
  port                     = 6379
  snapshot_window          = "07:00-09:00"
  snapshot_retention_limit = 7
  apply_immediately        = true
}
`, rName)
}

func testAccClusterConfig_numCacheNodes(rName string, numCacheNodes int) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  apply_immediately = true
  cluster_id        = %[1]q
  engine            = "memcached"
  node_type         = "cache.t3.small"
  num_cache_nodes   = %[2]d
}
`, rName, numCacheNodes)
}

func testAccClusterConfig_numCacheNodesPreferredAvailabilityZones(rName string, numCacheNodes int) string {
	preferredAvailabilityZones := make([]string, numCacheNodes)
	for i := range preferredAvailabilityZones {
		preferredAvailabilityZones[i] = `data.aws_availability_zones.available.names[0]`
	}

	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  apply_immediately            = true
  cluster_id                   = %[1]q
  engine                       = "memcached"
  node_type                    = "cache.t3.small"
  num_cache_nodes              = %[2]d
  preferred_availability_zones = [%[3]s]
}
`, rName, numCacheNodes, strings.Join(preferredAvailabilityZones, ",")))
}

func testAccClusterConfig_inVPC(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 1), fmt.Sprintf(`
resource "aws_elasticache_subnet_group" "test" {
  name        = %[1]q
  description = %[1]q
  subnet_ids  = aws_subnet.test[*].id
}

resource "aws_security_group" "test" {
  name        = %[1]q
  description = %[1]q
  vpc_id      = aws_vpc.test.id

  ingress {
    from_port   = -1
    to_port     = -1
    protocol    = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_elasticache_cluster" "test" {
  # Including uppercase letters in this name to ensure
  # that we correctly handle the fact that the API
  # normalizes names to lowercase.
  cluster_id             = %[1]q
  node_type              = "cache.t3.small"
  num_cache_nodes        = 1
  engine                 = "redis"
  port                   = 6379
  subnet_group_name      = aws_elasticache_subnet_group.test.name
  security_group_ids     = [aws_security_group.test.id]
  notification_topic_arn = aws_sns_topic.test.arn
  availability_zone      = data.aws_availability_zones.available.names[0]
}

resource "aws_sns_topic" "test" {
  name = %[1]q
}
`, rName))
}

func testAccClusterConfig_multiAZInVPC(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_elasticache_subnet_group" "test" {
  name        = %[1]q
  description = %[1]q
  subnet_ids  = aws_subnet.test[*].id
}

resource "aws_security_group" "test" {
  name        = %[1]q
  description = %[1]q
  vpc_id      = aws_vpc.test.id

  ingress {
    from_port   = -1
    to_port     = -1
    protocol    = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_elasticache_cluster" "test" {
  cluster_id         = %[1]q
  engine             = "memcached"
  node_type          = "cache.t3.small"
  num_cache_nodes    = 2
  port               = 11211
  subnet_group_name  = aws_elasticache_subnet_group.test.name
  security_group_ids = [aws_security_group.test.id]
  az_mode            = "cross-az"
  preferred_availability_zones = [
    data.aws_availability_zones.available.names[0],
    data.aws_availability_zones.available.names[1]
  ]
}
`, rName))
}

func testAccClusterConfig_redisDefaultPort(rName string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name        = %[1]q
  description = %[1]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group_rule" "test" {
  cidr_blocks       = ["0.0.0.0/0"]
  from_port         = aws_elasticache_cluster.test.port
  protocol          = "tcp"
  security_group_id = aws_security_group.test.id
  to_port           = aws_elasticache_cluster.test.port
  type              = "ingress"
}

resource "aws_elasticache_cluster" "test" {
  cluster_id      = %[1]q
  engine          = "redis"
  node_type       = "cache.t2.micro"
  num_cache_nodes = 1
}
`, rName)
}

func testAccClusterConfig_azModeMemcached(rName, azMode string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  apply_immediately = true
  az_mode           = %[2]q
  cluster_id        = %[1]q
  engine            = "memcached"
  node_type         = "cache.t3.medium"
  num_cache_nodes   = 1
  port              = 11211
}
`, rName, azMode)
}

func testAccClusterConfig_azModeRedis(rName, azMode string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  apply_immediately = true
  az_mode           = %[2]q
  cluster_id        = %[1]q
  engine            = "redis"
  node_type         = "cache.t3.medium"
  num_cache_nodes   = 1
  port              = 6379
}
`, rName, azMode)
}

func testAccClusterConfig_engineVersionMemcached(rName, engineVersion string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  apply_immediately = true
  cluster_id        = %[1]q
  engine            = "memcached"
  engine_version    = %[2]q
  node_type         = "cache.t3.medium"
  num_cache_nodes   = 1
  port              = 11211
}
`, rName, engineVersion)
}

func testAccClusterConfig_engineVersionRedis(rName, engineVersion string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  apply_immediately = true
  cluster_id        = %[1]q
  engine            = "redis"
  engine_version    = %[2]q
  node_type         = "cache.t3.medium"
  num_cache_nodes   = 1
  port              = 6379
}
`, rName, engineVersion)
}

func testAccClusterConfig_nodeTypeMemcached(rName, nodeType string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  apply_immediately = true
  cluster_id        = %[1]q
  engine            = "memcached"
  node_type         = %[2]q
  num_cache_nodes   = 1
  port              = 11211
}
`, rName, nodeType)
}

func testAccClusterConfig_nodeTypeRedis(rName, nodeType string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  apply_immediately = true
  cluster_id        = %[1]q
  engine            = "redis"
  node_type         = %[2]q
  num_cache_nodes   = 1
  port              = 6379
}
`, rName, nodeType)
}

func testAccClusterConfig_numCacheNodesRedis(rName string, numCacheNodes int) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  apply_immediately = true
  cluster_id        = %[1]q
  engine            = "redis"
  node_type         = "cache.t3.medium"
  num_cache_nodes   = %[2]d
  port              = 6379
}
`, rName, numCacheNodes)
}

func testAccClusterConfig_replicationGroupIDAvailabilityZone(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  description          = "Terraform Acceptance Testing"
  replication_group_id = %[1]q
  node_type            = "cache.t3.medium"
  num_cache_clusters   = 1
  port                 = 6379

  lifecycle {
    ignore_changes = [num_cache_clusters]
  }
}

resource "aws_elasticache_cluster" "test" {
  availability_zone    = data.aws_availability_zones.available.names[0]
  cluster_id           = "%[1]s-1"
  replication_group_id = aws_elasticache_replication_group.test.id
}
`, rName))
}

func testAccClusterConfig_replicationGroupIDTransitEncryption(rName string, enabled bool) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptIn(), fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  description                = "Terraform Acceptance Testing"
  replication_group_id       = %[1]q
  node_type                  = "cache.t3.medium"
  num_cache_clusters         = 1
  port                       = 6379
  transit_encryption_enabled = %[2]t

  lifecycle {
    ignore_changes = [num_cache_clusters]
  }
}

resource "aws_elasticache_cluster" "test" {
  availability_zone    = data.aws_availability_zones.available.names[0]
  cluster_id           = "%[1]s-1"
  replication_group_id = aws_elasticache_replication_group.test.id
}
`, rName, enabled))
}

func testAccClusterConfig_replicationGroupIDReplica(rName string, count int) string {
	return fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  description          = "Terraform Acceptance Testing"
  replication_group_id = %[1]q
  node_type            = "cache.t3.medium"
  num_cache_clusters   = 1
  port                 = 6379

  lifecycle {
    ignore_changes = [num_cache_clusters]
  }
}

resource "aws_elasticache_cluster" "test" {
  count                = %[2]d
  cluster_id           = "%[1]s-${count.index}"
  replication_group_id = aws_elasticache_replication_group.test.id
}
`, rName, count)
}

func testAccClusterConfig_memcachedFinalSnapshot(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  cluster_id      = %[1]q
  engine          = "memcached"
  node_type       = "cache.t3.small"
  num_cache_nodes = 1

  final_snapshot_identifier = %[1]q
}
`, rName)
}

func testAccClusterConfig_redisFinalSnapshot(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  cluster_id      = %[1]q
  engine          = "redis"
  node_type       = "cache.t3.small"
  num_cache_nodes = 1

  final_snapshot_identifier = %[1]q
}
`, rName)
}

func testAccClusterConfig_redisAutoMinorVersionUpgrade(rName string, enable bool) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  cluster_id      = %[1]q
  engine          = "redis"
  engine_version  = "6.0"
  node_type       = "cache.t3.small"
  num_cache_nodes = 1

  auto_minor_version_upgrade = %[2]t
}
`, rName, enable)
}

func testAccClusterConfig_dataSourceEngineRedisLogDeliveryConfigurations(rName string, slowLogDeliveryEnabled bool, slowDeliveryDestination awstypes.DestinationType, slowDeliveryFormat awstypes.LogFormat, engineLogDeliveryEnabled bool, engineDeliveryDestination awstypes.DestinationType, engineLogDeliveryFormat awstypes.LogFormat) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "p" {
  statement {
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents"
    ]
    resources = ["${aws_cloudwatch_log_group.lg.arn}:log-stream:*"]
    principals {
      identifiers = ["delivery.logs.amazonaws.com"]
      type        = "Service"
    }
  }
}

resource "aws_cloudwatch_log_resource_policy" "rp" {
  policy_document = data.aws_iam_policy_document.p.json
  policy_name     = %[1]q
  depends_on = [
    aws_cloudwatch_log_group.lg
  ]
}

resource "aws_cloudwatch_log_group" "lg" {
  retention_in_days = 1
  name              = %[1]q
}

resource "aws_s3_bucket" "b" {
  force_destroy = true
}

resource "aws_iam_role" "r" {
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "firehose.amazonaws.com"
        }
      },
    ]
  })
  inline_policy {
    name = "my_inline_s3_policy"
    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          Action = [
            "s3:AbortMultipartUpload",
            "s3:GetBucketLocation",
            "s3:GetObject",
            "s3:ListBucket",
            "s3:ListBucketMultipartUploads",
            "s3:PutObject",
            "s3:PutObjectAcl",
          ]
          Effect   = "Allow"
          Resource = [aws_s3_bucket.b.arn, "${aws_s3_bucket.b.arn}/*"]
        },
      ]
    })
  }
}

resource "aws_kinesis_firehose_delivery_stream" "ds" {
  name        = %[1]q
  destination = "extended_s3"

  extended_s3_configuration {
    bucket_arn = aws_s3_bucket.b.arn
    role_arn   = aws_iam_role.r.arn
  }

  lifecycle {
    ignore_changes = [
      tags["LogDeliveryEnabled"],
    ]
  }
}

resource "aws_elasticache_cluster" "test" {
  cluster_id        = %[1]q
  engine            = "redis"
  node_type         = "cache.t3.micro"
  num_cache_nodes   = 1
  port              = 6379
  apply_immediately = true
  dynamic "log_delivery_configuration" {
    for_each = tobool("%[2]t") ? [""] : []
    content {
      destination      = (%[3]q == "cloudwatch-logs") ? aws_cloudwatch_log_group.lg.name : ((%[3]q == "kinesis-firehose") ? aws_kinesis_firehose_delivery_stream.ds.name : null)
      destination_type = %[3]q
      log_format       = %[4]q
      log_type         = "slow-log"
    }
  }
  dynamic "log_delivery_configuration" {
    for_each = tobool("%[5]t") ? [""] : []
    content {
      destination      = (%[6]q == "cloudwatch-logs") ? aws_cloudwatch_log_group.lg.name : ((%[6]q == "kinesis-firehose") ? aws_kinesis_firehose_delivery_stream.ds.name : null)
      destination_type = %[6]q
      log_format       = %[7]q
      log_type         = "engine-log"
    }
  }
}

data "aws_elasticache_cluster" "test" {
  cluster_id = aws_elasticache_cluster.test.cluster_id
}
`, rName, slowLogDeliveryEnabled, slowDeliveryDestination, slowDeliveryFormat, engineLogDeliveryEnabled, engineDeliveryDestination, engineLogDeliveryFormat)
}

func testAccClusterConfig_tags1(rName, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  cluster_id      = %[1]q
  engine          = "memcached"
  node_type       = "cache.t3.small"
  num_cache_nodes = 1

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tag1Key, tag1Value)
}

func testAccClusterConfig_tags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  cluster_id      = %[1]q
  engine          = "memcached"
  node_type       = "cache.t3.small"
  num_cache_nodes = 1

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tag1Key, tag1Value, tag2Key, tag2Value)
}

func testAccClusterConfig_versionAndTag(rName, version, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  cluster_id        = %[1]q
  node_type         = "cache.t3.small"
  num_cache_nodes   = 1
  engine            = "redis"
  engine_version    = %[2]q
  apply_immediately = true

  tags = {
    %[3]q = %[4]q
  }
}
`, rName, version, tagKey1, tagValue1)
}

func testAccClusterConfig_transitEncryption(rName, engine, version string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  apply_immediately          = true
  cluster_id                 = "%[1]s"
  engine                     = "%[2]s"
  engine_version             = "%[3]s"
  node_type                  = "cache.t3.medium"
  num_cache_nodes            = 1
  transit_encryption_enabled = true
}
`, rName, engine, version)
}
