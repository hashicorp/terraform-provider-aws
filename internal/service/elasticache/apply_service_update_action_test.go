// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package elasticache_test

import (
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// Testing Strategy
//
// The ElastiCache service automatically applies current service updates when a new Cluster or Replication Group is created.
// New service updates are not published frequently, so it is challenging to test a successful application of the action.
// The testing approach is to trigger the action on a previously-applied service update, which results in an error.

func TestAccElastiCacheApplyServiceUpdateAction_replicationGroupID(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ElastiCacheEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		ExternalProviders: map[string]resource.ExternalProvider{
			"time": {
				Source:            "hashicorp/time",
				VersionConstraint: "0.14.0",
			},
		},
		CheckDestroy: acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config:      testAccApplyServiceUpdateActionConfig_replicationGroupID(rName),
				ExpectError: regexache.MustCompile(`The update action is not in a valid status\. Current status:\n\s*complete\.`),
			},
		},
	})
}

func TestAccElastiCacheApplyServiceUpdateAction_cacheClusterID(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.ElastiCacheEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		ExternalProviders: map[string]resource.ExternalProvider{
			"time": {
				Source:            "hashicorp/time",
				VersionConstraint: "0.14.0",
			},
		},
		CheckDestroy: acctest.CheckDestroyNoop,
		Steps: []resource.TestStep{
			{
				Config:      testAccApplyServiceUpdateActionConfig_cacheClusterID(rName),
				ExpectError: regexache.MustCompile(`The update action is not in a valid status\. Current status:\n\s*complete\.`),
			},
		},
	})
}

func testAccApplyServiceUpdateActionConfig_replicationGroupID(rName string) string {
	return acctest.ConfigCompose(
		testAccReplicationGroupConfig_basic_engine(rName, "valkey"), `
action "aws_elasticache_apply_service_update" "test" {
  config {
    replication_group_id = local.update_action.replication_group_id
    service_update_name  = local.update_action.service_update_name
  }
}

locals {
  complete_update_actions = [
    for action in data.aws_elasticache_service_update_actions.test.update_actions : action
    if action.update_action_status == "complete"
  ]

  update_action = local.complete_update_actions[0]
}

resource "terraform_data" "trigger" {
  input = "trigger"
  lifecycle {
    action_trigger {
      events  = [before_create, before_update]
      actions = [action.aws_elasticache_apply_service_update.test]
    }
  }
}

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

func testAccApplyServiceUpdateActionConfig_cacheClusterID(rName string) string {
	return acctest.ConfigCompose(
		testAccClusterConfig_engineRedis(rName), `
action "aws_elasticache_apply_service_update" "test" {
  config {
    cache_cluster_id    = local.update_action.cache_cluster_id
    service_update_name = local.update_action.service_update_name
  }
}

locals {
  complete_update_actions = [
    for action in data.aws_elasticache_service_update_actions.test.update_actions : action
    if action.update_action_status == "complete"
  ]

  update_action = local.complete_update_actions[0]
}

resource "terraform_data" "trigger" {
  input = "trigger"
  lifecycle {
    action_trigger {
      events  = [before_create, before_update]
      actions = [action.aws_elasticache_apply_service_update.test]
    }
  }
}

data "aws_elasticache_service_update_actions" "test" {
  cache_cluster_id = aws_elasticache_cluster.test.cluster_id

  depends_on = [time_sleep.wait]
}

# It takes approximately 10 minutes for Update Actions to be registered after the Replication Group becomes available.
resource "time_sleep" "wait" {
  create_duration = "10m"

  depends_on = [aws_elasticache_cluster.test]
}
`)
}
