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

const (
	globalReplicationGroupDefaultCreatedTimeout = 60 * time.Minute
	globalReplicationGroupDefaultUpdatedTimeout = 60 * time.Minute
	globalReplicationGroupDefaultDeletedTimeout = 20 * time.Minute

	globalReplicationGroupAvailableMinTimeout = 10 * time.Second
	globalReplicationGroupAvailableDelay      = 30 * time.Second

	globalReplicationGroupDeletedMinTimeout = 10 * time.Second
	globalReplicationGroupDeletedDelay      = 30 * time.Second
)

// waitGlobalReplicationGroupAvailable waits for a Global Replication Group to be available,
// with status either "available" or "primary-only"
func waitGlobalReplicationGroupAvailable(ctx context.Context, conn *elasticache.ElastiCache, globalReplicationGroupID string, timeout time.Duration) (*elasticache.GlobalReplicationGroup, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{GlobalReplicationGroupStatusCreating, GlobalReplicationGroupStatusModifying},
		Target:     []string{GlobalReplicationGroupStatusAvailable, GlobalReplicationGroupStatusPrimaryOnly},
		Refresh:    statusGlobalReplicationGroup(ctx, conn, globalReplicationGroupID),
		Timeout:    timeout,
		MinTimeout: globalReplicationGroupAvailableMinTimeout,
		Delay:      globalReplicationGroupAvailableDelay,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if v, ok := outputRaw.(*elasticache.GlobalReplicationGroup); ok {
		return v, err
	}
	return nil, err
}

// waitGlobalReplicationGroupDeleted waits for a Global Replication Group to be deleted
func waitGlobalReplicationGroupDeleted(ctx context.Context, conn *elasticache.ElastiCache, globalReplicationGroupID string, timeout time.Duration) (*elasticache.GlobalReplicationGroup, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			GlobalReplicationGroupStatusAvailable,
			GlobalReplicationGroupStatusPrimaryOnly,
			GlobalReplicationGroupStatusModifying,
			GlobalReplicationGroupStatusDeleting,
		},
		Target:     []string{},
		Refresh:    statusGlobalReplicationGroup(ctx, conn, globalReplicationGroupID),
		Timeout:    timeout,
		MinTimeout: globalReplicationGroupDeletedMinTimeout,
		Delay:      globalReplicationGroupDeletedDelay,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if v, ok := outputRaw.(*elasticache.GlobalReplicationGroup); ok {
		return v, err
	}
	return nil, err
}

const (
	// GlobalReplicationGroupDisassociationReadyTimeout specifies how long to wait for a global replication group
	// to be in a valid state before disassociating
	GlobalReplicationGroupDisassociationReadyTimeout = 45 * time.Minute

	// globalReplicationGroupDisassociationTimeout specifies how long to wait for the actual disassociation
	globalReplicationGroupDisassociationTimeout = 20 * time.Minute

	globalReplicationGroupDisassociationMinTimeout = 10 * time.Second
	globalReplicationGroupDisassociationDelay      = 30 * time.Second
)

func waitGlobalReplicationGroupMemberDetached(ctx context.Context, conn *elasticache.ElastiCache, globalReplicationGroupID, id string) (*elasticache.GlobalReplicationGroupMember, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			GlobalReplicationGroupMemberStatusAssociated,
		},
		Target:     []string{},
		Refresh:    statusGlobalReplicationGroupMember(ctx, conn, globalReplicationGroupID, id),
		Timeout:    globalReplicationGroupDisassociationTimeout,
		MinTimeout: globalReplicationGroupDisassociationMinTimeout,
		Delay:      globalReplicationGroupDisassociationDelay,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if v, ok := outputRaw.(*elasticache.GlobalReplicationGroupMember); ok {
		return v, err
	}
	return nil, err
}
