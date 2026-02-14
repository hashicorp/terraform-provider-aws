// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package elasticache_test

import (
	"context"
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfelasticache "github.com/hashicorp/terraform-provider-aws/internal/service/elasticache"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccElastiCacheServerlessCache_basicRedis(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elasticache_serverless_cache.test"
	var serverlessElasticCache awstypes.ServerlessCache

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckServerlessCacheDestroy(ctx, t),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccServerlessCacheConfig_basicRedis(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServerlessCacheExists(ctx, t, resourceName, &serverlessElasticCache),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "elasticache", "serverlesscache:{name}"),
					resource.TestCheckResourceAttrSet(resourceName, "cache_usage_limits.#"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreateTime),
					resource.TestCheckResourceAttrSet(resourceName, "daily_snapshot_time"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint.#"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngine),
					resource.TestCheckResourceAttrSet(resourceName, "full_engine_version"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, "reader_endpoint.#"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(resourceName, "subnet_ids.#"),
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

func TestAccElastiCacheServerlessCache_basicValkey(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elasticache_serverless_cache.test"
	var serverlessElasticCache awstypes.ServerlessCache

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckServerlessCacheDestroy(ctx, t),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccServerlessCacheConfig_basicValkey(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServerlessCacheExists(ctx, t, resourceName, &serverlessElasticCache),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "elasticache", "serverlesscache:{name}"),
					resource.TestCheckResourceAttrSet(resourceName, "cache_usage_limits.#"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreateTime),
					resource.TestCheckResourceAttrSet(resourceName, "daily_snapshot_time"),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint.#"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngine),
					resource.TestCheckResourceAttrSet(resourceName, "full_engine_version"),
					resource.TestCheckResourceAttrSet(resourceName, "reader_endpoint.#"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(resourceName, "subnet_ids.#"),
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

func TestAccElastiCacheServerlessCache_full(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elasticache_serverless_cache.test"
	var serverlessElasticCache awstypes.ServerlessCache

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckServerlessCacheDestroy(ctx, t),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccServerlessCacheConfig_full(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServerlessCacheExists(ctx, t, resourceName, &serverlessElasticCache),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "elasticache", "serverlesscache:{name}"),
					resource.TestCheckResourceAttrSet(resourceName, "cache_usage_limits.#"),
					resource.TestCheckResourceAttr(resourceName, "cache_usage_limits.0.data_storage.0.maximum", "10"),
					resource.TestCheckResourceAttr(resourceName, "cache_usage_limits.0.data_storage.0.minimum", "1"),
					resource.TestCheckResourceAttr(resourceName, "cache_usage_limits.0.data_storage.0.unit", "GB"),
					resource.TestCheckResourceAttr(resourceName, "cache_usage_limits.0.ecpu_per_second.0.maximum", "10000"),
					resource.TestCheckResourceAttr(resourceName, "cache_usage_limits.0.ecpu_per_second.0.minimum", "1000"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreateTime),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint.#"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngine),
					resource.TestCheckResourceAttrSet(resourceName, "full_engine_version"),
					resource.TestCheckResourceAttrSet(resourceName, "reader_endpoint.#"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(resourceName, "subnet_ids.#"),
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

func TestAccElastiCacheServerlessCache_fullRedis(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elasticache_serverless_cache.test"
	var serverlessElasticCache awstypes.ServerlessCache

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckServerlessCacheDestroy(ctx, t),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccServerlessCacheConfig_fullRedis(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServerlessCacheExists(ctx, t, resourceName, &serverlessElasticCache),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "elasticache", "serverlesscache:{name}"),
					resource.TestCheckResourceAttrSet(resourceName, "cache_usage_limits.#"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreateTime),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint.#"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngine),
					resource.TestCheckResourceAttrSet(resourceName, "full_engine_version"),
					resource.TestCheckResourceAttrSet(resourceName, "reader_endpoint.#"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(resourceName, "subnet_ids.#"),
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

func TestAccElastiCacheServerlessCache_redisUpdateWithUserGroup(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elasticache_serverless_cache.test"
	var serverlessElasticCache awstypes.ServerlessCache

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckServerlessCacheDestroy(ctx, t),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccServerlessCacheConfig_redisUpdateWithUserGroup(rName, "test description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServerlessCacheExists(ctx, t, resourceName, &serverlessElasticCache),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "elasticache", "serverlesscache:{name}"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test description"),
				),
			},
			{
				Config: testAccServerlessCacheConfig_redisUpdateWithUserGroup(rName, "test description updated"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServerlessCacheExists(ctx, t, resourceName, &serverlessElasticCache),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "elasticache", "serverlesscache:{name}"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test description updated"),
				),
			},
		},
	})
}

func TestAccElastiCacheServerlessCache_updateWithUserGroupRemoval(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	userId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	userGroupId := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elasticache_serverless_cache.test"
	userResourceName := "aws_elasticache_user.test"
	userGroupResourceName := "aws_elasticache_user_group.test"

	var serverlessElasticCache awstypes.ServerlessCache

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckServerlessCacheDestroy(ctx, t),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccServerlessCacheConfig_updateWithUserIdUserGroup(rName, "test description", userId, userGroupId, tfelasticache.EngineRedis, tfelasticache.EngineRedis, tfelasticache.EngineRedis),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServerlessCacheExists(ctx, t, resourceName, &serverlessElasticCache),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "elasticache", "serverlesscache:{name}"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, tfelasticache.EngineRedis),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test description"),
					resource.TestCheckResourceAttr(resourceName, "user_group_id", userGroupId),
					resource.TestCheckResourceAttr(userResourceName, names.AttrEngine, tfelasticache.EngineRedis),
					resource.TestCheckResourceAttr(userGroupResourceName, names.AttrEngine, tfelasticache.EngineRedis),
				),
			},
			{
				Config: testAccServerlessCacheConfig_updateWithUserIdUserGroup(rName, "test description", userId, userGroupId, tfelasticache.EngineRedis, tfelasticache.EngineRedis, tfelasticache.EngineValkey),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServerlessCacheExists(ctx, t, resourceName, &serverlessElasticCache),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "elasticache", "serverlesscache:{name}"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, tfelasticache.EngineValkey),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test description"),
					resource.TestCheckResourceAttr(resourceName, "user_group_id", userGroupId),
					resource.TestCheckResourceAttr(userResourceName, names.AttrEngine, tfelasticache.EngineRedis),
					resource.TestCheckResourceAttr(userGroupResourceName, names.AttrEngine, tfelasticache.EngineRedis),
				),
			},
			{
				Config: testAccServerlessCacheConfig_updateWithUserIdUserGroup(rName, "test description", userId, userGroupId, tfelasticache.EngineRedis, tfelasticache.EngineValkey, tfelasticache.EngineValkey),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServerlessCacheExists(ctx, t, resourceName, &serverlessElasticCache),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "elasticache", "serverlesscache:{name}"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, tfelasticache.EngineValkey),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test description"),
					resource.TestCheckResourceAttr(resourceName, "user_group_id", userGroupId),
					resource.TestCheckResourceAttr(userResourceName, names.AttrEngine, tfelasticache.EngineRedis),
					resource.TestCheckResourceAttr(userGroupResourceName, names.AttrEngine, tfelasticache.EngineValkey),
				),
			},
			{
				Config: testAccServerlessCacheConfig_updateWithUserIdUserGroup(rName, "test description", userId, userGroupId, tfelasticache.EngineValkey, tfelasticache.EngineValkey, tfelasticache.EngineValkey),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServerlessCacheExists(ctx, t, resourceName, &serverlessElasticCache),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "elasticache", "serverlesscache:{name}"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, tfelasticache.EngineValkey),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test description"),
					resource.TestCheckResourceAttr(resourceName, "user_group_id", userGroupId),
					resource.TestCheckResourceAttr(userResourceName, names.AttrEngine, tfelasticache.EngineValkey),
					resource.TestCheckResourceAttr(userGroupResourceName, names.AttrEngine, tfelasticache.EngineValkey),
				),
			},
			{
				Config: testAccServerlessCacheConfig_updateWithUserIdUserGroupRemoval(rName, "test description", userId, userGroupId, tfelasticache.EngineValkey),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServerlessCacheExists(ctx, t, resourceName, &serverlessElasticCache),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "elasticache", "serverlesscache:{name}"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, tfelasticache.EngineValkey),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test description"),
					resource.TestCheckNoResourceAttr(resourceName, "user_group_id"),
				),
			},
			{
				Config: testAccServerlessCacheConfig_redisUpdateWithUserIdUserGroupDelete(rName, "test description", tfelasticache.EngineValkey),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServerlessCacheExists(ctx, t, resourceName, &serverlessElasticCache),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "elasticache", "serverlesscache:{name}"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, tfelasticache.EngineValkey),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "test description"),
					resource.TestCheckNoResourceAttr(resourceName, "user_group_id"),
				),
			},
		},
	})
}

func TestAccElastiCacheServerlessCache_fullValkey(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elasticache_serverless_cache.test"
	var serverlessElasticCache awstypes.ServerlessCache

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckServerlessCacheDestroy(ctx, t),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccServerlessCacheConfig_fullValkey(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServerlessCacheExists(ctx, t, resourceName, &serverlessElasticCache),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "elasticache", "serverlesscache:{name}"),
					resource.TestCheckResourceAttrSet(resourceName, "cache_usage_limits.#"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreateTime),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint.#"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngine),
					resource.TestCheckResourceAttrSet(resourceName, "full_engine_version"),
					resource.TestCheckResourceAttrSet(resourceName, "reader_endpoint.#"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(resourceName, "subnet_ids.#"),
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

func TestAccElastiCacheServerlessCache_description(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	descriptionOld := "Memcached Serverless Cluster"
	descriptionNew := "Memcached Serverless Cluster updated"
	resourceName := "aws_elasticache_serverless_cache.test"
	var serverlessElasticCache awstypes.ServerlessCache

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckServerlessCacheDestroy(ctx, t),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccServerlessCacheConfig_description(rName, descriptionOld),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServerlessCacheExists(ctx, t, resourceName, &serverlessElasticCache),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "elasticache", "serverlesscache:{name}"),
					resource.TestCheckResourceAttrSet(resourceName, "cache_usage_limits.#"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreateTime),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, descriptionOld),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint.#"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngine),
					resource.TestCheckResourceAttrSet(resourceName, "full_engine_version"),
					resource.TestCheckResourceAttrSet(resourceName, "reader_endpoint.#"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(resourceName, "subnet_ids.#"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccServerlessCacheConfig_description(rName, descriptionNew),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServerlessCacheExists(ctx, t, resourceName, &serverlessElasticCache),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "elasticache", "serverlesscache:{name}"),
					resource.TestCheckResourceAttrSet(resourceName, "cache_usage_limits.#"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreateTime),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, descriptionNew),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint.#"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngine),
					resource.TestCheckResourceAttrSet(resourceName, "full_engine_version"),
					resource.TestCheckResourceAttrSet(resourceName, "reader_endpoint.#"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(resourceName, "subnet_ids.#"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func TestAccElastiCacheServerlessCache_cacheUsageLimits(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	descriptionOld := "Memcached Serverless Cluster"
	descriptionNew := "Memcached Serverless Cluster updated"
	resourceName := "aws_elasticache_serverless_cache.test"
	var v awstypes.ServerlessCache

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckServerlessCacheDestroy(ctx, t),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccServerlessCacheConfig_cacheUsageLimits(rName, descriptionOld, 1, 1000),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServerlessCacheExists(ctx, t, resourceName, &v),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "elasticache", "serverlesscache:{name}"),
					resource.TestCheckResourceAttrSet(resourceName, "cache_usage_limits.#"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreateTime),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, descriptionOld),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint.#"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngine),
					resource.TestCheckResourceAttrSet(resourceName, "full_engine_version"),
					resource.TestCheckResourceAttrSet(resourceName, "reader_endpoint.#"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(resourceName, "subnet_ids.#"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccServerlessCacheConfig_cacheUsageLimits(rName, descriptionOld, 2, 1000),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServerlessCacheExists(ctx, t, resourceName, &v),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "elasticache", "serverlesscache:{name}"),
					resource.TestCheckResourceAttrSet(resourceName, "cache_usage_limits.#"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreateTime),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, descriptionOld),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint.#"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngine),
					resource.TestCheckResourceAttrSet(resourceName, "full_engine_version"),
					resource.TestCheckResourceAttrSet(resourceName, "reader_endpoint.#"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(resourceName, "subnet_ids.#"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				Config: testAccServerlessCacheConfig_cacheUsageLimits(rName, descriptionNew, 2, 1000),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServerlessCacheExists(ctx, t, resourceName, &v),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "elasticache", "serverlesscache:{name}"),
					resource.TestCheckResourceAttrSet(resourceName, "cache_usage_limits.#"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreateTime),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, descriptionNew),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint.#"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngine),
					resource.TestCheckResourceAttrSet(resourceName, "full_engine_version"),
					resource.TestCheckResourceAttrSet(resourceName, "reader_endpoint.#"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(resourceName, "subnet_ids.#"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				Config: testAccServerlessCacheConfig_cacheUsageLimits(rName, descriptionNew, 2, 1010),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServerlessCacheExists(ctx, t, resourceName, &v),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "elasticache", "serverlesscache:{name}"),
					resource.TestCheckResourceAttrSet(resourceName, "cache_usage_limits.#"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrCreateTime),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, descriptionNew),
					resource.TestCheckResourceAttrSet(resourceName, "endpoint.#"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrEngine),
					resource.TestCheckResourceAttrSet(resourceName, "full_engine_version"),
					resource.TestCheckResourceAttrSet(resourceName, "reader_endpoint.#"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrSet(resourceName, "subnet_ids.#"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func TestAccElastiCacheServerlessCache_engine(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elasticache_serverless_cache.test"
	var v awstypes.ServerlessCache

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckServerlessCacheDestroy(ctx, t),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccServerlessCacheConfig_engine(rName, tfelasticache.EngineRedis),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServerlessCacheExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, tfelasticache.EngineRedis),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: testAccServerlessCacheConfig_engine(rName, tfelasticache.EngineValkey),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServerlessCacheExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, tfelasticache.EngineValkey),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				Config: testAccServerlessCacheConfig_engine(rName, tfelasticache.EngineRedis),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServerlessCacheExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, tfelasticache.EngineRedis),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
		},
	})
}

func TestAccElastiCacheServerlessCache_valkeyMajorEngineVersion(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elasticache_serverless_cache.test"
	var v awstypes.ServerlessCache

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckServerlessCacheDestroy(ctx, t),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccServerlessCacheConfig_majorEngineVersion(rName, tfelasticache.EngineValkey, "7"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServerlessCacheExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, tfelasticache.EngineValkey),
					resource.TestCheckResourceAttr(resourceName, "major_engine_version", "7"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
			{
				Config: testAccServerlessCacheConfig_majorEngineVersion(rName, tfelasticache.EngineValkey, "8"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServerlessCacheExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, tfelasticache.EngineValkey),
					resource.TestCheckResourceAttr(resourceName, "major_engine_version", "8"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionUpdate),
					},
				},
			},
			{
				Config: testAccServerlessCacheConfig_majorEngineVersion(rName, tfelasticache.EngineValkey, "7"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServerlessCacheExists(ctx, t, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrEngine, tfelasticache.EngineValkey),
					resource.TestCheckResourceAttr(resourceName, "major_engine_version", "7"),
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
			},
		},
	})
}

func TestAccElastiCacheServerlessCache_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elasticache_serverless_cache.test"
	var serverlessElasticCache awstypes.ServerlessCache

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckServerlessCacheDestroy(ctx, t),

		Steps: []resource.TestStep{
			{
				Config: testAccServerlessCacheConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckServerlessCacheExists(ctx, t, resourceName, &serverlessElasticCache),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tfelasticache.ResourceServerlessCache, resourceName),
				),
				ExpectNonEmptyPlan: true,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PostApplyPostRefresh: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction(resourceName, plancheck.ResourceActionCreate),
					},
				},
			},
		},
	})
}

func TestAccElastiCacheServerlessCache_tags(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elasticache_serverless_cache.test"
	var serverlessElasticCache awstypes.ServerlessCache

	tags1 := `
  tags = {
    key1 = "value1"
  }
`
	tags2 := `
  tags = {
    key1 = "value1"
    key2 = "value2"
  }
`
	tags3 := `
  tags = {
    key2 = "value2"
  }
`
	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy: resource.ComposeAggregateTestCheckFunc(
			testAccCheckServerlessCacheDestroy(ctx, t),
		),
		Steps: []resource.TestStep{
			{
				Config: testAccServerlessCacheConfig_tags(rName, tags1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerlessCacheExists(ctx, t, resourceName, &serverlessElasticCache),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				Config: testAccServerlessCacheConfig_tags(rName, tags2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerlessCacheExists(ctx, t, resourceName, &serverlessElasticCache),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccServerlessCacheConfig_tags(rName, tags3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerlessCacheExists(ctx, t, resourceName, &serverlessElasticCache),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckServerlessCacheExists(ctx context.Context, t *testing.T, n string, v *awstypes.ServerlessCache) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).ElastiCacheClient(ctx)

		output, err := tfelasticache.FindServerlessCacheByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckServerlessCacheDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).ElastiCacheClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_elasticache_serverless_cache" {
				continue
			}

			_, err := tfelasticache.FindServerlessCacheByID(ctx, conn, rs.Primary.ID)
			if retry.NotFound(err) {
				continue
			}
			if err != nil {
				return err
			}

			return fmt.Errorf("ElastiCache Serverless Cache (%s) still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccServerlessCacheConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_serverless_cache" "test" {
  engine = "memcached"
  name   = %[1]q
}
`, rName)
}

func testAccServerlessCacheConfig_basicRedis(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_serverless_cache" "test" {
  engine = "redis"
  name   = %[1]q
}
`, rName)
}

func testAccServerlessCacheConfig_basicValkey(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_serverless_cache" "test" {
  engine = "valkey"
  name   = %[1]q
}
`, rName)
}

func testAccServerlessCacheConfig_engine(rName, engine string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_serverless_cache" "test" {
  name   = %[1]q
  engine = %[2]q
}
`, rName, engine)
}

func testAccServerlessCacheConfig_majorEngineVersion(rName, engine, majorEngineVersion string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_serverless_cache" "test" {
  name                 = %[1]q
  engine               = %[2]q
  major_engine_version = %[3]q
}
`, rName, engine, majorEngineVersion)
}

func testAccServerlessCacheConfig_full(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_elasticache_serverless_cache" "test" {
  engine = "memcached"
  name   = %[1]q

  cache_usage_limits {
    data_storage {
      maximum = 10
      minimum = 1
      unit    = "GB"
    }
    ecpu_per_second {
      maximum = 10000
      minimum = 1000
    }
  }

  description          = "Test Full Memcached Attributes"
  kms_key_id           = aws_kms_key.test.arn
  major_engine_version = "1.6"
  security_group_ids   = [aws_security_group.test.id]
  subnet_ids           = aws_subnet.test[*].id
  tags = {
    Name = %[1]q
  }
}

resource "aws_kms_key" "test" {
  description             = "tf-test-cmk-kms-key-id"
  deletion_window_in_days = 7
  enable_key_rotation     = true
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
}
`, rName))
}

func testAccServerlessCacheConfig_fullRedis(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_elasticache_serverless_cache" "test" {
  engine = "redis"
  name   = %[1]q

  cache_usage_limits {
    data_storage {
      maximum = 10
      unit    = "GB"
    }
    ecpu_per_second {
      maximum = 1000
    }
  }

  daily_snapshot_time      = "09:00"
  description              = "Test Full Redis Attributes"
  kms_key_id               = aws_kms_key.test.arn
  major_engine_version     = "7"
  snapshot_retention_limit = 1
  security_group_ids       = [aws_security_group.test.id]
  subnet_ids               = aws_subnet.test[*].id

  tags = {
    Name = %[1]q
  }
}

resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

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
`, rName))
}

func testAccServerlessCacheConfig_redisUpdateWithUserGroup(rName, description string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_elasticache_user" "test" {
  user_id       = "testuserid"
  user_name     = "default"
  access_string = "on ~* +@all"
  engine        = "REDIS"
  passwords     = ["password123456789"]
}

resource "aws_elasticache_user_group" "test" {
  engine        = "REDIS"
  user_group_id = "usergroupid"
  user_ids      = [aws_elasticache_user.test.user_id]
}

resource "aws_elasticache_serverless_cache" "test" {
  engine = "redis"
  name   = %[1]q
  cache_usage_limits {
    data_storage {
      maximum = 12
      unit    = "GB"
    }
    ecpu_per_second {
      maximum = 5100
    }
  }
  daily_snapshot_time      = "09:00"
  description              = %[2]q
  major_engine_version     = "7"
  snapshot_retention_limit = 1
  user_group_id            = aws_elasticache_user_group.test.id
}
`, rName, description))
}

func testAccServerlessCacheConfig_updateWithUserIdUserGroup(rName, description string, userId string, userGroupId string, userEngine string, userGroupEngine string, engine string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_elasticache_user" "test" {
  user_id       = %[3]q
  user_name     = "default"
  access_string = "on ~* +@all"
  engine        = %[5]q
  passwords     = ["password123456789"]
}

resource "aws_elasticache_user_group" "test" {
  engine        = %[6]q
  user_group_id = %[4]q
  user_ids      = [aws_elasticache_user.test.user_id]
}

resource "aws_elasticache_serverless_cache" "test" {
  engine = %[7]q
  name   = %[1]q
  cache_usage_limits {
    data_storage {
      maximum = 12
      unit    = "GB"
    }
    ecpu_per_second {
      maximum = 5100
    }
  }
  daily_snapshot_time      = "09:00"
  description              = %[2]q
  major_engine_version     = "7"
  snapshot_retention_limit = 1
  user_group_id            = aws_elasticache_user_group.test.id
}
`, rName, description, userId, userGroupId, userEngine, userGroupEngine, engine))
}

func testAccServerlessCacheConfig_updateWithUserIdUserGroupRemoval(rName, description string, userId string, userGroupId string, engine string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_elasticache_user" "test" {
  user_id       = %[3]q
  user_name     = "default"
  access_string = "on ~* +@all"
  engine        = "redis"
  passwords     = ["password123456789"]
}

resource "aws_elasticache_user_group" "test" {
  engine        = "redis"
  user_group_id = %[4]q
  user_ids      = [aws_elasticache_user.test.user_id]
}

resource "aws_elasticache_serverless_cache" "test" {
  engine = %[5]q
  name   = %[1]q
  cache_usage_limits {
    data_storage {
      maximum = 12
      unit    = "GB"
    }
    ecpu_per_second {
      maximum = 5100
    }
  }
  daily_snapshot_time      = "09:00"
  description              = %[2]q
  major_engine_version     = "7"
  snapshot_retention_limit = 1
}
`, rName, description, userId, userGroupId, engine))
}

func testAccServerlessCacheConfig_redisUpdateWithUserIdUserGroupDelete(rName, description string, engine string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_elasticache_serverless_cache" "test" {
  engine = %[3]q
  name   = %[1]q
  cache_usage_limits {
    data_storage {
      maximum = 12
      unit    = "GB"
    }
    ecpu_per_second {
      maximum = 5100
    }
  }
  daily_snapshot_time      = "09:00"
  description              = %[2]q
  major_engine_version     = "7"
  snapshot_retention_limit = 1
}
`, rName, description, engine))
}

func testAccServerlessCacheConfig_fullValkey(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigVPCWithSubnets(rName, 2), fmt.Sprintf(`
resource "aws_elasticache_serverless_cache" "test" {
  engine = "valkey"
  name   = %[1]q

  cache_usage_limits {
    data_storage {
      maximum = 10
      unit    = "GB"
    }
    ecpu_per_second {
      maximum = 1000
    }
  }

  daily_snapshot_time      = "09:00"
  description              = "Test Full Valkey Attributes"
  kms_key_id               = aws_kms_key.test.arn
  major_engine_version     = "7"
  snapshot_retention_limit = 1
  security_group_ids       = [aws_security_group.test.id]
  subnet_ids               = aws_subnet.test[*].id

  tags = {
    Name = %[1]q
  }
}

resource "aws_kms_key" "test" {
  description             = %[1]q
  deletion_window_in_days = 7
  enable_key_rotation     = true
}

resource "aws_security_group" "test" {
  name   = %[1]q
  vpc_id = aws_vpc.test.id

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
`, rName))
}

func testAccServerlessCacheConfig_description(rName, desc string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_serverless_cache" "test" {
  engine      = "memcached"
  name        = %[1]q
  description = %[2]q
}
`, rName, desc)
}

func testAccServerlessCacheConfig_cacheUsageLimits(rName, desc string, d1, d2 int) string {
	return fmt.Sprintf(`
resource "aws_elasticache_serverless_cache" "test" {
  engine      = "memcached"
  name        = %[1]q
  description = %[2]q
  cache_usage_limits {
    data_storage {
      maximum = %[3]d
      unit    = "GB"
    }
    ecpu_per_second {
      maximum = %[4]d
    }
  }
}
`, rName, desc, d1, d2)
}

func testAccServerlessCacheConfig_tags(rName, tags string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_serverless_cache" "test" {
  engine = "memcached"
  name   = %[1]q

%[2]s
}
`, rName, tags)
}
