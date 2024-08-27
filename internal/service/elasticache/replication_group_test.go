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
	"time"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfelasticache "github.com/hashicorp/terraform-provider-aws/internal/service/elasticache"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccElastiCacheReplicationGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg awstypes.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, "redis"),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "elasticache", fmt.Sprintf("replicationgroup:%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "num_cache_clusters", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "multi_az_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "num_node_groups", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "replicas_per_node_group", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", acctest.CtFalse),
					testCheckEngineStuffDefault(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrAutoMinorVersionUpgrade, acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "data_tiering_enabled", acctest.CtFalse),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "auth_token_update_strategy"},
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_basic_v5(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg awstypes.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_v5(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, "redis"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, "5.0.6"),
					resource.TestCheckResourceAttr(resourceName, "engine_version_actual", "5.0.6"),
					// Even though it is ignored, the API returns `true` in this case
					resource.TestCheckResourceAttr(resourceName, names.AttrAutoMinorVersionUpgrade, acctest.CtTrue),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "auth_token_update_strategy"},
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_uppercase(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg awstypes.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_uppercase(strings.ToUpper(rName)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "replication_group_id", rName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "auth_token_update_strategy"},
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_EngineVersion_v7(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg awstypes.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_v7(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, "redis"),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, "7.0"),
					resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexache.MustCompile(`^7\.[[:digit:]]+\.[[:digit:]]+$`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "auth_token_update_strategy"},
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_EngineVersion_update(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v1, v2, v3, v4, v5, v6 awstypes.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_engineVersion(rName, "4.0.10"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, "4.0.10"),
					resource.TestCheckResourceAttr(resourceName, "engine_version_actual", "4.0.10"),
				),
			},
			{
				Config: testAccReplicationGroupConfig_engineVersion(rName, "6.0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &v2),
					testAccCheckReplicationGroupNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, "6.0"),
					resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexache.MustCompile(`^6\.0\.[[:digit:]]+$`)),
				),
			},
			{
				Config: testAccReplicationGroupConfig_engineVersion(rName, "6.2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &v3),
					testAccCheckReplicationGroupNotRecreated(&v2, &v3),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, "6.2"),
					resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexache.MustCompile(`^6\.2\.[[:digit:]]+$`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "auth_token_update_strategy"},
			},
			{
				Config: testAccReplicationGroupConfig_engineVersion(rName, "5.0.6"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &v4),
					testAccCheckReplicationGroupRecreated(&v3, &v4),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, "5.0.6"),
					resource.TestCheckResourceAttr(resourceName, "engine_version_actual", "5.0.6"),
				),
			},
			{
				Config: testAccReplicationGroupConfig_engineVersion(rName, "6.x"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &v5),
					testAccCheckReplicationGroupNotRecreated(&v4, &v5),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, "6.x"),
					resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexache.MustCompile(`^6\.[[:digit:]]+\.[[:digit:]]+$`)),
				),
			},
			{
				Config: testAccReplicationGroupConfig_engineVersion(rName, "6.0"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &v6),
					testAccCheckReplicationGroupRecreated(&v5, &v6),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, "6.0"),
					resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexache.MustCompile(`^6\.0\.[[:digit:]]+$`)),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_EngineVersion_6xToRealVersion(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var v1, v2 awstypes.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_engineVersion(rName, "6.x"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &v1),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, "6.x"),
					resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexache.MustCompile(`^6\.[[:digit:]]+\.[[:digit:]]+$`)),
				),
			},
			{
				// TODO: This will break if there's a Redis 6.x version higher than 6.2.
				// If we create an `aws_elasticache_engine_versions` data source, we can use that to get the expected version
				Config: testAccReplicationGroupConfig_engineVersion(rName, "6.2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &v2),
					testAccCheckReplicationGroupNotRecreated(&v1, &v2),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, "6.2"),
					resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexache.MustCompile(`^6\.2\.[[:digit:]]+$`)),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg awstypes.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfelasticache.ResourceReplicationGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_updateDescription(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg awstypes.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "num_cache_clusters", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test description"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "auth_token_update_strategy"},
			},
			{
				Config: testAccReplicationGroupConfig_updatedDescription(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "num_cache_clusters", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "updated description"),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_updateMaintenanceWindow(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg awstypes.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window", "tue:06:30-tue:07:30"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "auth_token_update_strategy"},
			},
			{
				Config: testAccReplicationGroupConfig_updatedMaintenanceWindow(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "maintenance_window", "wed:03:00-wed:06:00"),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_updateUserGroups(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg awstypes.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	userGroup := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_user(rName, userGroup, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					testAccCheckReplicationGroupUserGroup(ctx, resourceName, fmt.Sprintf("%s-%d", userGroup, 0)),
					resource.TestCheckTypeSetElemAttr(resourceName, "user_group_ids.*", fmt.Sprintf("%s-%d", userGroup, 0)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "auth_token_update_strategy"},
			},
			{
				Config: testAccReplicationGroupConfig_user(rName, userGroup, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					testAccCheckReplicationGroupUserGroup(ctx, resourceName, fmt.Sprintf("%s-%d", userGroup, 1)),
					resource.TestCheckTypeSetElemAttr(resourceName, "user_group_ids.*", fmt.Sprintf("%s-%d", userGroup, 1)),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_updateNodeSize(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg awstypes.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "num_cache_clusters", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "node_type", "cache.t3.small"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "auth_token_update_strategy"},
			},
			{
				Config: testAccReplicationGroupConfig_updatedNodeSize(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "num_cache_clusters", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "node_type", "cache.t3.medium"),
				),
			},
		},
	})
}

// This is a test to prove that we panic we get in https://github.com/hashicorp/terraform/issues/9097
func TestAccElastiCacheReplicationGroup_updateParameterGroup(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg awstypes.ReplicationGroup
	parameterGroupResourceName1 := "aws_elasticache_parameter_group.test.0"
	parameterGroupResourceName2 := "aws_elasticache_parameter_group.test.1"
	resourceName := "aws_elasticache_replication_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_parameterName(rName, 0),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrParameterGroupName, parameterGroupResourceName1, names.AttrName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					"auth_token_update_strategy",
					names.AttrEngineVersion, // because we can't ignore the diff between `6.x` and `6.2`
				},
			},
			{
				Config: testAccReplicationGroupConfig_parameterName(rName, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrParameterGroupName, parameterGroupResourceName2, names.AttrName),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_authToken(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg awstypes.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"
	token1 := sdkacctest.RandString(16)
	token2 := sdkacctest.RandString(16)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_authTokenSetup(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "auth_token_update_strategy"},
			},
			{
				// When adding an auth_token to a previously passwordless replication
				// group, the SET strategy can be used.
				Config: testAccReplicationGroupConfig_authToken(rName, token1, string(awstypes.AuthTokenUpdateStrategyTypeSet)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "transit_encryption_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "auth_token", token1),
					resource.TestCheckResourceAttr(resourceName, "auth_token_update_strategy", string(awstypes.AuthTokenUpdateStrategyTypeSet)),
				),
			},
			{
				Config: testAccReplicationGroupConfig_authToken(rName, token2, string(awstypes.AuthTokenUpdateStrategyTypeRotate)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "transit_encryption_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "auth_token", token2),
					resource.TestCheckResourceAttr(resourceName, "auth_token_update_strategy", string(awstypes.AuthTokenUpdateStrategyTypeRotate)),
				),
			},
			{
				// To explicitly set an auth token and remove the previous one, the modify request
				// should include the auth_token to be kept and the SET auth_token_update_strategy.
				//
				// Ref: https://docs.aws.amazon.com/AmazonElastiCache/latest/red-ug/auth.html#auth-modifyng-token
				Config: testAccReplicationGroupConfig_authToken(rName, token2, string(awstypes.AuthTokenUpdateStrategyTypeSet)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "transit_encryption_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "auth_token", token2),
					resource.TestCheckResourceAttr(resourceName, "auth_token_update_strategy", string(awstypes.AuthTokenUpdateStrategyTypeSet)),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_stateUpgrade5270(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg awstypes.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		CheckDestroy: testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				// At v5.26.0 the resource's schema is v1 and auth_token_update_strategy is not an argument
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "5.26.0",
					},
				},
				Config: testAccReplicationGroupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
				),
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccReplicationGroupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
				),
			},
		},
	})
}

// https://github.com/hashicorp/terraform-provider-aws/issues/38464.
func TestAccElastiCacheReplicationGroup_stateUpgrade5590(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg awstypes.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:   acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		CheckDestroy: testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				ExternalProviders: map[string]resource.ExternalProvider{
					"aws": {
						Source:            "hashicorp/aws",
						VersionConstraint: "4.67.0",
					},
				},
				Config: testAccReplicationGroupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
				),
			},
			{
				ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
				Config:                   testAccReplicationGroupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_vpc(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg awstypes.ReplicationGroup
	resourceName := "aws_elasticache_replication_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_inVPC(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "num_cache_clusters", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "preferred_cache_cluster_azs.#", acctest.Ct1),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "auth_token_update_strategy", "preferred_cache_cluster_azs"},
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_multiAzNotInVPC(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg awstypes.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_multiAZNotInVPCPreferredCacheClusterAZsNotRepeated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "num_cache_clusters", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "multi_az_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "preferred_cache_cluster_azs.#", acctest.Ct2),
					resource.TestCheckResourceAttrPair(resourceName, "preferred_cache_cluster_azs.0", "data.aws_availability_zones.available", "names.0"),
					resource.TestCheckResourceAttrPair(resourceName, "preferred_cache_cluster_azs.1", "data.aws_availability_zones.available", "names.1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "auth_token_update_strategy", "preferred_cache_cluster_azs"},
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_multiAzNotInVPC_repeated(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg awstypes.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_multiAZNotInVPCPreferredCacheClusterAZsRepeated(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "num_cache_clusters", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "multi_az_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "preferred_cache_cluster_azs.#", acctest.Ct4),
					resource.TestCheckResourceAttrPair(resourceName, "preferred_cache_cluster_azs.0", "data.aws_availability_zones.available", "names.0"),
					resource.TestCheckResourceAttrPair(resourceName, "preferred_cache_cluster_azs.1", "data.aws_availability_zones.available", "names.1"),
					resource.TestCheckResourceAttrPair(resourceName, "preferred_cache_cluster_azs.2", "data.aws_availability_zones.available", "names.0"),
					resource.TestCheckResourceAttrPair(resourceName, "preferred_cache_cluster_azs.3", "data.aws_availability_zones.available", "names.1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "auth_token_update_strategy", "preferred_cache_cluster_azs"},
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_multiAzInVPC(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg awstypes.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_multiAZInVPC(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "num_cache_clusters", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "multi_az_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "snapshot_window", "02:00-03:00"),
					resource.TestCheckResourceAttr(resourceName, "snapshot_retention_limit", "7"),
					resource.TestCheckResourceAttrSet(resourceName, "primary_endpoint_address"),
					func(s *terraform.State) error {
						return resource.TestMatchResourceAttr(resourceName, "primary_endpoint_address", regexache.MustCompile(fmt.Sprintf("%s\\..+\\.%s", aws.ToString(rg.ReplicationGroupId), acctest.PartitionDNSSuffix())))(s)
					},
					resource.TestCheckResourceAttrSet(resourceName, "reader_endpoint_address"),
					func(s *terraform.State) error {
						return resource.TestMatchResourceAttr(resourceName, "reader_endpoint_address", regexache.MustCompile(fmt.Sprintf("%s-ro\\..+\\.%s", aws.ToString(rg.ReplicationGroupId), acctest.PartitionDNSSuffix())))(s)
					},
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "auth_token_update_strategy", "preferred_cache_cluster_azs"},
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_deprecatedAvailabilityZones_multiAzInVPC(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg awstypes.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_multiAZInVPC(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "num_cache_clusters", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "multi_az_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "snapshot_window", "02:00-03:00"),
					resource.TestCheckResourceAttr(resourceName, "snapshot_retention_limit", "7"),
					resource.TestCheckResourceAttrSet(resourceName, "primary_endpoint_address"),
					func(s *terraform.State) error {
						return resource.TestMatchResourceAttr(resourceName, "primary_endpoint_address", regexache.MustCompile(fmt.Sprintf("%s\\..+\\.%s", aws.ToString(rg.ReplicationGroupId), acctest.PartitionDNSSuffix())))(s)
					},
					resource.TestCheckResourceAttrSet(resourceName, "reader_endpoint_address"),
					func(s *terraform.State) error {
						return resource.TestMatchResourceAttr(resourceName, "reader_endpoint_address", regexache.MustCompile(fmt.Sprintf("%s-ro\\..+\\.%s", aws.ToString(rg.ReplicationGroupId), acctest.PartitionDNSSuffix())))(s)
					},
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "auth_token_update_strategy", "preferred_cache_cluster_azs"},
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_ValidationMultiAz_noAutomaticFailover(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccReplicationGroupConfig_multiAZNoAutomaticFailover(rName),
				ExpectError: regexache.MustCompile("automatic_failover_enabled must be true if multi_az_enabled is true"),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_ipDiscovery(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg awstypes.ReplicationGroup
	resourceName := "aws_elasticache_replication_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_ipDiscovery(rName, "ipv6"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "6379"),
					resource.TestCheckResourceAttr(resourceName, "node_type", "cache.t3.small"),
					resource.TestCheckResourceAttr(resourceName, "num_node_groups", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "replicas_per_node_group", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "num_cache_clusters", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, names.AttrParameterGroupName, "default.redis7.cluster.on"),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "ip_discovery", "ipv6"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "auth_token_update_strategy", "preferred_cache_cluster_azs"},
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_networkType(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg awstypes.ReplicationGroup
	resourceName := "aws_elasticache_replication_group.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_networkType(rName, "ipv6", "dual_stack"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "6379"),
					resource.TestCheckResourceAttr(resourceName, "node_type", "cache.t3.small"),
					resource.TestCheckResourceAttr(resourceName, "num_node_groups", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "replicas_per_node_group", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "num_cache_clusters", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, names.AttrParameterGroupName, "default.redis7.cluster.on"),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "ip_discovery", "ipv6"),
					resource.TestCheckResourceAttr(resourceName, "network_type", "dual_stack"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "auth_token_update_strategy", "preferred_cache_cluster_azs"},
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_ClusterMode_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg awstypes.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_nativeRedisCluster(rName, 2, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "num_node_groups", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "replicas_per_node_group", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "6379"),
					resource.TestCheckResourceAttrSet(resourceName, "configuration_endpoint_address"),
					resource.TestCheckResourceAttr(resourceName, "multi_az_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "num_cache_clusters", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", acctest.Ct4),
					testCheckEngineStuffClusterEnabledDefault(ctx, resourceName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "auth_token_update_strategy"},
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_ClusterMode_nonClusteredParameterGroup(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg awstypes.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_nativeRedisClusterNonClusteredParameter(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, names.AttrParameterGroupName, "default.redis6.x"),
					resource.TestCheckResourceAttr(resourceName, "num_node_groups", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "replicas_per_node_group", acctest.Ct1),
					resource.TestMatchResourceAttr(resourceName, "primary_endpoint_address", regexache.MustCompile(fmt.Sprintf("%s\\..+\\.%s", rName, acctest.PartitionDNSSuffix()))),
					resource.TestCheckNoResourceAttr(resourceName, "configuration_endpoint_address"),
					resource.TestCheckResourceAttr(resourceName, "num_cache_clusters", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", acctest.Ct2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					"auth_token_update_strategy",
					names.AttrEngineVersion, // because we can't ignore the diff between `6.x` and `6.2`
				},
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_ClusterModeUpdateNumNodeGroups_scaleUp(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg awstypes.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"
	clusterDataSourcePrefix := "data.aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_nativeRedisCluster(rName, 2, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "num_cache_clusters", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "num_node_groups", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "replicas_per_node_group", acctest.Ct1),
					testAccReplicationGroupCheckMemberClusterTags(resourceName, clusterDataSourcePrefix, 4, []kvp{
						{names.AttrKey, names.AttrValue},
					}),
					testCheckEngineStuffClusterEnabledDefault(ctx, resourceName),
				),
			},
			{
				Config: testAccReplicationGroupConfig_nativeRedisCluster(rName, 3, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "num_node_groups", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "replicas_per_node_group", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "num_cache_clusters", "6"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "6"),
					testAccReplicationGroupCheckMemberClusterTags(resourceName, clusterDataSourcePrefix, 6, []kvp{
						{names.AttrKey, names.AttrValue},
					}),
					testCheckEngineStuffClusterEnabledDefault(ctx, resourceName),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_ClusterModeUpdateNumNodeGroups_scaleDown(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg awstypes.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_nativeRedisCluster(rName, 3, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "num_node_groups", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "replicas_per_node_group", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "num_cache_clusters", "6"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "6"),
					testCheckEngineStuffClusterEnabledDefault(ctx, resourceName),
				),
			},
			{
				Config: testAccReplicationGroupConfig_nativeRedisCluster(rName, 2, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "num_node_groups", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "replicas_per_node_group", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "num_cache_clusters", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", acctest.Ct4),
					testCheckEngineStuffClusterEnabledDefault(ctx, resourceName),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_ClusterMode_updateReplicasPerNodeGroup(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg awstypes.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_nativeRedisCluster(rName, 2, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "num_node_groups", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "replicas_per_node_group", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "num_cache_clusters", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", acctest.Ct4),
					testCheckEngineStuffClusterEnabledDefault(ctx, resourceName),
				),
			},
			{
				Config: testAccReplicationGroupConfig_nativeRedisCluster(rName, 2, 3),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "num_node_groups", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "replicas_per_node_group", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "num_cache_clusters", "8"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "8"),
					testCheckEngineStuffClusterEnabledDefault(ctx, resourceName),
				),
			},
			{
				Config: testAccReplicationGroupConfig_nativeRedisCluster(rName, 2, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "num_node_groups", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "replicas_per_node_group", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "num_cache_clusters", "6"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "6"),
					testCheckEngineStuffClusterEnabledDefault(ctx, resourceName),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_ClusterModeUpdateNumNodeGroupsAndReplicasPerNodeGroup_scaleUp(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg awstypes.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_nativeRedisCluster(rName, 2, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "num_node_groups", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "replicas_per_node_group", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "num_cache_clusters", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", acctest.Ct4),
					testCheckEngineStuffClusterEnabledDefault(ctx, resourceName),
				),
			},
			{
				Config: testAccReplicationGroupConfig_nativeRedisCluster(rName, 3, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "num_node_groups", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "replicas_per_node_group", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "num_cache_clusters", "9"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "9"),
					testCheckEngineStuffClusterEnabledDefault(ctx, resourceName),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_ClusterModeUpdateNumNodeGroupsAndReplicasPerNodeGroup_scaleDown(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg awstypes.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_nativeRedisCluster(rName, 3, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "num_node_groups", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "replicas_per_node_group", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "num_cache_clusters", "9"),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", "9"),
					testCheckEngineStuffClusterEnabledDefault(ctx, resourceName),
				),
			},
			{
				Config: testAccReplicationGroupConfig_nativeRedisCluster(rName, 2, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "num_node_groups", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "replicas_per_node_group", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "num_cache_clusters", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", acctest.Ct4),
					testCheckEngineStuffClusterEnabledDefault(ctx, resourceName),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_ClusterMode_singleNode(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg awstypes.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_nativeRedisClusterSingleNode(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrParameterGroupName, "default.redis6.x.cluster.on"),
					resource.TestCheckResourceAttr(resourceName, "num_node_groups", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "replicas_per_node_group", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "multi_az_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "num_cache_clusters", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					"auth_token_update_strategy",
					names.AttrEngineVersion, // because we can't ignore the diff between `6.x` and `6.2`
				},
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_ClusterMode_updateFromDisabled_Compatible_Enabled(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg awstypes.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroup_ClusterMode_updateFromDisabled_Compatible_Enabled(rName, "disabled", false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode", "disabled"),
					resource.TestCheckResourceAttr(resourceName, "num_node_groups", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "replicas_per_node_group", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "num_cache_clusters", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", acctest.Ct2),
				),
			},
			{
				Config: testAccReplicationGroup_ClusterMode_updateFromDisabled_Compatible_Enabled(rName, "compatible", true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode", "compatible"),
					resource.TestCheckResourceAttr(resourceName, "num_node_groups", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "replicas_per_node_group", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "num_cache_clusters", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", acctest.Ct2),
				),
			},
			{
				Config: testAccReplicationGroup_ClusterMode_updateFromDisabled_Compatible_Enabled(rName, names.AttrEnabled, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "cluster_mode", names.AttrEnabled),
					resource.TestCheckResourceAttr(resourceName, "num_node_groups", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "replicas_per_node_group", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "num_cache_clusters", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_cacheClustersConflictsWithReplicasPerNodeGroup(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccReplicationGroupConfig_cacheClustersConflictsWithReplicasPerNodeGroup(rName),
				ExpectError: regexache.MustCompile(`"replicas_per_node_group": conflicts with num_cache_clusters`),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_clusteringAndCacheNodesCausesError(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccReplicationGroupConfig_nativeRedisClusterError(rName),
				ExpectError: regexache.MustCompile(`"num_node_groups": conflicts with num_cache_clusters`),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_enableSnapshotting(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg awstypes.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "snapshot_retention_limit", acctest.Ct0),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "auth_token_update_strategy"},
			},
			{
				Config: testAccReplicationGroupConfig_enableSnapshotting(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "snapshot_retention_limit", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_transitEncryptionWithAuthToken(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg awstypes.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"
	authToken := sdkacctest.RandString(16)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_transitEncryptionWithAuthToken(rName, authToken),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "auth_token", authToken),
					resource.TestCheckResourceAttr(resourceName, "transit_encryption_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, "transit_encryption_mode"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "auth_token", "auth_token_update_strategy", "preferred_cache_cluster_azs"},
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_transitEncryption5x(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg1, rg2 awstypes.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_transitEncryptionEnabled5x(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg1),
					resource.TestCheckResourceAttr(resourceName, "transit_encryption_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttrSet(resourceName, "transit_encryption_mode"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "auth_token", "auth_token_update_strategy", "preferred_cache_cluster_azs"},
			},
			{
				// With Redis engine versions < 7.0.5, transit_encryption_enabled can only be set
				// during cluster creation. Modifying the argument should force a replacement.
				Config: testAccReplicationGroupConfig_transitEncryptionDisabled5x(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg2),
					testAccCheckReplicationGroupRecreated(&rg1, &rg2),
					resource.TestCheckResourceAttr(resourceName, "transit_encryption_enabled", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_transitEncryption7x(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg1, rg2, rg3, rg4 awstypes.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_transitEncryptionEnabled7x(rName, string(awstypes.TransitEncryptionModePreferred)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg1),
					resource.TestCheckResourceAttr(resourceName, "transit_encryption_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "transit_encryption_mode", string(awstypes.TransitEncryptionModePreferred)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "auth_token", "auth_token_update_strategy", "preferred_cache_cluster_azs"},
			},
			{
				// With Redis engine versions >= 7.0.5, transit_encryption_mode can be modified in-place.
				Config: testAccReplicationGroupConfig_transitEncryptionEnabled7x(rName, string(awstypes.TransitEncryptionModeRequired)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg2),
					testAccCheckReplicationGroupNotRecreated(&rg1, &rg2),
					resource.TestCheckResourceAttr(resourceName, "transit_encryption_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "transit_encryption_mode", string(awstypes.TransitEncryptionModeRequired)),
				),
			},
			{
				// Before disabling transit encryption, mode must be transitioned back to "preferred" first.
				Config: testAccReplicationGroupConfig_transitEncryptionEnabled7x(rName, string(awstypes.TransitEncryptionModePreferred)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg3),
					testAccCheckReplicationGroupNotRecreated(&rg2, &rg3),
					resource.TestCheckResourceAttr(resourceName, "transit_encryption_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "transit_encryption_mode", string(awstypes.TransitEncryptionModePreferred)),
				),
			},
			{
				// With Redis engine versions >= 7.0.5, transit_encryption_enabled can be modified in-place.
				Config: testAccReplicationGroupConfig_transitEncryptionDisabled7x(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg4),
					testAccCheckReplicationGroupNotRecreated(&rg3, &rg4),
					resource.TestCheckResourceAttr(resourceName, "transit_encryption_enabled", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_enableAtRestEncryption(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg awstypes.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_enableAtRestEncryption(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, "at_rest_encryption_enabled", acctest.CtTrue),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "auth_token_update_strategy"},
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_useCMKKMSKeyID(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg awstypes.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_useCMKKMSKeyID(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrKMSKeyID),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_NumberCacheClusters_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var replicationGroup awstypes.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"
	clusterDataSourcePrefix := "data.aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_numberCacheClusters(rName, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "multi_az_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "num_cache_clusters", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", acctest.Ct2),
					testAccReplicationGroupCheckMemberClusterTags(resourceName, clusterDataSourcePrefix, 2, []kvp{
						{names.AttrKey, names.AttrValue},
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "auth_token_update_strategy"},
			},
			{
				Config: testAccReplicationGroupConfig_numberCacheClusters(rName, 4),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "multi_az_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "num_cache_clusters", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", acctest.Ct4),
					testAccReplicationGroupCheckMemberClusterTags(resourceName, clusterDataSourcePrefix, 4, []kvp{
						{names.AttrKey, names.AttrValue},
					}),
				),
			},
			{
				Config: testAccReplicationGroupConfig_numberCacheClusters(rName, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "multi_az_enabled", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "num_cache_clusters", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", acctest.Ct2),
					testAccReplicationGroupCheckMemberClusterTags(resourceName, clusterDataSourcePrefix, 2, []kvp{
						{names.AttrKey, names.AttrValue},
					}),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_NumberCacheClusters_autoFailoverDisabled(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var replicationGroup awstypes.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	autoFailoverEnabled := false
	multiAZEnabled := false

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_failoverMultiAZ(rName, 3, autoFailoverEnabled, multiAZEnabled),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", strconv.FormatBool(autoFailoverEnabled)),
					resource.TestCheckResourceAttr(resourceName, "multi_az_enabled", strconv.FormatBool(multiAZEnabled)),
					resource.TestCheckResourceAttr(resourceName, "num_cache_clusters", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", acctest.Ct3),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "auth_token_update_strategy"},
			},
			{
				PreConfig: func() {
					// Ensure that primary is on the node we are trying to delete
					conn := acctest.Provider.Meta().(*conns.AWSClient).ElastiCacheClient(ctx)
					timeout := 40 * time.Minute

					if err := resourceReplicationGroupSetPrimaryClusterID(ctx, conn, rName, formatReplicationGroupClusterID(rName, 3), timeout); err != nil {
						t.Fatalf("error changing primary cache cluster: %s", err)
					}
				},
				Config: testAccReplicationGroupConfig_failoverMultiAZ(rName, 2, autoFailoverEnabled, multiAZEnabled),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", strconv.FormatBool(autoFailoverEnabled)),
					resource.TestCheckResourceAttr(resourceName, "multi_az_enabled", strconv.FormatBool(multiAZEnabled)),
					resource.TestCheckResourceAttr(resourceName, "num_cache_clusters", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_NumberCacheClusters_autoFailoverEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var replicationGroup awstypes.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	autoFailoverEnabled := true
	multiAZEnabled := false

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_failoverMultiAZ(rName, 3, autoFailoverEnabled, multiAZEnabled),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", strconv.FormatBool(autoFailoverEnabled)),
					resource.TestCheckResourceAttr(resourceName, "multi_az_enabled", strconv.FormatBool(multiAZEnabled)),
					resource.TestCheckResourceAttr(resourceName, "num_cache_clusters", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", acctest.Ct3),
				),
			},
			{
				PreConfig: func() {
					// Ensure that primary is on the node we are trying to delete
					conn := acctest.Provider.Meta().(*conns.AWSClient).ElastiCacheClient(ctx)
					timeout := 40 * time.Minute

					// Must disable automatic failover first
					if err := resourceReplicationGroupDisableAutomaticFailover(ctx, conn, rName, timeout); err != nil {
						t.Fatalf("error disabling automatic failover: %s", err)
					}

					// Set primary
					if err := resourceReplicationGroupSetPrimaryClusterID(ctx, conn, rName, formatReplicationGroupClusterID(rName, 3), timeout); err != nil {
						t.Fatalf("error changing primary cache cluster: %s", err)
					}

					// Re-enable automatic failover like nothing ever happened
					if err := resourceReplicationGroupEnableAutomaticFailover(ctx, conn, rName, multiAZEnabled, timeout); err != nil {
						t.Fatalf("error re-enabling automatic failover: %s", err)
					}
				},
				Config: testAccReplicationGroupConfig_failoverMultiAZ(rName, 2, autoFailoverEnabled, multiAZEnabled),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", strconv.FormatBool(autoFailoverEnabled)),
					resource.TestCheckResourceAttr(resourceName, "multi_az_enabled", strconv.FormatBool(multiAZEnabled)),
					resource.TestCheckResourceAttr(resourceName, "num_cache_clusters", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_autoFailoverEnabled_validateNumberCacheClusters(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	const (
		autoFailoverEnabled = true
		multiAZDisabled     = false
	)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccReplicationGroupConfig_failoverMultiAZ(rName, 1, autoFailoverEnabled, multiAZDisabled),
				ExpectError: regexache.MustCompile(`"num_cache_clusters": must be at least 2 if automatic_failover_enabled is true`),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_NumberCacheClusters_multiAZEnabled(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var replicationGroup awstypes.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	autoFailoverEnabled := true
	multiAZEnabled := true

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_failoverMultiAZ(rName, 3, autoFailoverEnabled, multiAZEnabled),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", strconv.FormatBool(autoFailoverEnabled)),
					resource.TestCheckResourceAttr(resourceName, "multi_az_enabled", strconv.FormatBool(multiAZEnabled)),
					resource.TestCheckResourceAttr(resourceName, "num_cache_clusters", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", acctest.Ct3),
				),
			},
			{
				PreConfig: func() {
					// Ensure that primary is on the node we are trying to delete
					conn := acctest.Provider.Meta().(*conns.AWSClient).ElastiCacheClient(ctx)
					timeout := 40 * time.Minute

					// Must disable automatic failover first
					if err := resourceReplicationGroupDisableAutomaticFailover(ctx, conn, rName, timeout); err != nil {
						t.Fatalf("error disabling automatic failover: %s", err)
					}

					// Set primary
					if err := resourceReplicationGroupSetPrimaryClusterID(ctx, conn, rName, formatReplicationGroupClusterID(rName, 3), timeout); err != nil {
						t.Fatalf("error changing primary cache cluster: %s", err)
					}

					// Re-enable automatic failover like nothing ever happened
					if err := resourceReplicationGroupEnableAutomaticFailover(ctx, conn, rName, multiAZEnabled, timeout); err != nil {
						t.Fatalf("error re-enabling automatic failover: %s", err)
					}
				},
				Config: testAccReplicationGroupConfig_failoverMultiAZ(rName, 2, autoFailoverEnabled, multiAZEnabled),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", strconv.FormatBool(autoFailoverEnabled)),
					resource.TestCheckResourceAttr(resourceName, "multi_az_enabled", strconv.FormatBool(multiAZEnabled)),
					resource.TestCheckResourceAttr(resourceName, "num_cache_clusters", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_NumberCacheClustersMemberClusterDisappears_noChange(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var replicationGroup awstypes.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_numberCacheClusters(rName, 3),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "num_cache_clusters", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", acctest.Ct3),
				),
			},
			{
				PreConfig: func() {
					// Remove one of the Cache Clusters
					conn := acctest.Provider.Meta().(*conns.AWSClient).ElastiCacheClient(ctx)
					timeout := 40 * time.Minute

					cacheClusterID := formatReplicationGroupClusterID(rName, 2)

					if err := tfelasticache.DeleteCacheCluster(ctx, conn, cacheClusterID, ""); err != nil {
						t.Fatalf("error deleting Cache Cluster (%s): %s", cacheClusterID, err)
					}

					if _, err := tfelasticache.WaitCacheClusterDeleted(ctx, conn, cacheClusterID, timeout); err != nil {
						t.Fatalf("error deleting Cache Cluster (%s): %s", cacheClusterID, err)
					}
				},
				Config: testAccReplicationGroupConfig_numberCacheClusters(rName, 3),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "num_cache_clusters", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", acctest.Ct3),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_NumberCacheClustersMemberClusterDisappears_addMemberCluster(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var replicationGroup awstypes.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_numberCacheClusters(rName, 3),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "num_cache_clusters", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", acctest.Ct3),
				),
			},
			{
				PreConfig: func() {
					// Remove one of the Cache Clusters
					conn := acctest.Provider.Meta().(*conns.AWSClient).ElastiCacheClient(ctx)
					timeout := 40 * time.Minute

					cacheClusterID := formatReplicationGroupClusterID(rName, 2)

					if err := tfelasticache.DeleteCacheCluster(ctx, conn, cacheClusterID, ""); err != nil {
						t.Fatalf("error deleting Cache Cluster (%s): %s", cacheClusterID, err)
					}

					if _, err := tfelasticache.WaitCacheClusterDeleted(ctx, conn, cacheClusterID, timeout); err != nil {
						t.Fatalf("error deleting Cache Cluster (%s): %s", cacheClusterID, err)
					}
				},
				Config: testAccReplicationGroupConfig_numberCacheClusters(rName, 4),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "num_cache_clusters", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", acctest.Ct4),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_NumberCacheClustersMemberClusterDisappearsRemoveMemberCluster_atTargetSize(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var replicationGroup awstypes.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_numberCacheClusters(rName, 3),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "num_cache_clusters", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", acctest.Ct3),
				),
			},
			{
				PreConfig: func() {
					// Remove one of the Cache Clusters
					conn := acctest.Provider.Meta().(*conns.AWSClient).ElastiCacheClient(ctx)
					timeout := 40 * time.Minute

					cacheClusterID := formatReplicationGroupClusterID(rName, 2)

					if err := tfelasticache.DeleteCacheCluster(ctx, conn, cacheClusterID, ""); err != nil {
						t.Fatalf("error deleting Cache Cluster (%s): %s", cacheClusterID, err)
					}

					if _, err := tfelasticache.WaitCacheClusterDeleted(ctx, conn, cacheClusterID, timeout); err != nil {
						t.Fatalf("error deleting Cache Cluster (%s): %s", cacheClusterID, err)
					}
				},
				Config: testAccReplicationGroupConfig_numberCacheClusters(rName, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "num_cache_clusters", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_NumberCacheClustersMemberClusterDisappearsRemoveMemberCluster_scaleDown(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var replicationGroup awstypes.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_numberCacheClusters(rName, 4),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "num_cache_clusters", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", acctest.Ct4),
				),
			},
			{
				PreConfig: func() {
					// Remove one of the Cache Clusters
					conn := acctest.Provider.Meta().(*conns.AWSClient).ElastiCacheClient(ctx)
					timeout := 40 * time.Minute

					cacheClusterID := formatReplicationGroupClusterID(rName, 2)

					if err := tfelasticache.DeleteCacheCluster(ctx, conn, cacheClusterID, ""); err != nil {
						t.Fatalf("error deleting Cache Cluster (%s): %s", cacheClusterID, err)
					}

					if _, err := tfelasticache.WaitCacheClusterDeleted(ctx, conn, cacheClusterID, timeout); err != nil {
						t.Fatalf("error deleting Cache Cluster (%s): %s", cacheClusterID, err)
					}
				},
				Config: testAccReplicationGroupConfig_numberCacheClusters(rName, 2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &replicationGroup),
					resource.TestCheckResourceAttr(resourceName, "num_cache_clusters", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "member_clusters.#", acctest.Ct2),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg awstypes.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"
	clusterDataSourcePrefix := "data.aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					testAccReplicationGroupCheckMemberClusterTags(resourceName, clusterDataSourcePrefix, 2, []kvp{
						{acctest.CtKey1, acctest.CtValue1},
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "auth_token_update_strategy"},
			},
			{
				Config: testAccReplicationGroupConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					testAccReplicationGroupCheckMemberClusterTags(resourceName, clusterDataSourcePrefix, 2, []kvp{
						{acctest.CtKey1, acctest.CtValue1Updated},
						{acctest.CtKey2, acctest.CtValue2},
					}),
				),
			},
			{
				Config: testAccReplicationGroupConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					testAccReplicationGroupCheckMemberClusterTags(resourceName, clusterDataSourcePrefix, 2, []kvp{
						{acctest.CtKey2, acctest.CtValue2},
					}),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_tagWithOtherModification(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg awstypes.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"
	clusterDataSourcePrefix := "data.aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_versionAndTag(rName, "5.0.4", acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, "5.0.4"),
					testAccReplicationGroupCheckMemberClusterTags(resourceName, clusterDataSourcePrefix, 2, []kvp{
						{acctest.CtKey1, acctest.CtValue1},
					}),
				),
			},
			{
				Config: testAccReplicationGroupConfig_versionAndTag(rName, "5.0.6", acctest.CtKey1, acctest.CtValue1Updated),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, "5.0.6"),
					testAccReplicationGroupCheckMemberClusterTags(resourceName, clusterDataSourcePrefix, 2, []kvp{
						{acctest.CtKey1, acctest.CtValue1Updated},
					}),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_finalSnapshot(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg awstypes.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_finalSnapshot(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, names.AttrFinalSnapshotIdentifier, rName),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_autoMinorVersionUpgrade(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg awstypes.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckClusterDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_autoMinorVersionUpgrade(rName, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, names.AttrAutoMinorVersionUpgrade, acctest.CtFalse),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "auth_token_update_strategy"},
			},
			{
				Config: testAccReplicationGroupConfig_autoMinorVersionUpgrade(rName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, names.AttrAutoMinorVersionUpgrade, acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_Validation_noNodeType(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 2),
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccReplicationGroupConfig_validationNoNodeType(rName),
				ExpectError: regexache.MustCompile(`"node_type" is required unless "global_replication_group_id" is set.`),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_Validation_globalReplicationGroupIdAndNodeType(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 2),
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccReplicationGroupConfig_validationGlobalIdAndNodeType(rName),
				ExpectError: regexache.MustCompile(`"global_replication_group_id": conflicts with node_type`),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_GlobalReplicationGroupID_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg awstypes.ReplicationGroup
	var pg awstypes.CacheParameterGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"
	primaryGroupResourceName := "aws_elasticache_replication_group.primary"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 2),
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckReplicationGroupDestroy(ctx),
			testAccCheckGlobalReplicationGroupMemberParameterGroupDestroy(ctx, &pg),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_globalIDBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					testAccCheckReplicationGroupParameterGroupExists(ctx, &rg, &pg),
					resource.TestCheckResourceAttrPair(resourceName, "global_replication_group_id", "aws_elasticache_global_replication_group.test", "global_replication_group_id"),
					resource.TestCheckResourceAttrPair(resourceName, "node_type", primaryGroupResourceName, "node_type"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrEngine, primaryGroupResourceName, names.AttrEngine),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrEngineVersion, primaryGroupResourceName, names.AttrEngineVersion),
					resource.TestMatchResourceAttr(resourceName, names.AttrParameterGroupName, regexache.MustCompile(fmt.Sprintf("^global-datastore-%s-", rName))),
					resource.TestCheckResourceAttr(resourceName, "num_cache_clusters", acctest.Ct1),
					resource.TestCheckResourceAttr(primaryGroupResourceName, "num_cache_clusters", acctest.Ct2),
				),
			},
			{
				Config:                  testAccReplicationGroupConfig_globalIDBasic(rName),
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "auth_token_update_strategy"},
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_GlobalReplicationGroupID_full(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg1, rg2 awstypes.ReplicationGroup
	var pg1, pg2 awstypes.CacheParameterGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"
	primaryGroupResourceName := "aws_elasticache_replication_group.primary"

	initialNumCacheClusters := 2
	updatedNumCacheClusters := 3

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 2),
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckReplicationGroupDestroy(ctx),
			testAccCheckGlobalReplicationGroupMemberParameterGroupDestroy(ctx, &pg1),
			testAccCheckGlobalReplicationGroupMemberParameterGroupDestroy(ctx, &pg2),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_globalIDFull(rName, initialNumCacheClusters),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg1),
					testAccCheckReplicationGroupParameterGroupExists(ctx, &rg1, &pg1),
					resource.TestCheckResourceAttrPair(resourceName, "global_replication_group_id", "aws_elasticache_global_replication_group.test", "global_replication_group_id"),
					resource.TestCheckResourceAttrPair(resourceName, "node_type", primaryGroupResourceName, "node_type"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrEngine, primaryGroupResourceName, names.AttrEngine),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrEngineVersion, primaryGroupResourceName, names.AttrEngineVersion),
					resource.TestMatchResourceAttr(resourceName, names.AttrParameterGroupName, regexache.MustCompile(fmt.Sprintf("^global-datastore-%s-", rName))),

					resource.TestCheckResourceAttr(resourceName, "num_cache_clusters", strconv.Itoa(initialNumCacheClusters)),
					resource.TestCheckResourceAttrPair(resourceName, "multi_az_enabled", primaryGroupResourceName, "multi_az_enabled"),
					resource.TestCheckResourceAttrPair(resourceName, "automatic_failover_enabled", primaryGroupResourceName, "automatic_failover_enabled"),

					resource.TestCheckResourceAttr(resourceName, names.AttrPort, "16379"),

					resource.TestCheckResourceAttrPair(resourceName, "at_rest_encryption_enabled", primaryGroupResourceName, "at_rest_encryption_enabled"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_encryption_enabled", primaryGroupResourceName, "transit_encryption_enabled"),
				),
			},
			{
				Config:                  testAccReplicationGroupConfig_globalIDBasic(rName),
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "auth_token_update_strategy"},
			},
			{
				Config: testAccReplicationGroupConfig_globalIDFull(rName, updatedNumCacheClusters),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg2),
					testAccCheckReplicationGroupParameterGroupExists(ctx, &rg2, &pg2),
					resource.TestCheckResourceAttr(resourceName, "num_cache_clusters", strconv.Itoa(updatedNumCacheClusters)),
				),
			},
		},
	})
}

// Test for out-of-band deletion
// Naming to allow grouping all TestAccAWSElastiCacheReplicationGroup_GlobalReplicationGroupID_* tests
func TestAccElastiCacheReplicationGroup_GlobalReplicationGroupID_disappears(t *testing.T) { // nosemgrep:ci.acceptance-test-naming-parent-disappears
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg awstypes.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 2),
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_globalIDBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfelasticache.ResourceReplicationGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_GlobalReplicationGroupIDClusterMode_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg1, rg2 awstypes.ReplicationGroup
	var pg1, pg2 awstypes.CacheParameterGroup
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"
	primaryGroupResourceName := "aws_elasticache_replication_group.primary"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 2),
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckReplicationGroupDestroy(ctx),
			testAccCheckGlobalReplicationGroupMemberParameterGroupDestroy(ctx, &pg1),
			testAccCheckGlobalReplicationGroupMemberParameterGroupDestroy(ctx, &pg2),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_globalIDClusterMode(rName, 2, 1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg1),
					testAccCheckReplicationGroupParameterGroupExists(ctx, &rg1, &pg1),
					resource.TestCheckResourceAttr(resourceName, "num_node_groups", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "replicas_per_node_group", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "automatic_failover_enabled", acctest.CtTrue),
					resource.TestMatchResourceAttr(resourceName, names.AttrParameterGroupName, regexache.MustCompile(fmt.Sprintf("^global-datastore-%s-", rName))),

					resource.TestCheckResourceAttr(primaryGroupResourceName, "num_node_groups", acctest.Ct2),
					resource.TestCheckResourceAttr(primaryGroupResourceName, "replicas_per_node_group", acctest.Ct2),
				),
			},
			{
				Config:                  testAccReplicationGroupConfig_globalIDBasic(rName),
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "auth_token_update_strategy"},
			},
			{
				Config: testAccReplicationGroupConfig_globalIDClusterMode(rName, 1, 3),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg2),
					testAccCheckReplicationGroupParameterGroupExists(ctx, &rg2, &pg2),
					resource.TestCheckResourceAttr(resourceName, "num_node_groups", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "replicas_per_node_group", acctest.Ct3),

					resource.TestCheckResourceAttr(primaryGroupResourceName, "num_node_groups", acctest.Ct2),
					resource.TestCheckResourceAttr(primaryGroupResourceName, "replicas_per_node_group", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_GlobalReplicationGroupIDClusterModeValidation_numNodeGroupsOnSecondary(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckMultipleRegion(t, 2)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5FactoriesMultipleRegions(ctx, t, 2),
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config:      testAccReplicationGroupConfig_globalIDClusterModeNumNodeOnSecondary(rName),
				ExpectError: regexache.MustCompile(`"global_replication_group_id": conflicts with num_node_groups`),
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_dataTiering(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var (
		rg      awstypes.ReplicationGroup
		version awstypes.CacheEngineVersion
	)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_dataTiering(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					testCheckRedisEngineVersionLatest(ctx, &version),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, "redis"),
					func(s *terraform.State) error {
						return resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, *version.EngineVersion)(s)
					},
					resource.TestCheckResourceAttr(resourceName, "data_tiering_enabled", acctest.CtTrue),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{names.AttrApplyImmediately, "auth_token_update_strategy"},
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_Engine_Redis_LogDeliveryConfigurations_ClusterMode_Disabled(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg awstypes.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_dataSourceEngineRedisLogDeliveryConfigurations(rName, false, true, string(awstypes.DestinationTypeCloudWatchLogs), string(awstypes.LogFormatText), true, string(awstypes.DestinationTypeCloudWatchLogs), string(awstypes.LogFormatText)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, "redis"),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", acctest.CtFalse),
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
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					"auth_token_update_strategy",
					names.AttrEngineVersion, // because we can't ignore the diff between `6.x` and `6.2`
				},
			},
			{
				Config: testAccReplicationGroupConfig_dataSourceEngineRedisLogDeliveryConfigurations(rName, false, true, string(awstypes.DestinationTypeCloudWatchLogs), string(awstypes.LogFormatText), true, string(awstypes.DestinationTypeKinesisFirehose), string(awstypes.LogFormatJson)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, "redis"),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", acctest.CtFalse),
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
				Config: testAccReplicationGroupConfig_dataSourceEngineRedisLogDeliveryConfigurations(rName, false, true, string(awstypes.DestinationTypeKinesisFirehose), string(awstypes.LogFormatJson), false, "", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
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
				Config: testAccReplicationGroupConfig_dataSourceEngineRedisLogDeliveryConfigurations(rName, false, false, "", "", false, "", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, "redis"),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", acctest.CtFalse),
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
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					"auth_token_update_strategy",
					names.AttrEngineVersion, // because we can't ignore the diff between `6.x` and `6.2`
				},
			},
		},
	})
}

func TestAccElastiCacheReplicationGroup_Engine_Redis_LogDeliveryConfigurations_ClusterMode_Enabled(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var rg awstypes.ReplicationGroup
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_elasticache_replication_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckReplicationGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccReplicationGroupConfig_dataSourceEngineRedisLogDeliveryConfigurations(rName, true, true, string(awstypes.DestinationTypeCloudWatchLogs), string(awstypes.LogFormatText), true, string(awstypes.DestinationTypeCloudWatchLogs), string(awstypes.LogFormatText)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, "redis"),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrParameterGroupName, "default.redis6.x.cluster.on"),
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
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					"auth_token_update_strategy",
					names.AttrEngineVersion, // because we can't ignore the diff between `6.x` and `6.2`
				},
			},
			{
				Config: testAccReplicationGroupConfig_dataSourceEngineRedisLogDeliveryConfigurations(rName, true, true, string(awstypes.DestinationTypeCloudWatchLogs), string(awstypes.LogFormatText), true, string(awstypes.DestinationTypeKinesisFirehose), string(awstypes.LogFormatJson)),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, "redis"),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrParameterGroupName, "default.redis6.x.cluster.on"),
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
				Config: testAccReplicationGroupConfig_dataSourceEngineRedisLogDeliveryConfigurations(rName, true, true, string(awstypes.DestinationTypeKinesisFirehose), string(awstypes.LogFormatJson), false, "", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, "redis"),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, names.AttrParameterGroupName, "default.redis6.x.cluster.on"),
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
				Config: testAccReplicationGroupConfig_dataSourceEngineRedisLogDeliveryConfigurations(rName, true, false, "", "", false, "", ""),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckReplicationGroupExists(ctx, resourceName, &rg),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, "redis"),
					resource.TestCheckResourceAttr(resourceName, "cluster_enabled", acctest.CtTrue),
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
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					names.AttrApplyImmediately,
					"auth_token_update_strategy",
					names.AttrEngineVersion, // because we can't ignore the diff between `6.x` and `6.2`
				},
			},
		},
	})
}

func testAccCheckReplicationGroupExists(ctx context.Context, n string, v *awstypes.ReplicationGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ElastiCacheClient(ctx)

		output, err := tfelasticache.FindReplicationGroupByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckReplicationGroupDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ElastiCacheClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_elasticache_replication_group" {
				continue
			}

			_, err := tfelasticache.FindReplicationGroupByID(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("ElastiCache Replication Group (%s) still exists", rs.Primary.ID)
		}
		return nil
	}
}

func testAccCheckReplicationGroupParameterGroupExists(ctx context.Context, rg *awstypes.ReplicationGroup, v *awstypes.CacheParameterGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ElastiCacheClient(ctx)

		cacheClusterID := aws.ToString(rg.NodeGroups[0].NodeGroupMembers[0].CacheClusterId)
		cluster, err := tfelasticache.FindCacheClusterByID(ctx, conn, cacheClusterID)

		if err != nil {
			return fmt.Errorf("reading ElastiCache Cluster (%s): %w", cacheClusterID, err)
		}

		name := aws.ToString(cluster.CacheParameterGroup.CacheParameterGroupName)
		output, err := tfelasticache.FindCacheParameterGroupByName(ctx, conn, name)

		if err != nil {
			return fmt.Errorf("reading ElastiCache Parameter Group (%s): %w", name, err)
		}

		*v = *output

		return nil
	}
}

func testAccCheckGlobalReplicationGroupMemberParameterGroupDestroy(ctx context.Context, v *awstypes.CacheParameterGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ElastiCacheClient(ctx)

		name := aws.ToString(v.CacheParameterGroupName)
		_, err := tfelasticache.FindCacheParameterGroupByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("ElastiCache Parameter Group (%s) still exists", name)
	}
}

func testAccCheckReplicationGroupUserGroup(ctx context.Context, n, userGroupID string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ElastiCacheClient(ctx)

		id := rs.Primary.ID
		output, err := tfelasticache.FindReplicationGroupByID(ctx, conn, id)

		if err != nil {
			return err
		}

		if len(output.UserGroupIds) < 1 {
			return fmt.Errorf("ElastiCache Replication Group (%s) was not assigned any User Groups", id)
		}

		if v := output.UserGroupIds[0]; v != userGroupID {
			return fmt.Errorf("ElastiCache Replication Group (%s) was not assigned User Group (%s), User Group was (%s) instead", n, userGroupID, v)
		}

		return nil
	}
}

func testAccCheckReplicationGroupRecreated(i, j *awstypes.ReplicationGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.ToTime(i.ReplicationGroupCreateTime).Equal(aws.ToTime(j.ReplicationGroupCreateTime)) {
			return errors.New("ElastiCache Replication Group not recreated")
		}

		return nil
	}
}

func testAccCheckReplicationGroupNotRecreated(i, j *awstypes.ReplicationGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !aws.ToTime(i.ReplicationGroupCreateTime).Equal(aws.ToTime(j.ReplicationGroupCreateTime)) {
			return errors.New("ElastiCache Replication Group recreated")
		}

		return nil
	}
}

func testCheckEngineStuffDefault(ctx context.Context, resourceName string) resource.TestCheckFunc {
	var (
		version        awstypes.CacheEngineVersion
		parameterGroup awstypes.CacheParameterGroup
	)

	checks := []resource.TestCheckFunc{
		testCheckRedisEngineVersionLatest(ctx, &version),
		testCheckRedisParameterGroupDefault(ctx, &version, &parameterGroup),
		func(s *terraform.State) error {
			return resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, *version.EngineVersion)(s)
		},
		func(s *terraform.State) error {
			return resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexache.MustCompile(fmt.Sprintf(`^%s\.[[:digit:]]+$`, *version.EngineVersion)))(s)
		},
		func(s *terraform.State) error {
			return resource.TestCheckResourceAttr(resourceName, names.AttrParameterGroupName, *parameterGroup.CacheParameterGroupName)(s)
		},
	}

	return resource.ComposeAggregateTestCheckFunc(checks...)
}

func testCheckRedisEngineVersionLatest(ctx context.Context, v *awstypes.CacheEngineVersion) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ElastiCacheClient(ctx)

		versions, err := conn.DescribeCacheEngineVersions(ctx, &elasticache.DescribeCacheEngineVersionsInput{
			Engine:      aws.String("redis"),
			DefaultOnly: aws.Bool(true),
		})
		if err != nil {
			return err
		}
		if versions == nil || len(versions.CacheEngineVersions) == 0 {
			return errors.New("empty result")
		}
		if l := len(versions.CacheEngineVersions); l > 1 {
			return fmt.Errorf("too many results: %d", l)
		}

		*v = versions.CacheEngineVersions[0]

		return nil
	}
}

func testCheckRedisParameterGroupDefault(ctx context.Context, version *awstypes.CacheEngineVersion, v *awstypes.CacheParameterGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ElastiCacheClient(ctx)

		output, err := tfelasticache.FindCacheParameterGroup(ctx, conn, &elasticache.DescribeCacheParameterGroupsInput{}, tfslices.PredicateAnd(
			func(v *awstypes.CacheParameterGroup) bool {
				return aws.ToString(v.CacheParameterGroupFamily) == aws.ToString(version.CacheParameterGroupFamily)
			},
			func(v *awstypes.CacheParameterGroup) bool {
				name := aws.ToString(v.CacheParameterGroupName)
				return strings.HasPrefix(name, "default.") && !strings.HasSuffix(name, ".cluster.on")
			},
		))

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testCheckEngineStuffClusterEnabledDefault(ctx context.Context, resourceName string) resource.TestCheckFunc {
	var (
		version        awstypes.CacheEngineVersion
		parameterGroup awstypes.CacheParameterGroup
	)

	checks := []resource.TestCheckFunc{
		testCheckRedisEngineVersionLatest(ctx, &version),
		testCheckRedisParameterGroupClusterEnabledDefault(ctx, &version, &parameterGroup),
		func(s *terraform.State) error {
			return resource.TestCheckResourceAttr(resourceName, names.AttrEngineVersion, *version.EngineVersion)(s)
		},
		func(s *terraform.State) error {
			return resource.TestMatchResourceAttr(resourceName, "engine_version_actual", regexache.MustCompile(fmt.Sprintf(`^%s\.[[:digit:]]+$`, *version.EngineVersion)))(s)
		},
		func(s *terraform.State) error {
			return resource.TestCheckResourceAttr(resourceName, names.AttrParameterGroupName, *parameterGroup.CacheParameterGroupName)(s)
		},
	}

	return resource.ComposeAggregateTestCheckFunc(checks...)
}

func testCheckRedisParameterGroupClusterEnabledDefault(ctx context.Context, version *awstypes.CacheEngineVersion, v *awstypes.CacheParameterGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).ElastiCacheClient(ctx)

		output, err := tfelasticache.FindCacheParameterGroup(ctx, conn, &elasticache.DescribeCacheParameterGroupsInput{}, tfslices.PredicateAnd(
			func(v *awstypes.CacheParameterGroup) bool {
				return aws.ToString(v.CacheParameterGroupFamily) == aws.ToString(version.CacheParameterGroupFamily)
			},
			func(v *awstypes.CacheParameterGroup) bool {
				name := aws.ToString(v.CacheParameterGroupName)
				return strings.HasPrefix(name, "default.") && strings.HasSuffix(name, ".cluster.on")
			},
		))

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

type kvp struct {
	key   string
	value string
}

func testAccReplicationGroupCheckMemberClusterTags(resourceName, dataSourceNamePrefix string, memberCount int, kvs []kvp) resource.TestCheckFunc {
	checks := testAccCheckResourceTags(resourceName, kvs)
	checks = append(checks, resource.TestCheckResourceAttr(resourceName, "member_clusters.#", strconv.Itoa(memberCount))) // sanity check

	for i := 0; i < memberCount; i++ {
		dataSourceName := fmt.Sprintf("%s.%d", dataSourceNamePrefix, i)
		checks = append(checks, testAccCheckResourceTags(dataSourceName, kvs)...)
	}
	return resource.ComposeAggregateTestCheckFunc(checks...)
}

func testAccCheckResourceTags(resourceName string, kvs []kvp) []resource.TestCheckFunc {
	checks := make([]resource.TestCheckFunc, 1, 1+len(kvs))
	checks[0] = resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, strconv.Itoa(len(kvs)))
	for _, kv := range kvs {
		checks = append(checks, resource.TestCheckResourceAttr(resourceName, fmt.Sprintf("tags.%s", kv.key), kv.value))
	}
	return checks
}

func testAccReplicationGroupConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id = %[1]q
  description          = "test description"
  node_type            = "cache.t3.small"
  port                 = 6379
  apply_immediately    = true
  maintenance_window   = "tue:06:30-tue:07:30"
  snapshot_window      = "01:00-02:00"
}
`, rName)
}

func testAccReplicationGroupConfig_cacheClustersConflictsWithReplicasPerNodeGroup(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id = %[1]q
  description          = "test description"
  node_type            = "cache.t3.small"

  automatic_failover_enabled = true
  num_cache_clusters         = 2
  replicas_per_node_group    = 0
}
`, rName)
}

func testAccReplicationGroupConfig_v5(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id = %[1]q
  description          = "test description"
  node_type            = "cache.t3.small"
  engine_version       = "5.0.6"
}
`, rName)
}

func testAccReplicationGroupConfig_v7(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id = %[1]q
  description          = "test description"
  node_type            = "cache.t3.small"
  engine_version       = "7.0"
  parameter_group_name = aws_elasticache_parameter_group.test.name
  engine               = "redis"
}

resource "aws_elasticache_parameter_group" "test" {
  name   = %[1]q
  family = "redis7"
}
`, rName)
}

func testAccReplicationGroupConfig_uppercase(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  node_type            = "cache.t2.micro"
  num_cache_clusters   = 1
  port                 = 6379
  description          = "test description"
  replication_group_id = %[1]q
  subnet_group_name    = aws_elasticache_subnet_group.test.name
}

resource "aws_elasticache_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}
`, rName),
	)
}

func testAccReplicationGroupConfig_engineVersion(rName, engineVersion string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id = %[1]q
  description          = "test description"
  node_type            = "cache.t3.small"
  num_cache_clusters   = 2
  engine_version       = %[2]q
  apply_immediately    = true
  maintenance_window   = "tue:06:30-tue:07:30"
  snapshot_window      = "01:00-02:00"
}
`, rName, engineVersion)
}

func testAccReplicationGroupConfig_enableSnapshotting(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id     = %[1]q
  description              = "test description"
  node_type                = "cache.t3.small"
  port                     = 6379
  apply_immediately        = true
  maintenance_window       = "tue:06:30-tue:07:30"
  snapshot_window          = "01:00-02:00"
  snapshot_retention_limit = 2
}
`, rName)
}

func testAccReplicationGroupConfig_parameterName(rName string, parameterGroupNameIndex int) string {
	return fmt.Sprintf(`
resource "aws_elasticache_parameter_group" "test" {
  count = 2

  # We do not have a data source for "latest" ElastiCache family
  # so unfortunately we must hardcode this for now
  family = "redis6.x"

  name = "%[1]s-${count.index}"

  parameter {
    name  = "maxmemory-policy"
    value = "allkeys-lru"
  }
}

resource "aws_elasticache_replication_group" "test" {
  apply_immediately    = true
  node_type            = "cache.t3.small"
  num_cache_clusters   = 2
  engine_version       = "6.x"
  parameter_group_name = aws_elasticache_parameter_group.test[%[2]d].name
  description          = "test description"
  replication_group_id = %[1]q
}
`, rName, parameterGroupNameIndex)
}

func testAccReplicationGroupConfig_updatedDescription(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id = %[1]q
  description          = "updated description"
  node_type            = "cache.t3.small"
  port                 = 6379
  apply_immediately    = true
}
`, rName)
}

func testAccReplicationGroupConfig_updatedMaintenanceWindow(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id = %[1]q
  description          = "updated description"
  node_type            = "cache.t3.small"
  port                 = 6379
  apply_immediately    = true
  maintenance_window   = "wed:03:00-wed:06:00"
  snapshot_window      = "01:00-02:00"
}
`, rName)
}

func testAccReplicationGroupConfig_updatedNodeSize(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id = %[1]q
  description          = "updated description"
  node_type            = "cache.t3.medium"
  port                 = 6379
  apply_immediately    = true
}
`, rName)
}

func testAccReplicationGroupConfig_user(rName, userGroup string, flag int) string {
	return fmt.Sprintf(`
resource "aws_elasticache_user" "test" {
  count = 2

  user_id       = "%[2]s-${count.index}"
  user_name     = "default"
  access_string = "on ~app::* -@all +@read +@hash +@bitmap +@geo -setbit -bitfield -hset -hsetnx -hmset -hincrby -hincrbyfloat -hdel -bitop -geoadd -georadius -georadiusbymember"
  engine        = "REDIS"
  passwords     = ["password123456789"]
}

resource "aws_elasticache_user_group" "test" {
  count = 2

  user_group_id = "%[2]s-${count.index}"
  engine        = "REDIS"
  user_ids      = [aws_elasticache_user.test[count.index].user_id]
}

resource "aws_elasticache_replication_group" "test" {
  replication_group_id       = %[1]q
  description                = "test description"
  node_type                  = "cache.t3.small"
  port                       = 6379
  apply_immediately          = true
  maintenance_window         = "tue:06:30-tue:07:30"
  snapshot_window            = "01:00-02:00"
  transit_encryption_enabled = true
  user_group_ids             = [aws_elasticache_user_group.test[%[3]d].id]
}
`, rName, userGroup, flag)
}

func testAccReplicationGroupConfig_inVPC(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 1),
		fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id        = %[1]q
  description                 = "test description"
  node_type                   = "cache.t3.small"
  num_cache_clusters          = 1
  port                        = 6379
  subnet_group_name           = aws_elasticache_subnet_group.test.name
  security_group_ids          = [aws_security_group.test.id]
  preferred_cache_cluster_azs = [data.aws_availability_zones.available.names[0]]
}

resource "aws_elasticache_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_security_group" "test" {
  name        = %[1]q
  description = "tf-test-security-group-descr"
  vpc_id      = aws_vpc.test.id

  ingress {
    from_port   = -1
    to_port     = -1
    protocol    = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
`, rName),
	)
}

func testAccReplicationGroupConfig_multiAZNotInVPCPreferredCacheClusterAZsNotRepeated(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id        = %[1]q
  description                 = "test description"
  num_cache_clusters          = 2
  node_type                   = "cache.t3.small"
  automatic_failover_enabled  = true
  multi_az_enabled            = true
  preferred_cache_cluster_azs = slice(data.aws_availability_zones.available.names, 0, 2)
}
`, rName),
	)
}

func testAccReplicationGroupConfig_multiAZNotInVPCPreferredCacheClusterAZsRepeated(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id        = %[1]q
  description                 = "test description"
  num_cache_clusters          = 4
  node_type                   = "cache.t3.small"
  automatic_failover_enabled  = true
  multi_az_enabled            = true
  preferred_cache_cluster_azs = concat(slice(data.aws_availability_zones.available.names, 0, 2), slice(data.aws_availability_zones.available.names, 0, 2))
}
`, rName),
	)
}

func testAccReplicationGroupConfig_multiAZInVPC(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id        = %[1]q
  description                 = "test description"
  node_type                   = "cache.t3.small"
  num_cache_clusters          = 2
  port                        = 6379
  subnet_group_name           = aws_elasticache_subnet_group.test.name
  security_group_ids          = [aws_security_group.test.id]
  preferred_cache_cluster_azs = [data.aws_availability_zones.available.names[0], data.aws_availability_zones.available.names[1]]
  automatic_failover_enabled  = true
  multi_az_enabled            = true
  snapshot_window             = "02:00-03:00"
  snapshot_retention_limit    = 7
}

resource "aws_elasticache_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_security_group" "test" {
  name        = %[1]q
  description = "tf-test-security-group-descr"
  vpc_id      = aws_vpc.test.id

  ingress {
    from_port   = -1
    to_port     = -1
    protocol    = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
`, rName),
	)
}

func testAccReplicationGroupConfig_ipDiscovery(rName, ipDiscovery string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnetsIPv6(rName, 2),
		fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id       = %[1]q
  description                = "test description"
  node_type                  = "cache.t3.small"
  num_node_groups            = 2
  replicas_per_node_group    = 1
  port                       = 6379
  parameter_group_name       = "default.redis7.cluster.on"
  automatic_failover_enabled = true
  subnet_group_name          = aws_elasticache_subnet_group.test.name
  ip_discovery               = %[2]q
  network_type               = "dual_stack"
  security_group_ids         = [aws_security_group.test.id]
}

resource "aws_elasticache_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_security_group" "test" {
  name        = %[1]q
  description = "tf-test-security-group-descr"
  vpc_id      = aws_vpc.test.id

  ingress {
    from_port   = -1
    to_port     = -1
    protocol    = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
`, rName, ipDiscovery),
	)
}

func testAccReplicationGroupConfig_networkType(rName, ipDiscovery, networkType string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnetsIPv6(rName, 2),
		fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id       = %[1]q
  description                = "test description"
  node_type                  = "cache.t3.small"
  num_node_groups            = 2
  replicas_per_node_group    = 1
  port                       = 6379
  parameter_group_name       = "default.redis7.cluster.on"
  automatic_failover_enabled = true
  subnet_group_name          = aws_elasticache_subnet_group.test.name
  ip_discovery               = %[2]q
  network_type               = %[3]q
  security_group_ids         = [aws_security_group.test.id]
}

resource "aws_elasticache_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_security_group" "test" {
  name        = %[1]q
  description = "tf-test-security-group-descr"
  vpc_id      = aws_vpc.test.id

  ingress {
    from_port   = -1
    to_port     = -1
    protocol    = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
`, rName, ipDiscovery, networkType),
	)
}

func testAccReplicationGroupConfig_multiAZNoAutomaticFailover(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id       = %[1]q
  description                = "test description"
  num_cache_clusters         = 1
  node_type                  = "cache.t3.small"
  automatic_failover_enabled = false
  multi_az_enabled           = true
}
`, rName)
}

func testAccReplicationGroupConfig_nativeRedisClusterError(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id       = %[1]q
  description                = "test description"
  node_type                  = "cache.t2.micro"
  port                       = 6379
  subnet_group_name          = aws_elasticache_subnet_group.test.name
  security_group_ids         = [aws_security_group.test.id]
  automatic_failover_enabled = true
  replicas_per_node_group    = 1
  num_node_groups            = 2
  num_cache_clusters         = 3
}

resource "aws_elasticache_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_security_group" "test" {
  name        = %[1]q
  description = "tf-test-security-group-descr"
  vpc_id      = aws_vpc.test.id

  ingress {
    from_port   = -1
    to_port     = -1
    protocol    = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
`, rName),
	)
}

func testAccReplicationGroupConfig_nativeRedisCluster(rName string, numNodeGroups, replicasPerNodeGroup int) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		testAccReplicationGroupClusterData(numNodeGroups*(1+replicasPerNodeGroup)),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "192.168.0.0/16"

  tags = {
    Name = "terraform-testacc-elasticache-replication-group-native-redis-cluster"
  }
}

resource "aws_subnet" "test" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "192.168.0.0/20"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-elasticache-replication-group-native-redis-cluster-test"
  }
}

resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "192.168.16.0/20"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = "tf-acc-elasticache-replication-group-native-redis-cluster-test"
  }
}

resource "aws_elasticache_subnet_group" "test" {
  name        = %[1]q
  description = "tf-test-cache-subnet-group-descr"

  subnet_ids = [
    aws_subnet.test.id,
    aws_subnet.test.id,
  ]
}

resource "aws_security_group" "test" {
  name        = %[1]q
  description = "tf-test-security-group-descr"
  vpc_id      = aws_vpc.test.id

  ingress {
    from_port   = -1
    to_port     = -1
    protocol    = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_elasticache_replication_group" "test" {
  replication_group_id       = %[1]q
  description                = "test description"
  node_type                  = "cache.t2.micro"
  port                       = 6379
  subnet_group_name          = aws_elasticache_subnet_group.test.name
  security_group_ids         = [aws_security_group.test.id]
  automatic_failover_enabled = true
  num_node_groups            = %[2]d
  replicas_per_node_group    = %[3]d

  tags = {
    key = "value"
  }
}
`, rName, numNodeGroups, replicasPerNodeGroup),
	)
}

func testAccReplicationGroupConfig_nativeRedisClusterNonClusteredParameter(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id       = %[1]q
  description                = "test description"
  node_type                  = "cache.t2.medium"
  automatic_failover_enabled = false
  engine_version             = "6.x"
  parameter_group_name       = "default.redis6.x"
  num_node_groups            = 1
  replicas_per_node_group    = 1
}
`, rName),
	)
}

func testAccReplicationGroupConfig_nativeRedisClusterSingleNode(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id       = %[1]q
  description                = "test description"
  node_type                  = "cache.t2.medium"
  automatic_failover_enabled = true
  engine_version             = "6.x"
  parameter_group_name       = "default.redis6.x.cluster.on"
  num_node_groups            = 1
  replicas_per_node_group    = 0
}
`, rName),
	)
}

func testAccReplicationGroup_ClusterMode_updateFromDisabled_Compatible_Enabled(rName, clusterMode string, enableClusterMode bool) string {
	return acctest.ConfigCompose(
		acctest.ConfigAvailableAZsNoOptIn(),
		fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id       = %[1]q
  description                = "test description"
  node_type                  = "cache.t2.medium"
  apply_immediately          = true
  automatic_failover_enabled = true
  cluster_mode               = %[2]q
  engine_version             = "7.1"
  parameter_group_name       = tobool("%[3]t") ? "default.redis7.cluster.on" : "default.redis7"
  num_node_groups            = 1
  replicas_per_node_group    = 1
  timeouts {
    create = "60m"
    update = "60m"
  }
}
`, rName, clusterMode, enableClusterMode),
	)
}

func testAccReplicationGroupConfig_useCMKKMSKeyID(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 1),
		fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id       = %[1]q
  description                = "test description"
  node_type                  = "cache.t2.micro"
  num_cache_clusters         = "1"
  port                       = 6379
  subnet_group_name          = aws_elasticache_subnet_group.test.name
  security_group_ids         = [aws_security_group.test.id]
  at_rest_encryption_enabled = true
  kms_key_id                 = aws_kms_key.test.arn
}

resource "aws_elasticache_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_security_group" "test" {
  name        = %[1]q
  description = "tf-test-security-group-descr"
  vpc_id      = aws_vpc.test.id

  ingress {
    from_port   = -1
    to_port     = -1
    protocol    = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_kms_key" "test" {
  description = "tf-test-cmk-kms-key-id"
}
`, rName),
	)
}

func testAccReplicationGroupConfig_enableAtRestEncryption(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 1),
		fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id       = %[1]q
  description                = "test description"
  node_type                  = "cache.t2.micro"
  num_cache_clusters         = "1"
  port                       = 6379
  subnet_group_name          = aws_elasticache_subnet_group.test.name
  security_group_ids         = [aws_security_group.test.id]
  at_rest_encryption_enabled = true
}

resource "aws_elasticache_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_security_group" "test" {
  name        = %[1]q
  description = "tf-test-security-group-descr"
  vpc_id      = aws_vpc.test.id

  ingress {
    from_port   = -1
    to_port     = -1
    protocol    = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
`, rName),
	)
}

// Dependencies shared across all tests exercising the transit_encryption_enabled argument
func testAccReplicationGroupConfig_transitEncryptionBase(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 1),
		fmt.Sprintf(`
resource "aws_elasticache_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_security_group" "test" {
  name        = %[1]q
  description = "tf-test-security-group-descr"
  vpc_id      = aws_vpc.test.id

  ingress {
    from_port   = -1
    to_port     = -1
    protocol    = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
`, rName),
	)
}

func testAccReplicationGroupConfig_transitEncryptionWithAuthToken(rName, authToken string) string {
	return acctest.ConfigCompose(
		testAccReplicationGroupConfig_transitEncryptionBase(rName),
		fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id       = %[1]q
  description                = "test description"
  node_type                  = "cache.t2.micro"
  num_cache_clusters         = "1"
  port                       = 6379
  subnet_group_name          = aws_elasticache_subnet_group.test.name
  security_group_ids         = [aws_security_group.test.id]
  parameter_group_name       = "default.redis5.0"
  engine_version             = "5.0.6"
  auth_token                 = %[2]q
  transit_encryption_enabled = true
  apply_immediately          = true
}
`, rName, authToken),
	)
}

func testAccReplicationGroupConfig_transitEncryptionEnabled5x(rName string) string {
	return acctest.ConfigCompose(
		testAccReplicationGroupConfig_transitEncryptionBase(rName),
		fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id       = %[1]q
  description                = "test description"
  node_type                  = "cache.t2.micro"
  num_cache_clusters         = "1"
  port                       = 6379
  subnet_group_name          = aws_elasticache_subnet_group.test.name
  security_group_ids         = [aws_security_group.test.id]
  parameter_group_name       = "default.redis5.0"
  engine_version             = "5.0.6"
  transit_encryption_enabled = true
  apply_immediately          = true
}
`, rName),
	)
}

func testAccReplicationGroupConfig_transitEncryptionDisabled5x(rName string) string {
	return acctest.ConfigCompose(
		testAccReplicationGroupConfig_transitEncryptionBase(rName),
		fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id       = %[1]q
  description                = "test description"
  node_type                  = "cache.t2.micro"
  num_cache_clusters         = "1"
  port                       = 6379
  subnet_group_name          = aws_elasticache_subnet_group.test.name
  security_group_ids         = [aws_security_group.test.id]
  parameter_group_name       = "default.redis5.0"
  engine_version             = "5.0.6"
  transit_encryption_enabled = false
  apply_immediately          = true
}
`, rName),
	)
}

func testAccReplicationGroupConfig_transitEncryptionEnabled7x(rName, transitEncryptionMode string) string {
	return acctest.ConfigCompose(
		testAccReplicationGroupConfig_transitEncryptionBase(rName),
		fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id       = %[1]q
  description                = "test description"
  node_type                  = "cache.t2.micro"
  num_cache_clusters         = "1"
  port                       = 6379
  subnet_group_name          = aws_elasticache_subnet_group.test.name
  security_group_ids         = [aws_security_group.test.id]
  parameter_group_name       = "default.redis7"
  engine_version             = "7.0"
  transit_encryption_enabled = true
  transit_encryption_mode    = %[2]q
  apply_immediately          = true
}
`, rName, transitEncryptionMode),
	)
}

func testAccReplicationGroupConfig_transitEncryptionDisabled7x(rName string) string {
	return acctest.ConfigCompose(
		testAccReplicationGroupConfig_transitEncryptionBase(rName),
		fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id       = %[1]q
  description                = "test description"
  node_type                  = "cache.t2.micro"
  num_cache_clusters         = "1"
  port                       = 6379
  subnet_group_name          = aws_elasticache_subnet_group.test.name
  security_group_ids         = [aws_security_group.test.id]
  parameter_group_name       = "default.redis7"
  engine_version             = "7.0"
  transit_encryption_enabled = false
  apply_immediately          = true
}
`, rName),
	)
}

// Identical to the _authToken configutaion, but with no authorization yet
// configured. This will execercise the case when authorization is added
// to a replication group which previously had none.
func testAccReplicationGroupConfig_authTokenSetup(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 1),
		fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id = %[1]q
  description          = "test description"
  node_type            = "cache.t2.micro"
  num_cache_clusters   = "1"
  port                 = 6379
  subnet_group_name    = aws_elasticache_subnet_group.test.name
  security_group_ids   = [aws_security_group.test.id]
  parameter_group_name = "default.redis5.0"
  engine_version       = "5.0.6"
}

resource "aws_elasticache_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_security_group" "test" {
  name        = %[1]q
  description = "tf-test-security-group-descr"
  vpc_id      = aws_vpc.test.id

  ingress {
    from_port   = -1
    to_port     = -1
    protocol    = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
`, rName))
}

func testAccReplicationGroupConfig_authToken(rName string, authToken string, updateStrategy string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 1),
		fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id       = %[1]q
  description                = "test description"
  node_type                  = "cache.t2.micro"
  num_cache_clusters         = "1"
  port                       = 6379
  subnet_group_name          = aws_elasticache_subnet_group.test.name
  security_group_ids         = [aws_security_group.test.id]
  parameter_group_name       = "default.redis5.0"
  engine_version             = "5.0.6"
  transit_encryption_enabled = true
  auth_token                 = %[2]q
  auth_token_update_strategy = %[3]q
}

resource "aws_elasticache_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_security_group" "test" {
  name        = %[1]q
  description = "tf-test-security-group-descr"
  vpc_id      = aws_vpc.test.id

  ingress {
    from_port   = -1
    to_port     = -1
    protocol    = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
`, rName, authToken, updateStrategy))
}

func testAccReplicationGroupConfig_numberCacheClusters(rName string, numberCacheClusters int) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
		testAccReplicationGroupClusterData(numberCacheClusters),
		fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  node_type            = "cache.t2.micro"
  num_cache_clusters   = %[2]d
  replication_group_id = %[1]q
  description          = "test description"
  subnet_group_name    = aws_elasticache_subnet_group.test.name

  tags = {
    key = "value"
  }
}

resource "aws_elasticache_subnet_group" "test" {
  name       = "%[1]s"
  subnet_ids = aws_subnet.test[*].id
}
`, rName, numberCacheClusters),
	)
}

func testAccReplicationGroupConfig_failoverMultiAZ(rName string, numberCacheClusters int, autoFailover, multiAZ bool) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 2),
		fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  # InvalidParameterCombination: Automatic failover is not supported for T1 and T2 cache node types.
  automatic_failover_enabled = %[3]t
  multi_az_enabled           = %[4]t
  node_type                  = "cache.t3.medium"
  num_cache_clusters         = %[2]d
  replication_group_id       = %[1]q
  description                = "test description"
  subnet_group_name          = aws_elasticache_subnet_group.test.name
}

resource "aws_elasticache_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}
`, rName, numberCacheClusters, autoFailover, multiAZ),
	)
}

func testAccReplicationGroupConfig_tags1(rName, tagKey1, tagValue1 string) string {
	const clusterCount = 2
	return acctest.ConfigCompose(
		testAccReplicationGroupClusterData(clusterCount),
		fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id = %[1]q
  description          = "test description"
  node_type            = "cache.t3.small"
  num_cache_clusters   = %[2]d
  port                 = 6379
  apply_immediately    = true
  maintenance_window   = "tue:06:30-tue:07:30"
  snapshot_window      = "01:00-02:00"

  tags = {
    %[3]q = %[4]q
  }
}
`, rName, clusterCount, tagKey1, tagValue1),
	)
}

func testAccReplicationGroupConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	const clusterCount = 2
	return acctest.ConfigCompose(
		testAccReplicationGroupClusterData(clusterCount),
		fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id = %[1]q
  description          = "test description"
  node_type            = "cache.t3.small"
  num_cache_clusters   = %[2]d
  port                 = 6379
  apply_immediately    = true
  maintenance_window   = "tue:06:30-tue:07:30"
  snapshot_window      = "01:00-02:00"

  tags = {
    %[3]q = %[4]q
    %[5]q = %[6]q
  }
}
`, rName, clusterCount, tagKey1, tagValue1, tagKey2, tagValue2),
	)
}

func testAccReplicationGroupConfig_versionAndTag(rName, version, tagKey1, tagValue1 string) string {
	const clusterCount = 2
	return acctest.ConfigCompose(
		testAccReplicationGroupClusterData(clusterCount),
		fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id = %[1]q
  description          = "test description"
  node_type            = "cache.t3.small"
  num_cache_clusters   = %[2]d
  apply_immediately    = true
  engine_version       = %[3]q

  tags = {
    %[4]q = %[5]q
  }
}
`, rName, clusterCount, version, tagKey1, tagValue1),
	)
}

func testAccReplicationGroupClusterData(count int) string {
	return fmt.Sprintf(`
data "aws_elasticache_cluster" "test" {
  count = %[1]d

  cluster_id = tolist(aws_elasticache_replication_group.test.member_clusters)[count.index]
}
`, count)
}

func testAccReplicationGroupConfig_finalSnapshot(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id = %[1]q
  description          = "test description"
  node_type            = "cache.t3.small"
  num_cache_clusters   = 1

  final_snapshot_identifier = %[1]q
}
`, rName)
}

func testAccReplicationGroupConfig_autoMinorVersionUpgrade(rName string, enable bool) string {
	return fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id       = %[1]q
  description                = "test description"
  node_type                  = "cache.t3.small"
  auto_minor_version_upgrade = %[2]t
}
`, rName, enable)
}

func testAccReplicationGroupConfig_validationNoNodeType(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id = %[1]q
  description          = "test description"
  num_cache_clusters   = 1
}
`, rName)
}

func testAccReplicationGroupConfig_validationGlobalIdAndNodeType(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(2),
		testAccVPCBaseWithProvider(rName, "test", acctest.ProviderName, 1),
		testAccVPCBaseWithProvider(rName, "primary", acctest.ProviderNameAlternate, 1),
		fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  provider = aws

  replication_group_id        = "%[1]s-s"
  description                 = "secondary"
  global_replication_group_id = aws_elasticache_global_replication_group.test.global_replication_group_id
  subnet_group_name           = aws_elasticache_subnet_group.test.name
  node_type                   = "cache.m5.large"
  num_cache_clusters          = 1
}

resource "aws_elasticache_global_replication_group" "test" {
  provider = awsalternate

  global_replication_group_id_suffix = %[1]q
  primary_replication_group_id       = aws_elasticache_replication_group.primary.id
}

resource "aws_elasticache_replication_group" "primary" {
  provider = awsalternate

  replication_group_id = "%[1]s-p"
  description          = "primary"
  subnet_group_name    = aws_elasticache_subnet_group.primary.name
  node_type            = "cache.m5.large"
  engine               = "redis"
  engine_version       = "5.0.6"
  num_cache_clusters   = 1
}
`, rName),
	)
}

func testAccReplicationGroupConfig_globalIDBasic(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(2),
		testAccVPCBaseWithProvider(rName, "test", acctest.ProviderName, 1),
		testAccVPCBaseWithProvider(rName, "primary", acctest.ProviderNameAlternate, 1),
		fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id        = "%[1]s-s"
  description                 = "secondary"
  global_replication_group_id = aws_elasticache_global_replication_group.test.global_replication_group_id
  subnet_group_name           = aws_elasticache_subnet_group.test.name
}

resource "aws_elasticache_global_replication_group" "test" {
  provider = awsalternate

  global_replication_group_id_suffix = %[1]q
  primary_replication_group_id       = aws_elasticache_replication_group.primary.id
}

resource "aws_elasticache_replication_group" "primary" {
  provider = awsalternate

  replication_group_id = "%[1]s-p"
  description          = "primary"
  subnet_group_name    = aws_elasticache_subnet_group.primary.name
  node_type            = "cache.m5.large"
  engine               = "redis"
  engine_version       = "5.0.6"
  num_cache_clusters   = 2
}
`, rName),
	)
}

func testAccReplicationGroupConfig_globalIDFull(rName string, numCacheClusters int) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(2),
		testAccVPCBaseWithProvider(rName, "test", acctest.ProviderName, 2),
		testAccVPCBaseWithProvider(rName, "primary", acctest.ProviderNameAlternate, 2),
		fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id        = "%[1]s-s"
  description                 = "secondary"
  global_replication_group_id = aws_elasticache_global_replication_group.test.global_replication_group_id
  subnet_group_name           = aws_elasticache_subnet_group.test.name
  num_cache_clusters          = %[2]d
  automatic_failover_enabled  = true
  multi_az_enabled            = true
  port                        = 16379
}

resource "aws_elasticache_global_replication_group" "test" {
  provider = awsalternate

  global_replication_group_id_suffix = %[1]q
  primary_replication_group_id       = aws_elasticache_replication_group.primary.id
}

resource "aws_elasticache_replication_group" "primary" {
  provider = awsalternate

  replication_group_id       = "%[1]s-p"
  description                = "primary"
  subnet_group_name          = aws_elasticache_subnet_group.primary.name
  node_type                  = "cache.m5.large"
  engine                     = "redis"
  engine_version             = "5.0.6"
  num_cache_clusters         = 2
  automatic_failover_enabled = true
  multi_az_enabled           = true

  port = 6379

  at_rest_encryption_enabled = true
  transit_encryption_enabled = true
}
`, rName, numCacheClusters),
	)
}

func testAccReplicationGroupConfig_globalIDClusterMode(rName string, primaryReplicaCount, secondaryReplicaCount int) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(2),
		testAccVPCBaseWithProvider(rName, "test", acctest.ProviderName, 2),
		testAccVPCBaseWithProvider(rName, "primary", acctest.ProviderNameAlternate, 2),
		fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id        = "%[1]s-s"
  description                 = "secondary"
  global_replication_group_id = aws_elasticache_global_replication_group.test.global_replication_group_id
  subnet_group_name           = aws_elasticache_subnet_group.test.name
  automatic_failover_enabled  = true
  replicas_per_node_group     = %[3]d
}

resource "aws_elasticache_global_replication_group" "test" {
  provider = awsalternate

  global_replication_group_id_suffix = %[1]q
  primary_replication_group_id       = aws_elasticache_replication_group.primary.id
}

resource "aws_elasticache_replication_group" "primary" {
  provider = awsalternate

  replication_group_id       = "%[1]s-p"
  description                = "primary"
  subnet_group_name          = aws_elasticache_subnet_group.primary.name
  engine                     = "redis"
  engine_version             = "6.2"
  node_type                  = "cache.m5.large"
  parameter_group_name       = "default.redis6.x.cluster.on"
  automatic_failover_enabled = true
  num_node_groups            = 2
  replicas_per_node_group    = %[2]d
}
`, rName, primaryReplicaCount, secondaryReplicaCount),
	)
}

func testAccReplicationGroupConfig_globalIDClusterModeNumNodeOnSecondary(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigMultipleRegionProvider(2),
		testAccVPCBaseWithProvider(rName, "test", acctest.ProviderName, 2),
		testAccVPCBaseWithProvider(rName, "primary", acctest.ProviderNameAlternate, 2),
		fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id        = "%[1]s-s"
  description                 = "secondary"
  global_replication_group_id = aws_elasticache_global_replication_group.test.global_replication_group_id
  subnet_group_name           = aws_elasticache_subnet_group.test.name
  automatic_failover_enabled  = true
  num_node_groups             = 2
  replicas_per_node_group     = 1
}

resource "aws_elasticache_global_replication_group" "test" {
  provider = awsalternate

  global_replication_group_id_suffix = %[1]q
  primary_replication_group_id       = aws_elasticache_replication_group.primary.id
}

resource "aws_elasticache_replication_group" "primary" {
  provider = awsalternate

  replication_group_id       = "%[1]s-p"
  description                = "primary"
  subnet_group_name          = aws_elasticache_subnet_group.primary.name
  engine                     = "redis"
  engine_version             = "6.2"
  node_type                  = "cache.m5.large"
  parameter_group_name       = "default.redis6.x.cluster.on"
  automatic_failover_enabled = true
  num_node_groups            = 2
  replicas_per_node_group    = 1
}
`, rName),
	)
}

func testAccReplicationGroupConfig_dataTiering(rName string) string {
	return acctest.ConfigCompose(
		acctest.ConfigVPCWithSubnets(rName, 1),
		fmt.Sprintf(`
resource "aws_elasticache_replication_group" "test" {
  replication_group_id = %[1]q
  description          = "test description"
  node_type            = "cache.r6gd.xlarge"
  data_tiering_enabled = true
  port                 = 6379
  subnet_group_name    = aws_elasticache_subnet_group.test.name
  security_group_ids   = [aws_security_group.test.id]
}

resource "aws_elasticache_subnet_group" "test" {
  name       = %[1]q
  subnet_ids = aws_subnet.test[*].id
}

resource "aws_security_group" "test" {
  name        = %[1]q
  description = "tf-test-security-group-descr"
  vpc_id      = aws_vpc.test.id

  ingress {
    from_port   = -1
    to_port     = -1
    protocol    = "icmp"
    cidr_blocks = ["0.0.0.0/0"]
  }
}
`, rName),
	)
}

func testAccReplicationGroupConfig_dataSourceEngineRedisLogDeliveryConfigurations(rName string, enableClusterMode bool, slowLogDeliveryEnabled bool, slowDeliveryDestination string, slowDeliveryFormat string, engineLogDeliveryEnabled bool, engineDeliveryDestination string, engineLogDeliveryFormat string) string {
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
  policy_name     = "%[1]s"
  depends_on = [
    aws_cloudwatch_log_group.lg
  ]
}

resource "aws_cloudwatch_log_group" "lg" {
  retention_in_days = 1
  name              = "%[1]s"
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
  name        = "%[1]s"
  destination = "extended_s3"
  extended_s3_configuration {
    role_arn   = aws_iam_role.r.arn
    bucket_arn = aws_s3_bucket.b.arn
  }
  lifecycle {
    ignore_changes = [
      tags["LogDeliveryEnabled"],
    ]
  }
}

resource "aws_elasticache_replication_group" "test" {
  replication_group_id       = "%[1]s"
  description                = "test description"
  node_type                  = "cache.t3.small"
  port                       = 6379
  apply_immediately          = true
  maintenance_window         = "tue:06:30-tue:07:30"
  snapshot_window            = "01:00-02:00"
  engine_version             = "6.x"
  parameter_group_name       = tobool("%[2]t") ? "default.redis6.x.cluster.on" : "default.redis6.x"
  automatic_failover_enabled = tobool("%[2]t")
  num_node_groups            = tobool("%[2]t") ? 1 : null
  replicas_per_node_group    = tobool("%[2]t") ? 0 : null

  dynamic "log_delivery_configuration" {
    for_each = tobool("%[3]t") ? [""] : []
    content {
      destination      = ("%[4]s" == "cloudwatch-logs") ? aws_cloudwatch_log_group.lg.name : (("%[4]s" == "kinesis-firehose") ? aws_kinesis_firehose_delivery_stream.ds.name : null)
      destination_type = "%[4]s"
      log_format       = "%[5]s"
      log_type         = "slow-log"
    }
  }
  dynamic "log_delivery_configuration" {
    for_each = tobool("%[6]t") ? [""] : []
    content {
      destination      = ("%[7]s" == "cloudwatch-logs") ? aws_cloudwatch_log_group.lg.name : (("%[7]s" == "kinesis-firehose") ? aws_kinesis_firehose_delivery_stream.ds.name : null)
      destination_type = "%[7]s"
      log_format       = "%[8]s"
      log_type         = "engine-log"
    }
  }
}

data "aws_elasticache_replication_group" "test" {
  replication_group_id = aws_elasticache_replication_group.test.replication_group_id
}
`, rName, enableClusterMode, slowLogDeliveryEnabled, slowDeliveryDestination, slowDeliveryFormat, engineLogDeliveryEnabled, engineDeliveryDestination, engineLogDeliveryFormat)
}

func resourceReplicationGroupDisableAutomaticFailover(ctx context.Context, conn *elasticache.Client, replicationGroupID string, timeout time.Duration) error {
	return resourceReplicationGroupModify(ctx, conn, timeout, &elasticache.ModifyReplicationGroupInput{
		ReplicationGroupId:       aws.String(replicationGroupID),
		ApplyImmediately:         aws.Bool(true),
		AutomaticFailoverEnabled: aws.Bool(false),
		MultiAZEnabled:           aws.Bool(false),
	})
}

func resourceReplicationGroupEnableAutomaticFailover(ctx context.Context, conn *elasticache.Client, replicationGroupID string, multiAZEnabled bool, timeout time.Duration) error {
	return resourceReplicationGroupModify(ctx, conn, timeout, &elasticache.ModifyReplicationGroupInput{
		ReplicationGroupId:       aws.String(replicationGroupID),
		ApplyImmediately:         aws.Bool(true),
		AutomaticFailoverEnabled: aws.Bool(true),
		MultiAZEnabled:           aws.Bool(multiAZEnabled),
	})
}

func resourceReplicationGroupSetPrimaryClusterID(ctx context.Context, conn *elasticache.Client, replicationGroupID, primaryClusterID string, timeout time.Duration) error {
	return resourceReplicationGroupModify(ctx, conn, timeout, &elasticache.ModifyReplicationGroupInput{
		ReplicationGroupId: aws.String(replicationGroupID),
		ApplyImmediately:   aws.Bool(true),
		PrimaryClusterId:   aws.String(primaryClusterID),
	})
}

func resourceReplicationGroupModify(ctx context.Context, conn *elasticache.Client, timeout time.Duration, input *elasticache.ModifyReplicationGroupInput) error {
	_, err := conn.ModifyReplicationGroup(ctx, input)
	if err != nil {
		return fmt.Errorf("error requesting modification: %w", err)
	}

	const (
		delay = 30 * time.Second
	)
	_, err = tfelasticache.WaitReplicationGroupAvailable(ctx, conn, aws.ToString(input.ReplicationGroupId), timeout, delay)
	if err != nil {
		return fmt.Errorf("error waiting for modification: %w", err)
	}
	return nil
}

func formatReplicationGroupClusterID(replicationGroupID string, clusterID int) string {
	return fmt.Sprintf("%s-%03d", replicationGroupID, clusterID)
}
