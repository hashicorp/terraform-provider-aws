// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package elasticache_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfelasticache "github.com/hashicorp/terraform-provider-aws/internal/service/elasticache"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccElastiCacheGlobalReplicationGroup_Redis_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalReplicationGroup awstypes.GlobalReplicationGroup
	var primaryReplicationGroup awstypes.ReplicationGroup

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	primaryReplicationGroupId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resourceName := "aws_elasticache_global_replication_group.test"
	primaryReplicationGroupResourceName := "aws_elasticache_replication_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckGlobalReplicationGroupDestroy(ctx, t),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_Redis_basic(rName, primaryReplicationGroupId),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, t, resourceName, &globalReplicationGroup),
					testAccCheckReplicationGroupExists(ctx, t, primaryReplicationGroupResourceName, &primaryReplicationGroup),
					acctest.MatchResourceAttrGlobalARN(ctx, resourceName, names.AttrARN, "elasticache", regexache.MustCompile(`globalreplicationgroup:`+tfelasticache.GlobalReplicationGroupRegionPrefixFormat+rName)),
					resource.TestCheckResourceAttrPair(resourceName, "at_rest_encryption_enabled", primaryReplicationGroupResourceName, "at_rest_encryption_enabled"),
					resource.TestCheckResourceAttr(resourceName, "auth_token_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, "automatic_failover_enabled", primaryReplicationGroupResourceName, "automatic_failover_enabled"),
					resource.TestCheckResourceAttrPair(resourceName, "cache_node_type", primaryReplicationGroupResourceName, "node_type"),
					resource.TestCheckResourceAttrPair(resourceName, "cluster_enabled", primaryReplicationGroupResourceName, "cluster_enabled"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrEngine, primaryReplicationGroupResourceName, names.AttrEngine),
					resource.TestCheckResourceAttrPair(resourceName, "engine_version_actual", primaryReplicationGroupResourceName, "engine_version_actual"),
					resource.TestCheckResourceAttr(resourceName, "global_replication_group_id_suffix", rName),
					resource.TestMatchResourceAttr(resourceName, "global_replication_group_id", regexache.MustCompile(tfelasticache.GlobalReplicationGroupRegionPrefixFormat+rName)),
					resource.TestCheckResourceAttr(resourceName, "global_replication_group_description", tfelasticache.EmptyDescription),
					resource.TestCheckResourceAttr(resourceName, "global_node_groups.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "num_node_groups", "0"),
					resource.TestCheckResourceAttr(resourceName, "primary_replication_group_id", primaryReplicationGroupId),
					resource.TestCheckResourceAttr(resourceName, "transit_encryption_enabled", acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_Valkey_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalReplicationGroup awstypes.GlobalReplicationGroup
	var primaryReplicationGroup awstypes.ReplicationGroup

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	primaryReplicationGroupId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resourceName := "aws_elasticache_global_replication_group.test"
	primaryReplicationGroupResourceName := "aws_elasticache_replication_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckGlobalReplicationGroupDestroy(ctx, t),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_Valkey_basic(rName, primaryReplicationGroupId),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, t, resourceName, &globalReplicationGroup),
					testAccCheckReplicationGroupExists(ctx, t, primaryReplicationGroupResourceName, &primaryReplicationGroup),
					acctest.MatchResourceAttrGlobalARN(ctx, resourceName, names.AttrARN, "elasticache", regexache.MustCompile(`globalreplicationgroup:`+tfelasticache.GlobalReplicationGroupRegionPrefixFormat+rName)),
					resource.TestCheckResourceAttrPair(resourceName, "at_rest_encryption_enabled", primaryReplicationGroupResourceName, "at_rest_encryption_enabled"),
					resource.TestCheckResourceAttr(resourceName, "auth_token_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, "automatic_failover_enabled", primaryReplicationGroupResourceName, "automatic_failover_enabled"),
					resource.TestCheckResourceAttrPair(resourceName, "cache_node_type", primaryReplicationGroupResourceName, "node_type"),
					resource.TestCheckResourceAttrPair(resourceName, "cluster_enabled", primaryReplicationGroupResourceName, "cluster_enabled"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrEngine, primaryReplicationGroupResourceName, names.AttrEngine),
					resource.TestCheckResourceAttrPair(resourceName, "engine_version_actual", primaryReplicationGroupResourceName, "engine_version_actual"),
					resource.TestCheckResourceAttr(resourceName, "global_replication_group_id_suffix", rName),
					resource.TestMatchResourceAttr(resourceName, "global_replication_group_id", regexache.MustCompile(tfelasticache.GlobalReplicationGroupRegionPrefixFormat+rName)),
					resource.TestCheckResourceAttr(resourceName, "global_replication_group_description", tfelasticache.EmptyDescription),
					resource.TestCheckResourceAttr(resourceName, "global_node_groups.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "num_node_groups", "0"),
					resource.TestCheckResourceAttr(resourceName, "primary_replication_group_id", primaryReplicationGroupId),
					resource.TestCheckResourceAttr(resourceName, "transit_encryption_enabled", acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalReplicationGroup awstypes.GlobalReplicationGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	primaryReplicationGroupId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elasticache_global_replication_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_Redis_basic(rName, primaryReplicationGroupId),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, t, resourceName, &globalReplicationGroup),
					acctest.CheckSDKResourceDisappears(ctx, t, tfelasticache.ResourceGlobalReplicationGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_description(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalReplicationGroup awstypes.GlobalReplicationGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	primaryReplicationGroupId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	description1 := sdkacctest.RandString(10)
	description2 := sdkacctest.RandString(10)
	resourceName := "aws_elasticache_global_replication_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_description(rName, primaryReplicationGroupId, description1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, t, resourceName, &globalReplicationGroup),
					resource.TestCheckResourceAttr(resourceName, "global_replication_group_description", description1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGlobalReplicationGroupConfig_description(rName, primaryReplicationGroupId, description2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, t, resourceName, &globalReplicationGroup),
					resource.TestCheckResourceAttr(resourceName, "global_replication_group_description", description2),
				),
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_nodeType_createNoChange(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalReplicationGroup awstypes.GlobalReplicationGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	primaryReplicationGroupId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	nodeType := "cache.m5.large"
	resourceName := "aws_elasticache_global_replication_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_nodeType_createNoChange(rName, primaryReplicationGroupId, nodeType),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, t, resourceName, &globalReplicationGroup),
					resource.TestCheckResourceAttr(resourceName, "cache_node_type", nodeType),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_nodeType_createWithChange(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalReplicationGroup awstypes.GlobalReplicationGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	primaryReplicationGroupId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	nodeType := "cache.m5.large"
	globalNodeType := "cache.m5.xlarge"
	resourceName := "aws_elasticache_global_replication_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_nodeType_createWithChange(rName, primaryReplicationGroupId, nodeType, globalNodeType),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, t, resourceName, &globalReplicationGroup),
					resource.TestCheckResourceAttr(resourceName, "cache_node_type", globalNodeType),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_nodeType_setNoChange(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalReplicationGroup awstypes.GlobalReplicationGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	primaryReplicationGroupId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	nodeType := "cache.m5.large"
	resourceName := "aws_elasticache_global_replication_group.test"
	primaryReplicationGroupResourceName := "aws_elasticache_replication_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_basic_nodeType(rName, primaryReplicationGroupId, nodeType),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, t, resourceName, &globalReplicationGroup),
					resource.TestCheckResourceAttrPair(resourceName, "cache_node_type", primaryReplicationGroupResourceName, "node_type"),
				),
			},
			{
				Config: testAccGlobalReplicationGroupConfig_nodeType_createNoChange(rName, primaryReplicationGroupId, nodeType),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, t, resourceName, &globalReplicationGroup),
					resource.TestCheckResourceAttr(resourceName, "cache_node_type", nodeType),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_nodeType_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalReplicationGroup awstypes.GlobalReplicationGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	primaryReplicationGroupId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	nodeType := "cache.m5.large"
	updatedNodeType := "cache.m5.xlarge"
	resourceName := "aws_elasticache_global_replication_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_nodeType_createNoChange(rName, primaryReplicationGroupId, nodeType),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, t, resourceName, &globalReplicationGroup),
					resource.TestCheckResourceAttr(resourceName, "cache_node_type", nodeType),
				),
			},
			{
				Config: testAccGlobalReplicationGroupConfig_nodeType_update(rName, primaryReplicationGroupId, updatedNodeType),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, t, resourceName, &globalReplicationGroup),
					resource.TestCheckResourceAttr(resourceName, "cache_node_type", updatedNodeType),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_automaticFailover_createNoChange(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalReplicationGroup awstypes.GlobalReplicationGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	primaryReplicationGroupId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elasticache_global_replication_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_automaticFailover_createNoChange(rName, primaryReplicationGroupId, acctest.CtTrue),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, t, resourceName, &globalReplicationGroup),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_automaticFailover_createWithChange(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalReplicationGroup awstypes.GlobalReplicationGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	primaryReplicationGroupId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elasticache_global_replication_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_automaticFailover_createWithChange(rName, primaryReplicationGroupId, acctest.CtFalse, acctest.CtTrue),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, t, resourceName, &globalReplicationGroup),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", acctest.CtTrue),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_automaticFailover_setNoChange(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalReplicationGroup awstypes.GlobalReplicationGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	primaryReplicationGroupId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elasticache_global_replication_group.test"
	primaryReplicationGroupResourceName := "aws_elasticache_replication_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_basic_automaticFailover(rName, primaryReplicationGroupId, acctest.CtFalse),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, t, resourceName, &globalReplicationGroup),
					resource.TestCheckResourceAttrPair(resourceName, "automatic_failover_enabled", primaryReplicationGroupResourceName, "automatic_failover_enabled"),
				),
			},
			{
				Config: testAccGlobalReplicationGroupConfig_automaticFailover_createNoChange(rName, primaryReplicationGroupId, acctest.CtFalse),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, t, resourceName, &globalReplicationGroup),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_automaticFailover_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalReplicationGroup awstypes.GlobalReplicationGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	primaryReplicationGroupId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elasticache_global_replication_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_automaticFailover_createNoChange(rName, primaryReplicationGroupId, acctest.CtTrue),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, t, resourceName, &globalReplicationGroup),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", acctest.CtTrue),
				),
			},
			{
				Config: testAccGlobalReplicationGroupConfig_automaticFailover_update(rName, primaryReplicationGroupId, acctest.CtFalse),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, t, resourceName, &globalReplicationGroup),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", acctest.CtFalse),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_multipleSecondaries(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalReplcationGroup awstypes.GlobalReplicationGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elasticache_global_replication_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 3)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 3),
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_multipleSecondaries(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, t, resourceName, &globalReplcationGroup),
				),
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_ReplaceSecondary_differentRegion(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalReplcationGroup awstypes.GlobalReplicationGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elasticache_global_replication_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 3)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 3),
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_replaceSecondaryDifferentRegionSetup(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, t, resourceName, &globalReplcationGroup),
				),
			},
			{
				Config: testAccGlobalReplicationGroupConfig_replaceSecondaryDifferentRegionMove(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, t, resourceName, &globalReplcationGroup),
				),
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_clusterMode_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalReplicationGroup awstypes.GlobalReplicationGroup
	var primaryReplicationGroup awstypes.ReplicationGroup

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resourceName := "aws_elasticache_global_replication_group.test"
	primaryReplicationGroupResourceName := "aws_elasticache_replication_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_clusterMode(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, t, resourceName, &globalReplicationGroup),
					testAccCheckReplicationGroupExists(ctx, t, primaryReplicationGroupResourceName, &primaryReplicationGroup),
					resource.TestCheckResourceAttrPair(resourceName, "automatic_failover_enabled", primaryReplicationGroupResourceName, "automatic_failover_enabled"),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrPair(resourceName, "global_node_groups.#", primaryReplicationGroupResourceName, "num_node_groups"),
					resource.TestCheckResourceAttrPair(resourceName, "num_node_groups", primaryReplicationGroupResourceName, "num_node_groups"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_SetNumNodeGroupsOnCreate_NoChange(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalReplicationGroup awstypes.GlobalReplicationGroup
	var primaryReplicationGroup awstypes.ReplicationGroup

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resourceName := "aws_elasticache_global_replication_group.test"
	primaryReplicationGroupResourceName := "aws_elasticache_replication_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_numNodeGroups(rName, 2, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, t, resourceName, &globalReplicationGroup),
					testAccCheckReplicationGroupExists(ctx, t, primaryReplicationGroupResourceName, &primaryReplicationGroup),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "global_node_groups.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "num_node_groups", "2"),
					resource.TestCheckResourceAttr(primaryReplicationGroupResourceName, "num_node_groups", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_SetNumNodeGroupsOnCreate_Increase(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalReplicationGroup awstypes.GlobalReplicationGroup
	var primaryReplicationGroup awstypes.ReplicationGroup

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resourceName := "aws_elasticache_global_replication_group.test"
	primaryReplicationGroupResourceName := "aws_elasticache_replication_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_numNodeGroups(rName, 2, 3),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, t, resourceName, &globalReplicationGroup),
					testAccCheckReplicationGroupExists(ctx, t, primaryReplicationGroupResourceName, &primaryReplicationGroup),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "global_node_groups.#", "3"),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "global_node_groups.*", map[string]*regexp.Regexp{
						"global_node_group_id": regexache.MustCompile(fmt.Sprintf("^[a-z]+-%s-0001$", rName)),
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "global_node_groups.*", map[string]*regexp.Regexp{
						"global_node_group_id": regexache.MustCompile(fmt.Sprintf("^[a-z]+-%s-0002$", rName)),
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "global_node_groups.*", map[string]*regexp.Regexp{
						"global_node_group_id": regexache.MustCompile(fmt.Sprintf("^[a-z]+-%s-0003$", rName)),
					}),
					resource.TestCheckResourceAttr(resourceName, "num_node_groups", "3"),
					resource.TestCheckResourceAttr(primaryReplicationGroupResourceName, "num_node_groups", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_SetNumNodeGroupsOnCreate_Decrease(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalReplicationGroup awstypes.GlobalReplicationGroup
	var primaryReplicationGroup awstypes.ReplicationGroup

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resourceName := "aws_elasticache_global_replication_group.test"
	primaryReplicationGroupResourceName := "aws_elasticache_replication_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_numNodeGroups(rName, 3, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, t, resourceName, &globalReplicationGroup),
					testAccCheckReplicationGroupExists(ctx, t, primaryReplicationGroupResourceName, &primaryReplicationGroup),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "global_node_groups.#", "1"),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "global_node_groups.*", map[string]*regexp.Regexp{
						"global_node_group_id": regexache.MustCompile(fmt.Sprintf("^[a-z]+-%s-0001$", rName)),
						"slots":                regexache.MustCompile("^0-16383$"), // all slots
					}),
					resource.TestCheckResourceAttr(resourceName, "num_node_groups", "1"),
					resource.TestCheckResourceAttr(primaryReplicationGroupResourceName, "num_node_groups", "3"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_SetNumNodeGroupsOnUpdate_Increase(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalReplicationGroup awstypes.GlobalReplicationGroup
	var primaryReplicationGroup awstypes.ReplicationGroup

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resourceName := "aws_elasticache_global_replication_group.test"
	primaryReplicationGroupResourceName := "aws_elasticache_replication_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_numNodeGroups_inherit(rName, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, t, resourceName, &globalReplicationGroup),
					testAccCheckReplicationGroupExists(ctx, t, primaryReplicationGroupResourceName, &primaryReplicationGroup),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "global_node_groups.#", "2"),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "global_node_groups.*", map[string]*regexp.Regexp{
						"global_node_group_id": regexache.MustCompile(fmt.Sprintf("^[a-z]+-%s-0001$", rName)),
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "global_node_groups.*", map[string]*regexp.Regexp{
						"global_node_group_id": regexache.MustCompile(fmt.Sprintf("^[a-z]+-%s-0002$", rName)),
					}),
					resource.TestCheckResourceAttr(resourceName, "num_node_groups", "2"),
					resource.TestCheckResourceAttr(primaryReplicationGroupResourceName, "num_node_groups", "2"),
				),
			},
			{
				Config: testAccGlobalReplicationGroupConfig_numNodeGroups(rName, 2, 3),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, t, resourceName, &globalReplicationGroup),
					testAccCheckReplicationGroupExists(ctx, t, primaryReplicationGroupResourceName, &primaryReplicationGroup),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "global_node_groups.#", "3"),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "global_node_groups.*", map[string]*regexp.Regexp{
						"global_node_group_id": regexache.MustCompile(fmt.Sprintf("^[a-z]+-%s-0001$", rName)),
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "global_node_groups.*", map[string]*regexp.Regexp{
						"global_node_group_id": regexache.MustCompile(fmt.Sprintf("^[a-z]+-%s-0002$", rName)),
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "global_node_groups.*", map[string]*regexp.Regexp{
						"global_node_group_id": regexache.MustCompile(fmt.Sprintf("^[a-z]+-%s-0003$", rName)),
					}),
					resource.TestCheckResourceAttr(resourceName, "num_node_groups", "3"),
					resource.TestCheckResourceAttr(primaryReplicationGroupResourceName, "num_node_groups", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_SetNumNodeGroupsOnUpdate_Decrease(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalReplicationGroup awstypes.GlobalReplicationGroup
	var primaryReplicationGroup awstypes.ReplicationGroup

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resourceName := "aws_elasticache_global_replication_group.test"
	primaryReplicationGroupResourceName := "aws_elasticache_replication_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_numNodeGroups_inherit(rName, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, t, resourceName, &globalReplicationGroup),
					testAccCheckReplicationGroupExists(ctx, t, primaryReplicationGroupResourceName, &primaryReplicationGroup),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "global_node_groups.#", "2"),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "global_node_groups.*", map[string]*regexp.Regexp{
						"global_node_group_id": regexache.MustCompile(fmt.Sprintf("^[a-z]+-%s-0001$", rName)),
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "global_node_groups.*", map[string]*regexp.Regexp{
						"global_node_group_id": regexache.MustCompile(fmt.Sprintf("^[a-z]+-%s-0002$", rName)),
					}),
					resource.TestCheckResourceAttr(resourceName, "num_node_groups", "2"),
					resource.TestCheckResourceAttr(primaryReplicationGroupResourceName, "num_node_groups", "2"),
				),
			},
			{
				Config: testAccGlobalReplicationGroupConfig_numNodeGroups(rName, 2, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, t, resourceName, &globalReplicationGroup),
					testAccCheckReplicationGroupExists(ctx, t, primaryReplicationGroupResourceName, &primaryReplicationGroup),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "global_node_groups.#", "1"),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "global_node_groups.*", map[string]*regexp.Regexp{
						"global_node_group_id": regexache.MustCompile(fmt.Sprintf("^[a-z]+-%s-0001$", rName)),
						"slots":                regexache.MustCompile("^0-16383$"), // all slots
					}),
					resource.TestCheckResourceAttr(resourceName, "num_node_groups", "1"),
					resource.TestCheckResourceAttr(primaryReplicationGroupResourceName, "num_node_groups", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_SetEngineVersionOnCreate_NoChange_Redis_v6(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalReplicationGroup awstypes.GlobalReplicationGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	primaryReplicationGroupId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elasticache_global_replication_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_Redis_engineVersion(rName, primaryReplicationGroupId, "6.2", "6.2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, t, resourceName, &globalReplicationGroup),
					resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexache.MustCompile(`^6\.2\.[[:digit:]]+$`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_SetEngineVersionOnCreate_NoChange_Redis_v6x(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalReplicationGroup awstypes.GlobalReplicationGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	primaryReplicationGroupId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elasticache_global_replication_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_Redis_engineVersion(rName, primaryReplicationGroupId, "6.2", "6.x"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, t, resourceName, &globalReplicationGroup),
					resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexache.MustCompile(`^6\.2\.[[:digit:]]+$`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrEngineVersion},
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_SetEngineVersionOnCreate_NoChange_Redis_v5(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalReplicationGroup awstypes.GlobalReplicationGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	primaryReplicationGroupId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elasticache_global_replication_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_Redis_engineVersion(rName, primaryReplicationGroupId, "5.0.6", "5.0.6"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, t, resourceName, &globalReplicationGroup),
					resource.TestCheckResourceAttr(resourceName, "engine_version_actual", "5.0.6"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_SetEngineVersionOnCreate_NoChange_Valkey_v7(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalReplicationGroup awstypes.GlobalReplicationGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	primaryReplicationGroupId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elasticache_global_replication_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_Valkey_engineVersion(rName, primaryReplicationGroupId, "7.2", "7.2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, t, resourceName, &globalReplicationGroup),
					resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexache.MustCompile(`^7\.2\.[[:digit:]]+$`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_SetEngineVersionOnCreate_MinorUpgrade(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalReplicationGroup awstypes.GlobalReplicationGroup
	var rg awstypes.ReplicationGroup

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	primaryReplicationGroupId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elasticache_global_replication_group.test"
	primaryReplicationGroupResourceName := "aws_elasticache_replication_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_Redis_engineVersion(rName, primaryReplicationGroupId, "6.0", "6.2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, t, resourceName, &globalReplicationGroup),
					testAccCheckReplicationGroupExists(ctx, t, primaryReplicationGroupResourceName, &rg),
					resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexache.MustCompile(`^6\.2\.[[:digit:]]+$`)),
					testAccMatchReplicationGroupActualVersion(ctx, t, &rg, regexache.MustCompile(`^6\.2\.[[:digit:]]+$`)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_SetEngineVersionOnCreate_MinorUpgrade_6x(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalReplicationGroup awstypes.GlobalReplicationGroup
	var rg awstypes.ReplicationGroup

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	primaryReplicationGroupId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elasticache_global_replication_group.test"
	primaryReplicationGroupResourceName := "aws_elasticache_replication_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_Redis_engineVersion(rName, primaryReplicationGroupId, "6.0", "6.x"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, t, resourceName, &globalReplicationGroup),
					testAccCheckReplicationGroupExists(ctx, t, primaryReplicationGroupResourceName, &rg),
					resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexache.MustCompile(`^6\.0\.[[:digit:]]+$`)),
					testAccMatchReplicationGroupActualVersion(ctx, t, &rg, regexache.MustCompile(`^6\.0\.[[:digit:]]+$`)),
				),
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_SetEngineVersionOnCreate_MajorUpgrade(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalReplicationGroup awstypes.GlobalReplicationGroup
	var rg awstypes.ReplicationGroup

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	primaryReplicationGroupId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elasticache_global_replication_group.test"
	primaryReplicationGroupResourceName := "aws_elasticache_replication_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_engineVersionParam(rName, primaryReplicationGroupId, "5.0.6", "6.2", "default.redis6.x"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, t, resourceName, &globalReplicationGroup),
					testAccCheckReplicationGroupExists(ctx, t, primaryReplicationGroupResourceName, &rg),
					resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexache.MustCompile(`^6\.2\.[[:digit:]]+$`)),
					testAccMatchReplicationGroupActualVersion(ctx, t, &rg, regexache.MustCompile(`^6\.2\.[[:digit:]]+$`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrParameterGroupName},
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_SetEngineVersionOnCreate_MajorUpgrade_6x(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalReplicationGroup awstypes.GlobalReplicationGroup
	var rg awstypes.ReplicationGroup

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	primaryReplicationGroupId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elasticache_global_replication_group.test"
	primaryReplicationGroupResourceName := "aws_elasticache_replication_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_engineVersionParam(rName, primaryReplicationGroupId, "5.0.6", "6.2", "default.redis6.x"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, t, resourceName, &globalReplicationGroup),
					testAccCheckReplicationGroupExists(ctx, t, primaryReplicationGroupResourceName, &rg),
					resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexache.MustCompile(`^6\.2\.[[:digit:]]+$`)),
					testAccMatchReplicationGroupActualVersion(ctx, t, &rg, regexache.MustCompile(`^6\.2\.[[:digit:]]+$`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrParameterGroupName},
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_SetEngineVersionOnCreate_MinorDowngrade(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	primaryReplicationGroupId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccGlobalReplicationGroupConfig_Redis_engineVersion(rName, primaryReplicationGroupId, "6.2", "6.0"),
				ExpectError: regexache.MustCompile(`cannot downgrade version when creating, is 6.2.[[:digit:]]+, want 6.0.[[:digit:]]+`),
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_SetParameterGroupOnCreate_NoVersion(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	primaryReplicationGroupId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccGlobalReplicationGroupConfig_param(rName, primaryReplicationGroupId, "6.2", "default.redis6.x"),
				ExpectError: regexache.MustCompile(`cannot change parameter group name without upgrading major engine version`),
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_SetParameterGroupOnCreate_MinorUpgrade(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	primaryReplicationGroupId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config:      testAccGlobalReplicationGroupConfig_engineVersionParam(rName, primaryReplicationGroupId, "6.0", "6.2", "default.redis6.x"),
				ExpectError: regexache.MustCompile(`cannot change parameter group name on minor engine version upgrade, upgrading from 6\.0\.[[:digit:]]+ to 6\.2\.[[:digit:]]+`),
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_SetEngineVersionOnUpdate_MinorUpgrade(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalReplicationGroup awstypes.GlobalReplicationGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	primaryReplicationGroupId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elasticache_global_replication_group.test"
	primaryReplicationGroupResourceName := "aws_elasticache_replication_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_Redis_engineVersionInherit(rName, primaryReplicationGroupId, "6.0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, t, resourceName, &globalReplicationGroup),
					resource.TestCheckResourceAttrPair(resourceName, "engine_version_actual", primaryReplicationGroupResourceName, "engine_version_actual"),
					resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexache.MustCompile(`^6\.0\.[[:digit:]]+$`)),
				),
			},
			{
				Config: testAccGlobalReplicationGroupConfig_Redis_engineVersion(rName, primaryReplicationGroupId, "6.0", "6.2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, t, resourceName, &globalReplicationGroup),
					resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexache.MustCompile(`^6\.2\.[[:digit:]]+$`)),
				),
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_SetEngineVersionOnUpdate_MinorUpgrade_6x(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalReplicationGroup awstypes.GlobalReplicationGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	primaryReplicationGroupId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elasticache_global_replication_group.test"
	primaryReplicationGroupResourceName := "aws_elasticache_replication_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_Redis_engineVersionInherit(rName, primaryReplicationGroupId, "6.0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, t, resourceName, &globalReplicationGroup),
					resource.TestCheckResourceAttrPair(resourceName, "engine_version_actual", primaryReplicationGroupResourceName, "engine_version_actual"),
					resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexache.MustCompile(`^6\.0\.[[:digit:]]+$`)),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: testAccGlobalReplicationGroupConfig_Redis_engineVersion(rName, primaryReplicationGroupId, "6.0", "6.x"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionNoop),
					},
				},
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_SetEngineVersionOnUpdate_MinorDowngrade(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalReplicationGroup awstypes.GlobalReplicationGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	primaryReplicationGroupId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elasticache_global_replication_group.test"
	primaryReplicationGroupResourceName := "aws_elasticache_replication_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_Redis_engineVersionInherit(rName, primaryReplicationGroupId, "6.2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, t, resourceName, &globalReplicationGroup),
					resource.TestCheckResourceAttrPair(resourceName, "engine_version_actual", primaryReplicationGroupResourceName, "engine_version_actual"),
					resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexache.MustCompile(`^6\.2\.[[:digit:]]+$`)),
				),
			},
			{
				Config:      testAccGlobalReplicationGroupConfig_Redis_engineVersion(rName, primaryReplicationGroupId, "6.2", "6.0"),
				ExpectError: regexache.MustCompile(`Downgrading ElastiCache Global Replication Group \(.*\) engine version requires replacement`),
			},
			// This step fails with: Error running pre-apply refresh
			// {
			// 	Config: testAccGlobalReplicationGroupConfig_engineVersion(rName, primaryReplicationGroupId, "6.0", "6.0"),
			// 	Taint: []string{
			// 		resourceName,
			// 		primaryReplicationGroupResourceName,
			// 	},
			// 	Check: resource.ComposeAggregateTestCheckFunc(
			// 		testAccCheckGlobalReplicationGroupExists(resourceName, &globalReplicationGroup),
			// 		resource.TestCheckResourceAttrPair(resourceName, "engine_version_actual", primaryReplicationGroupResourceName, "engine_version_actual"),
			// 		resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexache.MustCompile(`^6\.0\.[[:digit:]]+$`)),
			// 	),
			// },
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_SetEngineVersionOnUpdate_MajorUpgrade(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalReplicationGroup awstypes.GlobalReplicationGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	primaryReplicationGroupId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elasticache_global_replication_group.test"
	primaryReplicationGroupResourceName := "aws_elasticache_replication_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_Redis_engineVersionInherit(rName, primaryReplicationGroupId, "5.0.6"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, t, resourceName, &globalReplicationGroup),
					resource.TestCheckResourceAttrPair(resourceName, "engine_version_actual", primaryReplicationGroupResourceName, "engine_version_actual"),
					resource.TestCheckResourceAttr(resourceName, "engine_version_actual", "5.0.6"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGlobalReplicationGroupConfig_engineVersionParam(rName, primaryReplicationGroupId, "5.0.6", "6.2", "default.redis6.x"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, t, resourceName, &globalReplicationGroup),
					resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexache.MustCompile(`^6\.2\.[[:digit:]]+$`)),
				),
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_SetEngineVersionOnUpdate_MajorUpgrade_6x(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalReplicationGroup awstypes.GlobalReplicationGroup
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	primaryReplicationGroupId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elasticache_global_replication_group.test"
	primaryReplicationGroupResourceName := "aws_elasticache_replication_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_Redis_engineVersionInherit(rName, primaryReplicationGroupId, "5.0.6"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, t, resourceName, &globalReplicationGroup),
					resource.TestCheckResourceAttrPair(resourceName, "engine_version_actual", primaryReplicationGroupResourceName, "engine_version_actual"),
					resource.TestCheckResourceAttr(resourceName, "engine_version_actual", "5.0.6"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccGlobalReplicationGroupConfig_engineVersionParam(rName, primaryReplicationGroupId, "5.0.6", "6.x", "default.redis6.x"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, t, resourceName, &globalReplicationGroup),
					resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexache.MustCompile(`^6\.2\.[[:digit:]]+$`)),
				),
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_SetParameterGroupOnUpdate_NoVersion(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalReplicationGroup awstypes.GlobalReplicationGroup

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	primaryReplicationGroupId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resourceName := "aws_elasticache_global_replication_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_Redis_engineVersionInherit(rName, primaryReplicationGroupId, "6.2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, t, resourceName, &globalReplicationGroup),
					resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexache.MustCompile(`^6\.2\.[[:digit:]]+$`)),
				),
			},
			{
				Config:      testAccGlobalReplicationGroupConfig_param(rName, primaryReplicationGroupId, "6.2", "default.redis6.x"),
				ExpectError: regexache.MustCompile(`cannot change parameter group name without upgrading major engine version`),
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_SetParameterGroupOnUpdate_MinorUpgrade(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalReplicationGroup awstypes.GlobalReplicationGroup

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	primaryReplicationGroupId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resourceName := "aws_elasticache_global_replication_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_Redis_engineVersionInherit(rName, primaryReplicationGroupId, "6.0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, t, resourceName, &globalReplicationGroup),
					resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexache.MustCompile(`^6\.0\.[[:digit:]]+$`)),
				),
			},
			{
				Config:      testAccGlobalReplicationGroupConfig_engineVersionParam(rName, primaryReplicationGroupId, "6.0", "6.2", "default.redis6.x"),
				ExpectError: regexache.MustCompile(`cannot change parameter group name on minor engine version upgrade, upgrading from 6\.0\.[[:digit:]]+ to 6\.2\.[[:digit:]]+`),
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_UpdateParameterGroupName(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalReplicationGroup awstypes.GlobalReplicationGroup

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	primaryReplicationGroupId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	parameterGroupName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	resourceName := "aws_elasticache_global_replication_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_engineVersionParam(rName, primaryReplicationGroupId, "5.0.6", "6.2", "default.redis6.x"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, t, resourceName, &globalReplicationGroup),
					resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexache.MustCompile(`^6\.2\.[[:digit:]]+$`)),
				),
			},
			{
				Config:      testAccGlobalReplicationGroupConfig_engineVersionCustomParam(rName, primaryReplicationGroupId, "5.0.6", "6.2", parameterGroupName, "redis6.x"),
				ExpectError: regexache.MustCompile(`cannot change parameter group name without upgrading major engine version`),
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_SetEngineOnCreate_ValkeyUpgrade(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalReplicationGroup awstypes.GlobalReplicationGroup

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	primaryReplicationGroupId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elasticache_global_replication_group.test"
	primaryReplicationGroupResourceName := "aws_elasticache_replication_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				// create global datastore with valkey 8.0 engine version from a redis 7.1 primary replication group
				Config: testAccGlobalReplicationGroupConfig_engineParam(rName, primaryReplicationGroupId, "redis", "7.1", "valkey", "8.0", "default.valkey8"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, t, resourceName, &globalReplicationGroup),
					resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexache.MustCompile(`^8\.0\.[[:digit:]]+$`)),
				),
			},
			{
				// check import of global datastore
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrParameterGroupName},
			},
			{
				// refresh primary replication group after being upgraded by the global datastore
				RefreshState: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(primaryReplicationGroupResourceName, names.AttrEngine, "valkey"),
					resource.TestCheckResourceAttr(primaryReplicationGroupResourceName, names.AttrEngineVersion, "8.0"),
					resource.TestMatchResourceAttr(primaryReplicationGroupResourceName, names.AttrParameterGroupName, regexache.MustCompile(`^global-datastore-.+$`)),
					resource.TestCheckResourceAttrPair(primaryReplicationGroupResourceName, "global_replication_group_id", resourceName, "global_replication_group_id"),
				),
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_SetEngineOnUpdate_ValkeyUpgrade(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalReplicationGroup awstypes.GlobalReplicationGroup

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	primaryReplicationGroupId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elasticache_global_replication_group.test"
	primaryReplicationGroupResourceName := "aws_elasticache_replication_group.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				// create global datastore using redis 7.1 primary replication group engine version
				Config: testAccGlobalReplicationGroupConfig_Redis_engineVersionInherit(rName, primaryReplicationGroupId, "7.1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, t, resourceName, &globalReplicationGroup),
					resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexache.MustCompile(`^7\.1\.[[:digit:]]+$`)),
				),
			},
			{
				// check import of global datastore
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrParameterGroupName},
			},
			{
				// refresh primary replication group after being associated to global datastore
				RefreshState: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(primaryReplicationGroupResourceName, names.AttrEngine, "redis"),
					resource.TestCheckResourceAttr(primaryReplicationGroupResourceName, names.AttrEngineVersion, "7.1"),
					resource.TestMatchResourceAttr(primaryReplicationGroupResourceName, names.AttrParameterGroupName, regexache.MustCompile(`^global-datastore-.+$`)),
					resource.TestCheckResourceAttrPair(primaryReplicationGroupResourceName, "global_replication_group_id", resourceName, "global_replication_group_id"),
				),
			},
			{
				// upgrade engine and version on global datastore
				Config: testAccGlobalReplicationGroupConfig_engineParam(rName, primaryReplicationGroupId, "redis", "7.1", "valkey", "7.2", "default.valkey7"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, t, resourceName, &globalReplicationGroup),
					resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexache.MustCompile(`^7\.2\.[[:digit:]]+$`)),
				),
			},
			{
				// check import of global datastore
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrParameterGroupName},
			},
			{
				// refresh primary replication group after being upgraded by the global datastore
				RefreshState: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(primaryReplicationGroupResourceName, names.AttrEngine, "valkey"),
					resource.TestCheckResourceAttr(primaryReplicationGroupResourceName, names.AttrEngineVersion, "7.2"),
					resource.TestMatchResourceAttr(primaryReplicationGroupResourceName, names.AttrParameterGroupName, regexache.MustCompile(`^global-datastore-.+$`)),
					resource.TestCheckResourceAttrPair(primaryReplicationGroupResourceName, "global_replication_group_id", resourceName, "global_replication_group_id"),
				),
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_InheritValkeyEngine_SecondaryReplicationGroup(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalReplicationGroup awstypes.GlobalReplicationGroup

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	primaryReplicationGroupId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	secondaryReplicationGroupId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elasticache_global_replication_group.test"
	primaryReplicationGroupResourceName := "aws_elasticache_replication_group.test"
	secondaryReplicationGroupResourceName := "aws_elasticache_replication_group.secondary"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
			testAccPreCheckGlobalReplicationGroup(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 2),
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				// create global datastore using Valkey 8.0 primary replication group and add secondary replication group
				Config: testAccGlobalReplicationGroupConfig_Valkey_inheritEngine_secondaryReplicationGroup(rName, primaryReplicationGroupId, secondaryReplicationGroupId),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, t, resourceName, &globalReplicationGroup),
					resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexache.MustCompile(`^8\.0\.[[:digit:]]+$`)),
				),
			},
			{
				// refresh replication groups to pick up all engine and version computed changes
				RefreshState: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(primaryReplicationGroupResourceName, names.AttrEngine, "valkey"),
					resource.TestCheckResourceAttr(primaryReplicationGroupResourceName, names.AttrEngineVersion, "8.0"),
					resource.TestMatchResourceAttr(primaryReplicationGroupResourceName, names.AttrParameterGroupName, regexache.MustCompile(`^global-datastore-.+$`)),
					resource.TestCheckResourceAttrPair(primaryReplicationGroupResourceName, "global_replication_group_id", resourceName, "global_replication_group_id"),

					resource.TestCheckResourceAttr(secondaryReplicationGroupResourceName, names.AttrEngine, "valkey"),
					resource.TestCheckResourceAttr(secondaryReplicationGroupResourceName, names.AttrEngineVersion, "8.0"),
					resource.TestMatchResourceAttr(secondaryReplicationGroupResourceName, names.AttrParameterGroupName, regexache.MustCompile(`^global-datastore-.+$`)),
					resource.TestCheckResourceAttrPair(secondaryReplicationGroupResourceName, "global_replication_group_id", resourceName, "global_replication_group_id"),
				),
			},
		},
	})
}

func testAccCheckGlobalReplicationGroupExists(ctx context.Context, t *testing.T, resourceName string, v *awstypes.GlobalReplicationGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ElastiCache Global Replication Group ID is set")
		}

		conn := acctest.ProviderMeta(ctx, t).ElastiCacheClient(ctx)
		grg, err := tfelasticache.FindGlobalReplicationGroupByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("retrieving ElastiCache Global Replication Group (%s): %w", rs.Primary.ID, err)
		}

		if aws.ToString(grg.Status) == "deleting" || aws.ToString(grg.Status) == "deleted" {
			return fmt.Errorf("ElastiCache Global Replication Group (%s) exists, but is in a non-available state: %s", rs.Primary.ID, aws.ToString(grg.Status))
		}

		*v = *grg

		return nil
	}
}

func testAccCheckGlobalReplicationGroupDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ElastiCacheClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_elasticache_global_replication_group" {
				continue
			}

			_, err := tfelasticache.FindGlobalReplicationGroupByID(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}
			return fmt.Errorf("ElastiCache Global Replication Group (%s) still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccPreCheckGlobalReplicationGroup(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).ElastiCacheClient(ctx)

	input := &elasticache.DescribeGlobalReplicationGroupsInput{}
	_, err := conn.DescribeGlobalReplicationGroups(ctx, input)

	if acctest.PreCheckSkipError(err) ||
		errs.IsAErrorMessageContains[*awstypes.InvalidParameterValueException](err, "Access Denied to API Version: APIGlobalDatastore") {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccMatchReplicationGroupActualVersion(ctx context.Context, t *testing.T, j *awstypes.ReplicationGroup, r *regexp.Regexp) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ElastiCacheClient(ctx)

		cacheCluster := j.NodeGroups[0].NodeGroupMembers[0]
		cluster, err := tfelasticache.FindCacheClusterByID(ctx, conn, aws.ToString(cacheCluster.CacheClusterId))
		if err != nil {
			return err
		}

		if !r.MatchString(aws.ToString(cluster.EngineVersion)) {
			return fmt.Errorf("Actual engine version didn't match %q, got %q", r.String(), aws.ToString(cluster.EngineVersion))
		}
		return nil
	}
}

func testAccGlobalReplicationGroupConfig_Redis_basic(rName, primaryReplicationGroupId string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_global_replication_group" "test" {
  global_replication_group_id_suffix = %[1]q
  primary_replication_group_id       = aws_elasticache_replication_group.test.id
}

resource "aws_elasticache_replication_group" "test" {
  replication_group_id = %[2]q
  description          = "test"

  engine             = "redis"
  engine_version     = "5.0.6"
  node_type          = "cache.m5.large"
  num_cache_clusters = 1
}
`, rName, primaryReplicationGroupId)
}

func testAccGlobalReplicationGroupConfig_Valkey_basic(rName, primaryReplicationGroupId string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_global_replication_group" "test" {
  global_replication_group_id_suffix = %[1]q
  primary_replication_group_id       = aws_elasticache_replication_group.test.id
}

resource "aws_elasticache_replication_group" "test" {
  replication_group_id = %[2]q
  description          = "test"

  engine             = "valkey"
  engine_version     = "7.2"
  node_type          = "cache.m5.large"
  num_cache_clusters = 1
}
`, rName, primaryReplicationGroupId)
}

func testAccGlobalReplicationGroupConfig_basic_nodeType(rName, primaryReplicationGroupId, nodeType string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_global_replication_group" "test" {
  global_replication_group_id_suffix = %[1]q
  primary_replication_group_id       = aws_elasticache_replication_group.test.id
}

resource "aws_elasticache_replication_group" "test" {
  replication_group_id = %[2]q
  description          = "test"

  engine             = "redis"
  engine_version     = "5.0.6"
  node_type          = %[3]q
  num_cache_clusters = 1
}
`, rName, primaryReplicationGroupId, nodeType)
}

func testAccGlobalReplicationGroupConfig_basic_automaticFailover(rName, primaryReplicationGroupId, automaticFailover string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_global_replication_group" "test" {
  global_replication_group_id_suffix = %[1]q
  primary_replication_group_id       = aws_elasticache_replication_group.test.id
}

resource "aws_elasticache_replication_group" "test" {
  replication_group_id = %[2]q
  description          = "test"

  node_type = "cache.m5.large"

  engine             = "redis"
  engine_version     = "5.0.6"
  num_cache_clusters = 2

  automatic_failover_enabled = %[3]s
}
`, rName, primaryReplicationGroupId, automaticFailover)
}

func testAccGlobalReplicationGroupConfig_description(rName, primaryReplicationGroupId, description string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_global_replication_group" "test" {
  global_replication_group_id_suffix = %[1]q
  primary_replication_group_id       = aws_elasticache_replication_group.test.id

  global_replication_group_description = %[3]q
}

resource "aws_elasticache_replication_group" "test" {
  replication_group_id = %[2]q
  description          = "test"

  engine             = "redis"
  engine_version     = "5.0.6"
  node_type          = "cache.m5.large"
  num_cache_clusters = 1
}
`, rName, primaryReplicationGroupId, description)
}

func testAccGlobalReplicationGroupConfig_nodeType_createNoChange(rName, primaryReplicationGroupId, nodeType string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_global_replication_group" "test" {
  global_replication_group_id_suffix = %[1]q
  primary_replication_group_id       = aws_elasticache_replication_group.test.id
  cache_node_type                    = %[3]q
}

resource "aws_elasticache_replication_group" "test" {
  replication_group_id = %[2]q
  description          = "test"

  engine             = "redis"
  engine_version     = "5.0.6"
  node_type          = %[3]q
  num_cache_clusters = 1
}
`, rName, primaryReplicationGroupId, nodeType)
}

func testAccGlobalReplicationGroupConfig_nodeType_createWithChange(rName, primaryReplicationGroupId, nodeType, globalNodeType string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_global_replication_group" "test" {
  global_replication_group_id_suffix = %[1]q
  primary_replication_group_id       = aws_elasticache_replication_group.test.id
  cache_node_type                    = %[4]q
}

resource "aws_elasticache_replication_group" "test" {
  replication_group_id = %[2]q
  description          = "test"

  engine             = "redis"
  engine_version     = "5.0.6"
  node_type          = %[3]q
  num_cache_clusters = 1

  lifecycle {
    ignore_changes = [node_type]
  }
}
`, rName, primaryReplicationGroupId, nodeType, globalNodeType)
}

func testAccGlobalReplicationGroupConfig_nodeType_update(rName, primaryReplicationGroupId, nodeType string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_global_replication_group" "test" {
  global_replication_group_id_suffix = %[1]q
  primary_replication_group_id       = aws_elasticache_replication_group.test.id
  cache_node_type                    = %[3]q
}

resource "aws_elasticache_replication_group" "test" {
  replication_group_id = %[2]q
  description          = "test"

  engine             = "redis"
  engine_version     = "5.0.6"
  num_cache_clusters = 1
}
`, rName, primaryReplicationGroupId, nodeType)
}

func testAccGlobalReplicationGroupConfig_automaticFailover_createNoChange(rName, primaryReplicationGroupId, automaticFailover string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_global_replication_group" "test" {
  global_replication_group_id_suffix = %[1]q
  primary_replication_group_id       = aws_elasticache_replication_group.test.id

  automatic_failover_enabled = %[3]s
}

resource "aws_elasticache_replication_group" "test" {
  replication_group_id = %[2]q
  description          = "test"

  node_type = "cache.m5.large"

  engine             = "redis"
  engine_version     = "5.0.6"
  num_cache_clusters = 2

  automatic_failover_enabled = %[3]s
}
`, rName, primaryReplicationGroupId, automaticFailover)
}

func testAccGlobalReplicationGroupConfig_automaticFailover_createWithChange(rName, primaryReplicationGroupId, automaticFailover, globalAutomaticFailover string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_global_replication_group" "test" {
  global_replication_group_id_suffix = %[1]q
  primary_replication_group_id       = aws_elasticache_replication_group.test.id

  automatic_failover_enabled = %[4]s
}

resource "aws_elasticache_replication_group" "test" {
  replication_group_id = %[2]q
  description          = "test"

  node_type = "cache.m5.large"

  engine             = "redis"
  engine_version     = "5.0.6"
  num_cache_clusters = 2

  automatic_failover_enabled = %[3]s

  lifecycle {
    ignore_changes = [automatic_failover_enabled]
  }
}
`, rName, primaryReplicationGroupId, automaticFailover, globalAutomaticFailover)
}

func testAccGlobalReplicationGroupConfig_automaticFailover_update(rName, primaryReplicationGroupId, globalAutomaticFailover string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_global_replication_group" "test" {
  global_replication_group_id_suffix = %[1]q
  primary_replication_group_id       = aws_elasticache_replication_group.test.id

  automatic_failover_enabled = %[3]s
}

resource "aws_elasticache_replication_group" "test" {
  replication_group_id = %[2]q
  description          = "test"

  engine             = "redis"
  engine_version     = "5.0.6"
  num_cache_clusters = 2

  lifecycle {
    ignore_changes = [automatic_failover_enabled]
  }
}
`, rName, primaryReplicationGroupId, globalAutomaticFailover)
}

func testAccGlobalReplicationGroupConfig_multipleSecondaries(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(3),
		testAccVPCBaseWithProvider(rName, "primary", acctest.ProviderName, 1),
		testAccVPCBaseWithProvider(rName, "alternate", acctest.ProviderNameAlternate, 1),
		testAccVPCBaseWithProvider(rName, "third", acctest.ProviderNameThird, 1),
		fmt.Sprintf(`
resource "aws_elasticache_global_replication_group" "test" {
  provider = aws

  global_replication_group_id_suffix = %[1]q
  primary_replication_group_id       = aws_elasticache_replication_group.primary.id
}

resource "aws_elasticache_replication_group" "primary" {
  provider = aws

  replication_group_id = "%[1]s-p"
  description          = "primary"

  subnet_group_name = aws_elasticache_subnet_group.primary.name

  node_type = "cache.m5.large"

  engine             = "redis"
  engine_version     = "5.0.6"
  num_cache_clusters = 1
}

resource "aws_elasticache_replication_group" "alternate" {
  provider = awsalternate

  replication_group_id        = "%[1]s-a"
  description                 = "alternate"
  global_replication_group_id = aws_elasticache_global_replication_group.test.global_replication_group_id

  subnet_group_name = aws_elasticache_subnet_group.alternate.name

  num_cache_clusters = 1
}

resource "aws_elasticache_replication_group" "third" {
  provider = awsthird

  replication_group_id        = "%[1]s-t"
  description                 = "third"
  global_replication_group_id = aws_elasticache_global_replication_group.test.global_replication_group_id

  subnet_group_name = aws_elasticache_subnet_group.third.name

  num_cache_clusters = 1
}
`, rName))
}

func testAccGlobalReplicationGroupConfig_replaceSecondaryDifferentRegionSetup(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(3),
		testAccVPCBaseWithProvider(rName, "primary", acctest.ProviderName, 1),
		testAccVPCBaseWithProvider(rName, "secondary", acctest.ProviderNameAlternate, 1),
		testAccVPCBaseWithProvider(rName, "third", acctest.ProviderNameThird, 1),
		fmt.Sprintf(`
resource "aws_elasticache_global_replication_group" "test" {
  provider = aws

  global_replication_group_id_suffix = %[1]q
  primary_replication_group_id       = aws_elasticache_replication_group.primary.id
}

resource "aws_elasticache_replication_group" "primary" {
  provider = aws

  replication_group_id = "%[1]s-p"
  description          = "primary"

  subnet_group_name = aws_elasticache_subnet_group.primary.name

  node_type = "cache.m5.large"

  engine             = "redis"
  engine_version     = "5.0.6"
  num_cache_clusters = 1
}

resource "aws_elasticache_replication_group" "secondary" {
  provider = awsalternate

  replication_group_id        = "%[1]s-a"
  description                 = "alternate"
  global_replication_group_id = aws_elasticache_global_replication_group.test.global_replication_group_id

  subnet_group_name = aws_elasticache_subnet_group.secondary.name

  num_cache_clusters = 1
}
`, rName))
}

func testAccGlobalReplicationGroupConfig_replaceSecondaryDifferentRegionMove(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(3),
		testAccVPCBaseWithProvider(rName, "primary", acctest.ProviderName, 1),
		testAccVPCBaseWithProvider(rName, "secondary", acctest.ProviderNameAlternate, 1),
		testAccVPCBaseWithProvider(rName, "third", acctest.ProviderNameThird, 1),
		fmt.Sprintf(`
resource "aws_elasticache_global_replication_group" "test" {
  provider = aws

  global_replication_group_id_suffix = %[1]q
  primary_replication_group_id       = aws_elasticache_replication_group.primary.id
}

resource "aws_elasticache_replication_group" "primary" {
  provider = aws

  replication_group_id = "%[1]s-p"
  description          = "primary"

  subnet_group_name = aws_elasticache_subnet_group.primary.name

  node_type = "cache.m5.large"

  engine             = "redis"
  engine_version     = "5.0.6"
  num_cache_clusters = 1
}

resource "aws_elasticache_replication_group" "third" {
  provider = awsthird

  replication_group_id        = "%[1]s-t"
  description                 = "third"
  global_replication_group_id = aws_elasticache_global_replication_group.test.global_replication_group_id

  subnet_group_name = aws_elasticache_subnet_group.third.name

  num_cache_clusters = 1
}
`, rName))
}

func testAccGlobalReplicationGroupConfig_clusterMode(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_global_replication_group" "test" {
  global_replication_group_id_suffix = %[1]q
  primary_replication_group_id       = aws_elasticache_replication_group.test.id
}

resource "aws_elasticache_replication_group" "test" {
  replication_group_id = %[1]q
  description          = "test"

  engine         = "redis"
  engine_version = "6.2"
  node_type      = "cache.m5.large"

  parameter_group_name       = "default.redis6.x.cluster.on"
  automatic_failover_enabled = true
  num_node_groups            = 2
  replicas_per_node_group    = 1
}
`, rName)
}

func testAccGlobalReplicationGroupConfig_numNodeGroups_inherit(rName string, numNodeGroups int) string {
	return fmt.Sprintf(`
resource "aws_elasticache_global_replication_group" "test" {
  global_replication_group_id_suffix = %[1]q
  primary_replication_group_id       = aws_elasticache_replication_group.test.id
}

resource "aws_elasticache_replication_group" "test" {
  replication_group_id = %[1]q
  description          = "test"

  engine         = "redis"
  engine_version = "6.2"
  node_type      = "cache.m5.large"

  parameter_group_name       = "default.redis6.x.cluster.on"
  automatic_failover_enabled = true
  num_node_groups            = %[2]d
  replicas_per_node_group    = 1

  lifecycle {
    ignore_changes = [member_clusters, num_node_groups]
  }
}
`, rName, numNodeGroups)
}

func testAccGlobalReplicationGroupConfig_numNodeGroups(rName string, numNodeGroups, globalNumNodeGroups int) string {
	return fmt.Sprintf(`
resource "aws_elasticache_global_replication_group" "test" {
  global_replication_group_id_suffix = %[1]q
  primary_replication_group_id       = aws_elasticache_replication_group.test.id

  num_node_groups = %[3]d
}

resource "aws_elasticache_replication_group" "test" {
  replication_group_id = %[1]q
  description          = "test"

  engine         = "redis"
  engine_version = "6.2"
  node_type      = "cache.m5.large"

  parameter_group_name       = "default.redis6.x.cluster.on"
  automatic_failover_enabled = true
  num_node_groups            = %[2]d
  replicas_per_node_group    = 1

  lifecycle {
    ignore_changes = [member_clusters, num_node_groups]
  }
}
`, rName, numNodeGroups, globalNumNodeGroups)
}

func testAccGlobalReplicationGroupConfig_Redis_engineVersionInherit(rName, primaryReplicationGroupId, repGroupEngineVersion string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_global_replication_group" "test" {
  global_replication_group_id_suffix = %[1]q
  primary_replication_group_id       = aws_elasticache_replication_group.test.id
}

resource "aws_elasticache_replication_group" "test" {
  replication_group_id = %[2]q
  description          = "test"

  engine             = "redis"
  engine_version     = %[3]q
  node_type          = "cache.m5.large"
  num_cache_clusters = 1
}
`, rName, primaryReplicationGroupId, repGroupEngineVersion)
}

func testAccGlobalReplicationGroupConfig_Redis_engineVersion(rName, primaryReplicationGroupId, repGroupEngineVersion, globalEngineVersion string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_global_replication_group" "test" {
  global_replication_group_id_suffix = %[1]q
  primary_replication_group_id       = aws_elasticache_replication_group.test.id

  engine_version = %[4]q
}

resource "aws_elasticache_replication_group" "test" {
  replication_group_id = %[2]q
  description          = "test"

  engine             = "redis"
  engine_version     = %[3]q
  node_type          = "cache.m5.large"
  num_cache_clusters = 1
}
`, rName, primaryReplicationGroupId, repGroupEngineVersion, globalEngineVersion)
}

func testAccGlobalReplicationGroupConfig_Valkey_engineVersion(rName, primaryReplicationGroupId, repGroupEngineVersion, globalEngineVersion string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_global_replication_group" "test" {
  global_replication_group_id_suffix = %[1]q
  primary_replication_group_id       = aws_elasticache_replication_group.test.id

  engine_version = %[4]q
}

resource "aws_elasticache_replication_group" "test" {
  replication_group_id = %[2]q
  description          = "test"

  engine             = "valkey"
  engine_version     = %[3]q
  node_type          = "cache.m5.large"
  num_cache_clusters = 1
}
`, rName, primaryReplicationGroupId, repGroupEngineVersion, globalEngineVersion)
}

func testAccGlobalReplicationGroupConfig_engineVersionParam(rName, primaryReplicationGroupId, repGroupEngineVersion, globalEngineVersion, parameterGroup string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_global_replication_group" "test" {
  global_replication_group_id_suffix = %[1]q
  primary_replication_group_id       = aws_elasticache_replication_group.test.id

  engine_version       = %[4]q
  parameter_group_name = %[5]q
}

resource "aws_elasticache_replication_group" "test" {
  replication_group_id = %[2]q
  description          = "test"

  engine             = "redis"
  engine_version     = %[3]q
  node_type          = "cache.m5.large"
  num_cache_clusters = 1
}
`, rName, primaryReplicationGroupId, repGroupEngineVersion, globalEngineVersion, parameterGroup)
}

func testAccGlobalReplicationGroupConfig_engineParam(rName, primaryReplicationGroupId, repGroupEngine, repGroupEngineVersion, globalEngine, globalEngineVersion, globalParamGroup string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_global_replication_group" "test" {
  global_replication_group_id_suffix = %[1]q
  primary_replication_group_id       = aws_elasticache_replication_group.test.id

  engine               = %[5]q
  engine_version       = %[6]q
  parameter_group_name = %[7]q
}

resource "aws_elasticache_replication_group" "test" {
  replication_group_id = %[2]q
  description          = "test"
  engine               = %[3]q
  engine_version       = %[4]q
  node_type            = "cache.m5.large"
  num_cache_clusters   = 1
}
`, rName, primaryReplicationGroupId, repGroupEngine, repGroupEngineVersion, globalEngine, globalEngineVersion, globalParamGroup)
}

func testAccGlobalReplicationGroupConfig_Valkey_inheritEngine_secondaryReplicationGroup(rName, primaryReplicationGroupId, secondaryReplicationGroupId string) string {
	return acctest.ConfigCompose(acctest.ConfigMultipleRegionProvider(2), fmt.Sprintf(`
resource "aws_elasticache_global_replication_group" "test" {
  global_replication_group_id_suffix = %[1]q
  primary_replication_group_id       = aws_elasticache_replication_group.test.id
}

resource "aws_elasticache_replication_group" "test" {
  replication_group_id = %[2]q
  description          = "test"
  engine               = "valkey"
  engine_version       = "8.0"
  node_type            = "cache.m5.large"
  num_cache_clusters   = 1
}

resource "aws_elasticache_replication_group" "secondary" {
  provider = awsalternate

  replication_group_id        = %[3]q
  description                 = "test secondary"
  global_replication_group_id = aws_elasticache_global_replication_group.test.id
}
`, rName, primaryReplicationGroupId, secondaryReplicationGroupId))
}

func testAccGlobalReplicationGroupConfig_engineVersionCustomParam(rName, primaryReplicationGroupId, repGroupEngineVersion, globalEngineVersion, parameterGroupName, parameterGroupFamily string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_global_replication_group" "test" {
  global_replication_group_id_suffix = %[1]q
  primary_replication_group_id       = aws_elasticache_replication_group.test.id

  engine_version       = %[4]q
  parameter_group_name = aws_elasticache_parameter_group.test.name
}

resource "aws_elasticache_replication_group" "test" {
  replication_group_id = %[2]q
  description          = "test"

  engine             = "redis"
  engine_version     = %[3]q
  node_type          = "cache.m5.large"
  num_cache_clusters = 1
}

resource "aws_elasticache_parameter_group" "test" {
  name        = %[5]q
  description = "test"
  family      = %[6]q
}
`, rName, primaryReplicationGroupId, repGroupEngineVersion, globalEngineVersion, parameterGroupName, parameterGroupFamily)
}

func testAccGlobalReplicationGroupConfig_param(rName, primaryReplicationGroupId, repGroupEngineVersion, parameterGroup string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_global_replication_group" "test" {
  global_replication_group_id_suffix = %[1]q
  primary_replication_group_id       = aws_elasticache_replication_group.test.id

  parameter_group_name = %[4]q
}

resource "aws_elasticache_replication_group" "test" {
  replication_group_id = %[2]q
  description          = "test"

  engine             = "redis"
  engine_version     = %[3]q
  node_type          = "cache.m5.large"
  num_cache_clusters = 1
}
`, rName, primaryReplicationGroupId, repGroupEngineVersion, parameterGroup)
}

func testAccVPCBaseWithProvider(rName, name, provider string, subnetCount int) string {
	return acctest.ConfigCompose(
		testAccAvailableAZsNoOptInConfigWithProvider(name, provider),
		fmt.Sprintf(`
resource "aws_vpc" "%[1]s" {
  provider = %[2]s

  cidr_block = "192.168.0.0/16"
}

resource "aws_subnet" "%[1]s" {
  provider = %[2]s

  count = %[4]d

  vpc_id            = aws_vpc.%[1]s.id
  cidr_block        = "192.168.${count.index}.0/24"
  availability_zone = data.aws_availability_zones.%[1]s.names[count.index]
}

resource "aws_elasticache_subnet_group" "%[1]s" {
  provider = %[2]s

  name       = %[3]q
  subnet_ids = aws_subnet.%[1]s[*].id
}
`, name, provider, rName, subnetCount),
	)
}

func testAccAvailableAZsNoOptInConfigWithProvider(name, provider string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "%[1]s" {
  provider = %[2]s

  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}
`, name, provider)
}
