// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticache

import (
	"context"
	"errors"

	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// customizeDiffValidateClusterAZMode validates that `num_cache_nodes` is greater than 1 when `az_mode` is "cross-az"
func customizeDiffValidateClusterAZMode(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
	if v, ok := diff.GetOk("az_mode"); !ok || v.(string) != string(awstypes.AZModeCrossAz) {
		return nil
	}
	if v, ok := diff.GetOk("num_cache_nodes"); !ok || v.(int) != 1 {
		return nil
	}

	return errors.New(`az_mode "cross-az" is not supported with num_cache_nodes = 1`)
}

// customizeDiffValidateClusterNumCacheNodes validates that `num_cache_nodes` is 1 when `engine` is "redis"
func customizeDiffValidateClusterNumCacheNodes(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
	if v, ok := diff.GetOk(names.AttrEngine); !ok || v.(string) == engineMemcached {
		return nil
	}

	if v, ok := diff.GetOk("num_cache_nodes"); !ok || v.(int) == 1 {
		return nil
	}
	return errors.New(`engine "redis" does not support num_cache_nodes > 1`)
}

// customizeDiffClusterMemcachedNodeType causes re-creation when `node_type` is changed and `engine` is "memcached"
func customizeDiffClusterMemcachedNodeType(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
	// Engine memcached does not currently support vertical scaling
	// https://docs.aws.amazon.com/AmazonElastiCache/latest/mem-ug/Scaling.html#Scaling.Memcached.Vertically
	if diff.Id() == "" || !diff.HasChange("node_type") {
		return nil
	}
	if v, ok := diff.GetOk(names.AttrEngine); !ok || v.(string) == engineRedis {
		return nil
	}
	return diff.ForceNew("node_type")
}

// customizeDiffValidateClusterMemcachedSnapshotIdentifier validates that `final_snapshot_identifier` is not set when `engine` is "memcached"
func customizeDiffValidateClusterMemcachedSnapshotIdentifier(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
	if v, ok := diff.GetOk(names.AttrEngine); !ok || v.(string) == engineRedis {
		return nil
	}
	if _, ok := diff.GetOk(names.AttrFinalSnapshotIdentifier); !ok {
		return nil
	}
	return errors.New(`engine "memcached" does not support final_snapshot_identifier`)
}

// customizeDiffValidateReplicationGroupAutomaticFailover validates that `automatic_failover_enabled` is set when `multi_az_enabled` is true
func customizeDiffValidateReplicationGroupAutomaticFailover(_ context.Context, diff *schema.ResourceDiff, v interface{}) error {
	if v := diff.Get("multi_az_enabled").(bool); !v {
		return nil
	}
	if v := diff.Get("automatic_failover_enabled").(bool); !v {
		return errors.New(`automatic_failover_enabled must be true if multi_az_enabled is true`)
	}
	return nil
}
