// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticache

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

const (
	CacheClusterCreatedTimeout = 40 * time.Minute
	CacheClusterUpdatedTimeout = 80 * time.Minute
	CacheClusterDeletedTimeout = 40 * time.Minute

	cacheClusterDeletedMinTimeout = 10 * time.Second
	cacheClusterDeletedDelay      = 30 * time.Second
)

// waitCacheClusterAvailable waits for a Cache Cluster to return Available
func waitCacheClusterAvailable(ctx context.Context, conn *elasticache.ElastiCache, cacheClusterID string, timeout time.Duration) (*elasticache.CacheCluster, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			cacheClusterStatusCreating,
			cacheClusterStatusModifying,
			cacheClusterStatusSnapshotting,
			cacheClusterStatusRebootingClusterNodes,
		},
		Target:     []string{cacheClusterStatusAvailable},
		Refresh:    StatusCacheCluster(ctx, conn, cacheClusterID),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if v, ok := outputRaw.(*elasticache.CacheCluster); ok {
		return v, err
	}
	return nil, err
}

// WaitCacheClusterDeleted waits for a Cache Cluster to be deleted
func WaitCacheClusterDeleted(ctx context.Context, conn *elasticache.ElastiCache, cacheClusterID string, timeout time.Duration) (*elasticache.CacheCluster, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			cacheClusterStatusCreating,
			cacheClusterStatusAvailable,
			cacheClusterStatusModifying,
			cacheClusterStatusDeleting,
			cacheClusterStatusIncompatibleNetwork,
			cacheClusterStatusRestoreFailed,
			cacheClusterStatusSnapshotting,
		},
		Target:     []string{},
		Refresh:    StatusCacheCluster(ctx, conn, cacheClusterID),
		Timeout:    timeout,
		MinTimeout: cacheClusterDeletedMinTimeout,
		Delay:      cacheClusterDeletedDelay,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if v, ok := outputRaw.(*elasticache.CacheCluster); ok {
		return v, err
	}
	return nil, err
}
