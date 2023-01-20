package elasticache_test

import (
	"context"
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfelasticache "github.com/hashicorp/terraform-provider-aws/internal/service/elasticache"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccElastiCacheGlobalReplicationGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalReplicationGroup elasticache.GlobalReplicationGroup
	var primaryReplicationGroup elasticache.ReplicationGroup
	var pg elasticache.CacheParameterGroup

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	primaryReplicationGroupId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_elasticache_global_replication_group.test"
	primaryReplicationGroupResourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckGlobalReplicationGroupDestroy(ctx),
			testAccCheckGlobalReplicationGroupMemberParameterGroupDestroy(ctx, &pg),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_basic(rName, primaryReplicationGroupId),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, resourceName, &globalReplicationGroup),
					testAccCheckReplicationGroupExists(ctx, primaryReplicationGroupResourceName, &primaryReplicationGroup),
					testAccCheckReplicationGroupParameterGroup(ctx, &primaryReplicationGroup, &pg),
					acctest.MatchResourceAttrGlobalARN(resourceName, "arn", "elasticache", regexp.MustCompile(`globalreplicationgroup:`+tfelasticache.GlobalReplicationGroupRegionPrefixFormat+rName)),
					resource.TestCheckResourceAttrPair(resourceName, "at_rest_encryption_enabled", primaryReplicationGroupResourceName, "at_rest_encryption_enabled"),
					resource.TestCheckResourceAttr(resourceName, "auth_token_enabled", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "automatic_failover_enabled", primaryReplicationGroupResourceName, "automatic_failover_enabled"),
					resource.TestCheckResourceAttrPair(resourceName, "cache_node_type", primaryReplicationGroupResourceName, "node_type"),
					resource.TestCheckResourceAttrPair(resourceName, "cluster_enabled", primaryReplicationGroupResourceName, "cluster_enabled"),
					resource.TestCheckResourceAttrPair(resourceName, "engine", primaryReplicationGroupResourceName, "engine"),
					resource.TestCheckResourceAttrPair(resourceName, "engine_version_actual", primaryReplicationGroupResourceName, "engine_version_actual"),
					resource.TestCheckResourceAttr(resourceName, "global_replication_group_id_suffix", rName),
					resource.TestMatchResourceAttr(resourceName, "global_replication_group_id", regexp.MustCompile(tfelasticache.GlobalReplicationGroupRegionPrefixFormat+rName)),
					resource.TestCheckResourceAttr(resourceName, "global_replication_group_description", tfelasticache.EmptyDescription),
					resource.TestCheckResourceAttr(resourceName, "global_node_groups.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "num_node_groups", "0"),
					resource.TestCheckResourceAttr(resourceName, "primary_replication_group_id", primaryReplicationGroupId),
					resource.TestCheckResourceAttr(resourceName, "transit_encryption_enabled", "false"),
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

	var globalReplicationGroup elasticache.GlobalReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	primaryReplicationGroupId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_global_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_basic(rName, primaryReplicationGroupId),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, resourceName, &globalReplicationGroup),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfelasticache.ResourceGlobalReplicationGroup(), resourceName),
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

	var globalReplicationGroup elasticache.GlobalReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	primaryReplicationGroupId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	description1 := sdkacctest.RandString(10)
	description2 := sdkacctest.RandString(10)
	resourceName := "aws_elasticache_global_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_description(rName, primaryReplicationGroupId, description1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, resourceName, &globalReplicationGroup),
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
					testAccCheckGlobalReplicationGroupExists(ctx, resourceName, &globalReplicationGroup),
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

	var globalReplicationGroup elasticache.GlobalReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	primaryReplicationGroupId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	nodeType := "cache.m5.large"
	resourceName := "aws_elasticache_global_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_nodeType_createNoChange(rName, primaryReplicationGroupId, nodeType),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, resourceName, &globalReplicationGroup),
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

	var globalReplicationGroup elasticache.GlobalReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	primaryReplicationGroupId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	nodeType := "cache.m5.large"
	globalNodeType := "cache.m5.xlarge"
	resourceName := "aws_elasticache_global_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_nodeType_createWithChange(rName, primaryReplicationGroupId, nodeType, globalNodeType),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, resourceName, &globalReplicationGroup),
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

	var globalReplicationGroup elasticache.GlobalReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	primaryReplicationGroupId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	nodeType := "cache.m5.large"
	resourceName := "aws_elasticache_global_replication_group.test"
	primaryReplicationGroupResourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_basic_nodeType(rName, primaryReplicationGroupId, nodeType),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, resourceName, &globalReplicationGroup),
					resource.TestCheckResourceAttrPair(resourceName, "cache_node_type", primaryReplicationGroupResourceName, "node_type"),
				),
			},
			{
				Config: testAccGlobalReplicationGroupConfig_nodeType_createNoChange(rName, primaryReplicationGroupId, nodeType),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, resourceName, &globalReplicationGroup),
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

	var globalReplicationGroup elasticache.GlobalReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	primaryReplicationGroupId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	nodeType := "cache.m5.large"
	updatedNodeType := "cache.m5.xlarge"
	resourceName := "aws_elasticache_global_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_nodeType_createNoChange(rName, primaryReplicationGroupId, nodeType),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, resourceName, &globalReplicationGroup),
					resource.TestCheckResourceAttr(resourceName, "cache_node_type", nodeType),
				),
			},
			{
				Config: testAccGlobalReplicationGroupConfig_nodeType_update(rName, primaryReplicationGroupId, updatedNodeType),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, resourceName, &globalReplicationGroup),
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

	var globalReplicationGroup elasticache.GlobalReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	enabled := "true"
	primaryReplicationGroupId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_global_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_automaticFailover_createNoChange(rName, primaryReplicationGroupId, enabled),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, resourceName, &globalReplicationGroup),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", enabled),
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

	var globalReplicationGroup elasticache.GlobalReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	enabled := "false"
	globalEnabled := "true"
	primaryReplicationGroupId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_global_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_automaticFailover_createWithChange(rName, primaryReplicationGroupId, enabled, globalEnabled),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, resourceName, &globalReplicationGroup),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", globalEnabled),
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

	var globalReplicationGroup elasticache.GlobalReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	primaryReplicationGroupId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	enabled := "false"
	resourceName := "aws_elasticache_global_replication_group.test"
	primaryReplicationGroupResourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_basic_automaticFailover(rName, primaryReplicationGroupId, enabled),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, resourceName, &globalReplicationGroup),
					resource.TestCheckResourceAttrPair(resourceName, "automatic_failover_enabled", primaryReplicationGroupResourceName, "automatic_failover_enabled"),
				),
			},
			{
				Config: testAccGlobalReplicationGroupConfig_automaticFailover_createNoChange(rName, primaryReplicationGroupId, enabled),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, resourceName, &globalReplicationGroup),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", enabled),
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

	var globalReplicationGroup elasticache.GlobalReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	primaryReplicationGroupId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	enabled := "true"
	updatedEnabled := "false"
	resourceName := "aws_elasticache_global_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_automaticFailover_createNoChange(rName, primaryReplicationGroupId, enabled),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, resourceName, &globalReplicationGroup),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", enabled),
				),
			},
			{
				Config: testAccGlobalReplicationGroupConfig_automaticFailover_update(rName, primaryReplicationGroupId, updatedEnabled),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, resourceName, &globalReplicationGroup),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", updatedEnabled),
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

	var globalReplcationGroup elasticache.GlobalReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_global_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 3)
		},
		ErrorCheck:               acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(t, 3),
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_multipleSecondaries(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, resourceName, &globalReplcationGroup),
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

	var globalReplcationGroup elasticache.GlobalReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_global_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckMultipleRegion(t, 3)
		},
		ErrorCheck:               acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(t, 3),
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_replaceSecondaryDifferentRegionSetup(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, resourceName, &globalReplcationGroup),
				),
			},
			{
				Config: testAccGlobalReplicationGroupConfig_replaceSecondaryDifferentRegionMove(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, resourceName, &globalReplcationGroup),
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

	var globalReplicationGroup elasticache.GlobalReplicationGroup
	var primaryReplicationGroup elasticache.ReplicationGroup

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_elasticache_global_replication_group.test"
	primaryReplicationGroupResourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_clusterMode(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, resourceName, &globalReplicationGroup),
					testAccCheckReplicationGroupExists(ctx, primaryReplicationGroupResourceName, &primaryReplicationGroup),
					resource.TestCheckResourceAttrPair(resourceName, "automatic_failover_enabled", primaryReplicationGroupResourceName, "automatic_failover_enabled"),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", "true"),
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

	var globalReplicationGroup elasticache.GlobalReplicationGroup
	var primaryReplicationGroup elasticache.ReplicationGroup

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_elasticache_global_replication_group.test"
	primaryReplicationGroupResourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_numNodeGroups(rName, 2, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, resourceName, &globalReplicationGroup),
					testAccCheckReplicationGroupExists(ctx, primaryReplicationGroupResourceName, &primaryReplicationGroup),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", "true"),
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

	var globalReplicationGroup elasticache.GlobalReplicationGroup
	var primaryReplicationGroup elasticache.ReplicationGroup

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_elasticache_global_replication_group.test"
	primaryReplicationGroupResourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_numNodeGroups(rName, 2, 3),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, resourceName, &globalReplicationGroup),
					testAccCheckReplicationGroupExists(ctx, primaryReplicationGroupResourceName, &primaryReplicationGroup),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "global_node_groups.#", "3"),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "global_node_groups.*", map[string]*regexp.Regexp{
						"global_node_group_id": regexp.MustCompile(fmt.Sprintf("^[a-z]+-%s-0001$", rName)),
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "global_node_groups.*", map[string]*regexp.Regexp{
						"global_node_group_id": regexp.MustCompile(fmt.Sprintf("^[a-z]+-%s-0002$", rName)),
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "global_node_groups.*", map[string]*regexp.Regexp{
						"global_node_group_id": regexp.MustCompile(fmt.Sprintf("^[a-z]+-%s-0003$", rName)),
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

	var globalReplicationGroup elasticache.GlobalReplicationGroup
	var primaryReplicationGroup elasticache.ReplicationGroup

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_elasticache_global_replication_group.test"
	primaryReplicationGroupResourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_numNodeGroups(rName, 3, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, resourceName, &globalReplicationGroup),
					testAccCheckReplicationGroupExists(ctx, primaryReplicationGroupResourceName, &primaryReplicationGroup),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "global_node_groups.#", "1"),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "global_node_groups.*", map[string]*regexp.Regexp{
						"global_node_group_id": regexp.MustCompile(fmt.Sprintf("^[a-z]+-%s-0001$", rName)),
						"slots":                regexp.MustCompile("^0-16383$"), // all slots
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

	var globalReplicationGroup elasticache.GlobalReplicationGroup
	var primaryReplicationGroup elasticache.ReplicationGroup

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_elasticache_global_replication_group.test"
	primaryReplicationGroupResourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_numNodeGroups_inherit(rName, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, resourceName, &globalReplicationGroup),
					testAccCheckReplicationGroupExists(ctx, primaryReplicationGroupResourceName, &primaryReplicationGroup),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "global_node_groups.#", "2"),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "global_node_groups.*", map[string]*regexp.Regexp{
						"global_node_group_id": regexp.MustCompile(fmt.Sprintf("^[a-z]+-%s-0001$", rName)),
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "global_node_groups.*", map[string]*regexp.Regexp{
						"global_node_group_id": regexp.MustCompile(fmt.Sprintf("^[a-z]+-%s-0002$", rName)),
					}),
					resource.TestCheckResourceAttr(resourceName, "num_node_groups", "2"),
					resource.TestCheckResourceAttr(primaryReplicationGroupResourceName, "num_node_groups", "2"),
				),
			},
			{
				Config: testAccGlobalReplicationGroupConfig_numNodeGroups(rName, 2, 3),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, resourceName, &globalReplicationGroup),
					testAccCheckReplicationGroupExists(ctx, primaryReplicationGroupResourceName, &primaryReplicationGroup),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "global_node_groups.#", "3"),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "global_node_groups.*", map[string]*regexp.Regexp{
						"global_node_group_id": regexp.MustCompile(fmt.Sprintf("^[a-z]+-%s-0001$", rName)),
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "global_node_groups.*", map[string]*regexp.Regexp{
						"global_node_group_id": regexp.MustCompile(fmt.Sprintf("^[a-z]+-%s-0002$", rName)),
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "global_node_groups.*", map[string]*regexp.Regexp{
						"global_node_group_id": regexp.MustCompile(fmt.Sprintf("^[a-z]+-%s-0003$", rName)),
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

	var globalReplicationGroup elasticache.GlobalReplicationGroup
	var primaryReplicationGroup elasticache.ReplicationGroup

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_elasticache_global_replication_group.test"
	primaryReplicationGroupResourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_numNodeGroups_inherit(rName, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, resourceName, &globalReplicationGroup),
					testAccCheckReplicationGroupExists(ctx, primaryReplicationGroupResourceName, &primaryReplicationGroup),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "global_node_groups.#", "2"),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "global_node_groups.*", map[string]*regexp.Regexp{
						"global_node_group_id": regexp.MustCompile(fmt.Sprintf("^[a-z]+-%s-0001$", rName)),
					}),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "global_node_groups.*", map[string]*regexp.Regexp{
						"global_node_group_id": regexp.MustCompile(fmt.Sprintf("^[a-z]+-%s-0002$", rName)),
					}),
					resource.TestCheckResourceAttr(resourceName, "num_node_groups", "2"),
					resource.TestCheckResourceAttr(primaryReplicationGroupResourceName, "num_node_groups", "2"),
				),
			},
			{
				Config: testAccGlobalReplicationGroupConfig_numNodeGroups(rName, 2, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, resourceName, &globalReplicationGroup),
					testAccCheckReplicationGroupExists(ctx, primaryReplicationGroupResourceName, &primaryReplicationGroup),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "global_node_groups.#", "1"),
					resource.TestMatchTypeSetElemNestedAttrs(resourceName, "global_node_groups.*", map[string]*regexp.Regexp{
						"global_node_group_id": regexp.MustCompile(fmt.Sprintf("^[a-z]+-%s-0001$", rName)),
						"slots":                regexp.MustCompile("^0-16383$"), // all slots
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

func TestAccElastiCacheGlobalReplicationGroup_SetEngineVersionOnCreate_NoChange_v6(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalReplicationGroup elasticache.GlobalReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	primaryReplicationGroupId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_global_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_engineVersion(rName, primaryReplicationGroupId, "6.2", "6.2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, resourceName, &globalReplicationGroup),
					resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexp.MustCompile(`^6\.2\.[[:digit:]]+$`)),
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

func TestAccElastiCacheGlobalReplicationGroup_SetEngineVersionOnCreate_NoChange_v6x(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalReplicationGroup elasticache.GlobalReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	primaryReplicationGroupId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_global_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_engineVersion(rName, primaryReplicationGroupId, "6.2", "6.x"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, resourceName, &globalReplicationGroup),
					resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexp.MustCompile(`^6\.2\.[[:digit:]]+$`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"engine_version"},
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_SetEngineVersionOnCreate_NoChange_v5(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalReplicationGroup elasticache.GlobalReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	primaryReplicationGroupId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_global_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_engineVersion(rName, primaryReplicationGroupId, "5.0.6", "5.0.6"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, resourceName, &globalReplicationGroup),
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

func TestAccElastiCacheGlobalReplicationGroup_SetEngineVersionOnCreate_MinorUpgrade(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalReplicationGroup elasticache.GlobalReplicationGroup
	var rg elasticache.ReplicationGroup

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	primaryReplicationGroupId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_global_replication_group.test"
	primaryReplicationGroupResourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_engineVersion(rName, primaryReplicationGroupId, "6.0", "6.2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, resourceName, &globalReplicationGroup),
					testAccCheckReplicationGroupExists(ctx, primaryReplicationGroupResourceName, &rg),
					resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexp.MustCompile(`^6\.2\.[[:digit:]]+$`)),
					testAccMatchReplicationGroupActualVersion(ctx, &rg, regexp.MustCompile(`^6\.2\.[[:digit:]]+$`)),
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

	var globalReplicationGroup elasticache.GlobalReplicationGroup
	var rg elasticache.ReplicationGroup

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	primaryReplicationGroupId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_global_replication_group.test"
	primaryReplicationGroupResourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_engineVersion(rName, primaryReplicationGroupId, "6.0", "6.x"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, resourceName, &globalReplicationGroup),
					testAccCheckReplicationGroupExists(ctx, primaryReplicationGroupResourceName, &rg),
					resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexp.MustCompile(`^6\.0\.[[:digit:]]+$`)),
					testAccMatchReplicationGroupActualVersion(ctx, &rg, regexp.MustCompile(`^6\.0\.[[:digit:]]+$`)),
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

	var globalReplicationGroup elasticache.GlobalReplicationGroup
	var rg elasticache.ReplicationGroup

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	primaryReplicationGroupId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_global_replication_group.test"
	primaryReplicationGroupResourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_engineVersionParam(rName, primaryReplicationGroupId, "5.0.6", "6.2", "default.redis6.x"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, resourceName, &globalReplicationGroup),
					testAccCheckReplicationGroupExists(ctx, primaryReplicationGroupResourceName, &rg),
					resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexp.MustCompile(`^6\.2\.[[:digit:]]+$`)),
					testAccMatchReplicationGroupActualVersion(ctx, &rg, regexp.MustCompile(`^6\.2\.[[:digit:]]+$`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"parameter_group_name"},
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_SetEngineVersionOnCreate_MajorUpgrade_6x(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalReplicationGroup elasticache.GlobalReplicationGroup
	var rg elasticache.ReplicationGroup

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	primaryReplicationGroupId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_global_replication_group.test"
	primaryReplicationGroupResourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_engineVersionParam(rName, primaryReplicationGroupId, "5.0.6", "6.2", "default.redis6.x"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, resourceName, &globalReplicationGroup),
					testAccCheckReplicationGroupExists(ctx, primaryReplicationGroupResourceName, &rg),
					resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexp.MustCompile(`^6\.2\.[[:digit:]]+$`)),
					testAccMatchReplicationGroupActualVersion(ctx, &rg, regexp.MustCompile(`^6\.2\.[[:digit:]]+$`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"parameter_group_name"},
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_SetEngineVersionOnCreate_MinorDowngrade(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	primaryReplicationGroupId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccGlobalReplicationGroupConfig_engineVersion(rName, primaryReplicationGroupId, "6.2", "6.0"),
				ExpectError: regexp.MustCompile(`cannot downgrade version when creating, is 6.2.[[:digit:]]+, want 6.0.[[:digit:]]+`),
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_SetParameterGroupOnCreate_NoVersion(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	primaryReplicationGroupId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccGlobalReplicationGroupConfig_param(rName, primaryReplicationGroupId, "6.2", "default.redis6.x"),
				ExpectError: regexp.MustCompile(`cannot change parameter group name without upgrading major engine version`),
				PlanOnly:    true,
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_SetParameterGroupOnCreate_MinorUpgrade(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	primaryReplicationGroupId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccGlobalReplicationGroupConfig_engineVersionParam(rName, primaryReplicationGroupId, "6.0", "6.2", "default.redis6.x"),
				ExpectError: regexp.MustCompile(`cannot change parameter group name on minor engine version upgrade, upgrading from 6\.0\.[[:digit:]]+ to 6\.2\.[[:digit:]]+`),
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_SetEngineVersionOnUpdate_MinorUpgrade(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalReplicationGroup elasticache.GlobalReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	primaryReplicationGroupId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_global_replication_group.test"
	primaryReplicationGroupResourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_engineVersionInherit(rName, primaryReplicationGroupId, "6.0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, resourceName, &globalReplicationGroup),
					resource.TestCheckResourceAttrPair(resourceName, "engine_version_actual", primaryReplicationGroupResourceName, "engine_version_actual"),
					resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexp.MustCompile(`^6\.0\.[[:digit:]]+$`)),
				),
			},
			{
				Config: testAccGlobalReplicationGroupConfig_engineVersion(rName, primaryReplicationGroupId, "6.0", "6.2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, resourceName, &globalReplicationGroup),
					resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexp.MustCompile(`^6\.2\.[[:digit:]]+$`)),
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

	var globalReplicationGroup elasticache.GlobalReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	primaryReplicationGroupId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_global_replication_group.test"
	primaryReplicationGroupResourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_engineVersionInherit(rName, primaryReplicationGroupId, "6.0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, resourceName, &globalReplicationGroup),
					resource.TestCheckResourceAttrPair(resourceName, "engine_version_actual", primaryReplicationGroupResourceName, "engine_version_actual"),
					resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexp.MustCompile(`^6\.0\.[[:digit:]]+$`)),
				),
			},
			{
				Config:   testAccGlobalReplicationGroupConfig_engineVersion(rName, primaryReplicationGroupId, "6.0", "6.x"),
				PlanOnly: true,
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_SetEngineVersionOnUpdate_MinorDowngrade(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalReplicationGroup elasticache.GlobalReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	primaryReplicationGroupId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_global_replication_group.test"
	primaryReplicationGroupResourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_engineVersionInherit(rName, primaryReplicationGroupId, "6.2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, resourceName, &globalReplicationGroup),
					resource.TestCheckResourceAttrPair(resourceName, "engine_version_actual", primaryReplicationGroupResourceName, "engine_version_actual"),
					resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexp.MustCompile(`^6\.2\.[[:digit:]]+$`)),
				),
			},
			{
				Config:      testAccGlobalReplicationGroupConfig_engineVersion(rName, primaryReplicationGroupId, "6.2", "6.0"),
				ExpectError: regexp.MustCompile(`Downgrading ElastiCache Global Replication Group \(.*\) engine version requires replacement`),
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
			// 		resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexp.MustCompile(`^6\.0\.[[:digit:]]+$`)),
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

	var globalReplicationGroup elasticache.GlobalReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	primaryReplicationGroupId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_global_replication_group.test"
	primaryReplicationGroupResourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_engineVersionInherit(rName, primaryReplicationGroupId, "5.0.6"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, resourceName, &globalReplicationGroup),
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
					testAccCheckGlobalReplicationGroupExists(ctx, resourceName, &globalReplicationGroup),
					resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexp.MustCompile(`^6\.2\.[[:digit:]]+$`)),
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

	var globalReplicationGroup elasticache.GlobalReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	primaryReplicationGroupId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_global_replication_group.test"
	primaryReplicationGroupResourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_engineVersionInherit(rName, primaryReplicationGroupId, "5.0.6"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, resourceName, &globalReplicationGroup),
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
					testAccCheckGlobalReplicationGroupExists(ctx, resourceName, &globalReplicationGroup),
					resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexp.MustCompile(`^6\.2\.[[:digit:]]+$`)),
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

	var globalReplicationGroup elasticache.GlobalReplicationGroup

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	primaryReplicationGroupId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_elasticache_global_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_engineVersionInherit(rName, primaryReplicationGroupId, "6.2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, resourceName, &globalReplicationGroup),
					resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexp.MustCompile(`^6\.2\.[[:digit:]]+$`)),
				),
			},
			{
				Config:      testAccGlobalReplicationGroupConfig_param(rName, primaryReplicationGroupId, "6.2", "default.redis6.x"),
				ExpectError: regexp.MustCompile(`cannot change parameter group name without upgrading major engine version`),
				PlanOnly:    true,
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_SetParameterGroupOnUpdate_MinorUpgrade(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalReplicationGroup elasticache.GlobalReplicationGroup

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	primaryReplicationGroupId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_elasticache_global_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_engineVersionInherit(rName, primaryReplicationGroupId, "6.0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, resourceName, &globalReplicationGroup),
					resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexp.MustCompile(`^6\.0\.[[:digit:]]+$`)),
				),
			},
			{
				Config:      testAccGlobalReplicationGroupConfig_engineVersionParam(rName, primaryReplicationGroupId, "6.0", "6.2", "default.redis6.x"),
				ExpectError: regexp.MustCompile(`cannot change parameter group name on minor engine version upgrade, upgrading from 6\.0\.[[:digit:]]+ to 6\.2\.[[:digit:]]+`),
				PlanOnly:    true,
			},
		},
	})
}

func TestAccElastiCacheGlobalReplicationGroup_UpdateParameterGroupName(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var globalReplicationGroup elasticache.GlobalReplicationGroup

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	primaryReplicationGroupId := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	parameterGroupName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_elasticache_global_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheckGlobalReplicationGroup(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, elasticache.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckGlobalReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccGlobalReplicationGroupConfig_engineVersionParam(rName, primaryReplicationGroupId, "5.0.6", "6.2", "default.redis6.x"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckGlobalReplicationGroupExists(ctx, resourceName, &globalReplicationGroup),
					resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexp.MustCompile(`^6\.2\.[[:digit:]]+$`)),
				),
			},
			{
				Config:      testAccGlobalReplicationGroupConfig_engineVersionCustomParam(rName, primaryReplicationGroupId, "5.0.6", "6.2", parameterGroupName, "redis6.x"),
				ExpectError: regexp.MustCompile(`cannot change parameter group name without upgrading major engine version`),
				PlanOnly:    true,
			},
		},
	})
}

func testAccCheckGlobalReplicationGroupExists(ctx context.Context, resourceName string, v *elasticache.GlobalReplicationGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ElastiCache Global Replication Group ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ElastiCacheConn()
		grg, err := tfelasticache.FindGlobalReplicationGroupByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("retrieving ElastiCache Global Replication Group (%s): %w", rs.Primary.ID, err)
		}

		if aws.StringValue(grg.Status) == tfelasticache.GlobalReplicationGroupStatusDeleting || aws.StringValue(grg.Status) == tfelasticache.GlobalReplicationGroupStatusDeleted {
			return fmt.Errorf("ElastiCache Global Replication Group (%s) exists, but is in a non-available state: %s", rs.Primary.ID, aws.StringValue(grg.Status))
		}

		*v = *grg

		return nil
	}
}

func testAccCheckGlobalReplicationGroupDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ElastiCacheConn()

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_elasticache_global_replication_group" {
				continue
			}

			_, err := tfelasticache.FindGlobalReplicationGroupByID(ctx, conn, rs.Primary.ID)
			if tfresource.NotFound(err) {
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
	conn := acctest.Provider.Meta().(*conns.AWSClient).ElastiCacheConn()

	input := &elasticache.DescribeGlobalReplicationGroupsInput{}
	_, err := conn.DescribeGlobalReplicationGroupsWithContext(ctx, input)

	if acctest.PreCheckSkipError(err) ||
		tfawserr.ErrMessageContains(err, elasticache.ErrCodeInvalidParameterValueException, "Access Denied to API Version: APIGlobalDatastore") {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccMatchReplicationGroupActualVersion(ctx context.Context, j *elasticache.ReplicationGroup, r *regexp.Regexp) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ElastiCacheConn()

		cacheCluster := j.NodeGroups[0].NodeGroupMembers[0]
		cluster, err := tfelasticache.FindCacheClusterByID(ctx, conn, aws.StringValue(cacheCluster.CacheClusterId))
		if err != nil {
			return err
		}

		if !r.MatchString(aws.StringValue(cluster.EngineVersion)) {
			return fmt.Errorf("Actual engine version didn't match %q, got %q", r.String(), aws.StringValue(cluster.EngineVersion))
		}
		return nil
	}
}

func testAccGlobalReplicationGroupConfig_basic(rName, primaryReplicationGroupId string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_global_replication_group" "test" {
  global_replication_group_id_suffix = %[1]q
  primary_replication_group_id       = aws_elasticache_replication_group.test.id
}

resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %[2]q
  replication_group_description = "test"

  engine                = "redis"
  engine_version        = "5.0.6"
  node_type             = "cache.m5.large"
  number_cache_clusters = 1
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
  replication_group_id          = %[2]q
  replication_group_description = "test"

  engine                = "redis"
  engine_version        = "5.0.6"
  node_type             = %[3]q
  number_cache_clusters = 1
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
  replication_group_id          = %[2]q
  replication_group_description = "test"

  node_type = "cache.m5.large"

  engine                = "redis"
  engine_version        = "5.0.6"
  number_cache_clusters = 2

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
  replication_group_id          = %[2]q
  replication_group_description = "test"

  engine                = "redis"
  engine_version        = "5.0.6"
  node_type             = "cache.m5.large"
  number_cache_clusters = 1
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
  replication_group_id          = %[2]q
  replication_group_description = "test"

  engine                = "redis"
  engine_version        = "5.0.6"
  node_type             = %[3]q
  number_cache_clusters = 1
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
  replication_group_id          = %[2]q
  replication_group_description = "test"

  engine                = "redis"
  engine_version        = "5.0.6"
  node_type             = %[3]q
  number_cache_clusters = 1

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
  replication_group_id          = %[2]q
  replication_group_description = "test"

  engine                = "redis"
  engine_version        = "5.0.6"
  number_cache_clusters = 1
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
  replication_group_id          = %[2]q
  replication_group_description = "test"

  node_type = "cache.m5.large"

  engine                = "redis"
  engine_version        = "5.0.6"
  number_cache_clusters = 2

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
  replication_group_id          = %[2]q
  replication_group_description = "test"

  node_type = "cache.m5.large"

  engine                = "redis"
  engine_version        = "5.0.6"
  number_cache_clusters = 2

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
  replication_group_id          = %[2]q
  replication_group_description = "test"

  engine                = "redis"
  engine_version        = "5.0.6"
  number_cache_clusters = 2

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

  replication_group_id          = "%[1]s-p"
  replication_group_description = "primary"

  subnet_group_name = aws_elasticache_subnet_group.primary.name

  node_type = "cache.m5.large"

  engine                = "redis"
  engine_version        = "5.0.6"
  number_cache_clusters = 1
}

resource "aws_elasticache_replication_group" "alternate" {
  provider = awsalternate

  replication_group_id          = "%[1]s-a"
  replication_group_description = "alternate"
  global_replication_group_id   = aws_elasticache_global_replication_group.test.global_replication_group_id

  subnet_group_name = aws_elasticache_subnet_group.alternate.name

  number_cache_clusters = 1
}

resource "aws_elasticache_replication_group" "third" {
  provider = awsthird

  replication_group_id          = "%[1]s-t"
  replication_group_description = "third"
  global_replication_group_id   = aws_elasticache_global_replication_group.test.global_replication_group_id

  subnet_group_name = aws_elasticache_subnet_group.third.name

  number_cache_clusters = 1
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

  replication_group_id          = "%[1]s-p"
  replication_group_description = "primary"

  subnet_group_name = aws_elasticache_subnet_group.primary.name

  node_type = "cache.m5.large"

  engine                = "redis"
  engine_version        = "5.0.6"
  number_cache_clusters = 1
}

resource "aws_elasticache_replication_group" "secondary" {
  provider = awsalternate

  replication_group_id          = "%[1]s-a"
  replication_group_description = "alternate"
  global_replication_group_id   = aws_elasticache_global_replication_group.test.global_replication_group_id

  subnet_group_name = aws_elasticache_subnet_group.secondary.name

  number_cache_clusters = 1
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

  replication_group_id          = "%[1]s-p"
  replication_group_description = "primary"

  subnet_group_name = aws_elasticache_subnet_group.primary.name

  node_type = "cache.m5.large"

  engine                = "redis"
  engine_version        = "5.0.6"
  number_cache_clusters = 1
}

resource "aws_elasticache_replication_group" "third" {
  provider = awsthird

  replication_group_id          = "%[1]s-t"
  replication_group_description = "third"
  global_replication_group_id   = aws_elasticache_global_replication_group.test.global_replication_group_id

  subnet_group_name = aws_elasticache_subnet_group.third.name

  number_cache_clusters = 1
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
  replication_group_id          = %[1]q
  replication_group_description = "test"

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
  replication_group_id          = %[1]q
  replication_group_description = "test"

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
  replication_group_id          = %[1]q
  replication_group_description = "test"

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

func testAccGlobalReplicationGroupConfig_engineVersionInherit(rName, primaryReplicationGroupId, repGroupEngineVersion string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_global_replication_group" "test" {
  global_replication_group_id_suffix = %[1]q
  primary_replication_group_id       = aws_elasticache_replication_group.test.id
}

resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %[2]q
  replication_group_description = "test"

  engine                = "redis"
  engine_version        = %[3]q
  node_type             = "cache.m5.large"
  number_cache_clusters = 1
}
`, rName, primaryReplicationGroupId, repGroupEngineVersion)
}

func testAccGlobalReplicationGroupConfig_engineVersion(rName, primaryReplicationGroupId, repGroupEngineVersion, globalEngineVersion string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_global_replication_group" "test" {
  global_replication_group_id_suffix = %[1]q
  primary_replication_group_id       = aws_elasticache_replication_group.test.id

  engine_version = %[4]q
}

resource "aws_elasticache_replication_group" "test" {
  replication_group_id          = %[2]q
  replication_group_description = "test"

  engine                = "redis"
  engine_version        = %[3]q
  node_type             = "cache.m5.large"
  number_cache_clusters = 1

  lifecycle {
    ignore_changes = [engine_version]
  }
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
  replication_group_id          = %[2]q
  replication_group_description = "test"

  engine                = "redis"
  engine_version        = %[3]q
  node_type             = "cache.m5.large"
  number_cache_clusters = 1

  lifecycle {
    ignore_changes = [engine_version]
  }
}
`, rName, primaryReplicationGroupId, repGroupEngineVersion, globalEngineVersion, parameterGroup)
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
  replication_group_id          = %[2]q
  replication_group_description = "test"

  engine                = "redis"
  engine_version        = %[3]q
  node_type             = "cache.m5.large"
  number_cache_clusters = 1

  lifecycle {
    ignore_changes = [engine_version]
  }
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
  replication_group_id          = %[2]q
  replication_group_description = "test"

  engine                = "redis"
  engine_version        = %[3]q
  node_type             = "cache.m5.large"
  number_cache_clusters = 1

  lifecycle {
    ignore_changes = [engine_version]
  }
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
