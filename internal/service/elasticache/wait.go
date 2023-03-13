package elasticache

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	ReplicationGroupDefaultCreatedTimeout = 60 * time.Minute
	ReplicationGroupDefaultUpdatedTimeout = 40 * time.Minute
	ReplicationGroupDefaultDeletedTimeout = 40 * time.Minute

	replicationGroupAvailableMinTimeout = 10 * time.Second
	replicationGroupAvailableDelay      = 30 * time.Second

	replicationGroupDeletedMinTimeout = 10 * time.Second
	replicationGroupDeletedDelay      = 30 * time.Second

	UserActiveTimeout  = 5 * time.Minute
	UserDeletedTimeout = 5 * time.Minute
)

// WaitReplicationGroupAvailable waits for a ReplicationGroup to return Available
func WaitReplicationGroupAvailable(ctx context.Context, conn *elasticache.ElastiCache, replicationGroupID string, timeout time.Duration) (*elasticache.ReplicationGroup, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			ReplicationGroupStatusCreating,
			ReplicationGroupStatusModifying,
			ReplicationGroupStatusSnapshotting,
		},
		Target:     []string{ReplicationGroupStatusAvailable},
		Refresh:    StatusReplicationGroup(ctx, conn, replicationGroupID),
		Timeout:    timeout,
		MinTimeout: replicationGroupAvailableMinTimeout,
		Delay:      replicationGroupAvailableDelay,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if v, ok := outputRaw.(*elasticache.ReplicationGroup); ok {
		return v, err
	}
	return nil, err
}

// WaitReplicationGroupDeleted waits for a ReplicationGroup to be deleted
func WaitReplicationGroupDeleted(ctx context.Context, conn *elasticache.ElastiCache, replicationGroupID string, timeout time.Duration) (*elasticache.ReplicationGroup, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			ReplicationGroupStatusCreating,
			ReplicationGroupStatusAvailable,
			ReplicationGroupStatusDeleting,
		},
		Target:     []string{},
		Refresh:    StatusReplicationGroup(ctx, conn, replicationGroupID),
		Timeout:    timeout,
		MinTimeout: replicationGroupDeletedMinTimeout,
		Delay:      replicationGroupDeletedDelay,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if v, ok := outputRaw.(*elasticache.ReplicationGroup); ok {
		return v, err
	}
	return nil, err
}

// WaitReplicationGroupMemberClustersAvailable waits for all of a ReplicationGroup's Member Clusters to return Available
func WaitReplicationGroupMemberClustersAvailable(ctx context.Context, conn *elasticache.ElastiCache, replicationGroupID string, timeout time.Duration) ([]*elasticache.CacheCluster, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			CacheClusterStatusCreating,
			CacheClusterStatusDeleting,
			CacheClusterStatusModifying,
		},
		Target:     []string{CacheClusterStatusAvailable},
		Refresh:    StatusReplicationGroupMemberClusters(ctx, conn, replicationGroupID),
		Timeout:    timeout,
		MinTimeout: cacheClusterAvailableMinTimeout,
		Delay:      cacheClusterAvailableDelay,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if v, ok := outputRaw.([]*elasticache.CacheCluster); ok {
		return v, err
	}
	return nil, err
}

const (
	CacheClusterCreatedTimeout = 40 * time.Minute
	CacheClusterUpdatedTimeout = 80 * time.Minute
	CacheClusterDeletedTimeout = 40 * time.Minute

	cacheClusterAvailableMinTimeout = 10 * time.Second
	cacheClusterAvailableDelay      = 30 * time.Second

	cacheClusterDeletedMinTimeout = 10 * time.Second
	cacheClusterDeletedDelay      = 30 * time.Second
)

// waitCacheClusterAvailable waits for a Cache Cluster to return Available
func waitCacheClusterAvailable(ctx context.Context, conn *elasticache.ElastiCache, cacheClusterID string, timeout time.Duration) (*elasticache.CacheCluster, error) { //nolint:unparam
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			CacheClusterStatusCreating,
			CacheClusterStatusModifying,
			CacheClusterStatusSnapshotting,
			CacheClusterStatusRebootingClusterNodes,
		},
		Target:     []string{CacheClusterStatusAvailable},
		Refresh:    StatusCacheCluster(ctx, conn, cacheClusterID),
		Timeout:    timeout,
		MinTimeout: cacheClusterAvailableMinTimeout,
		Delay:      cacheClusterAvailableDelay,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if v, ok := outputRaw.(*elasticache.CacheCluster); ok {
		return v, err
	}
	return nil, err
}

// WaitCacheClusterDeleted waits for a Cache Cluster to be deleted
func WaitCacheClusterDeleted(ctx context.Context, conn *elasticache.ElastiCache, cacheClusterID string, timeout time.Duration) (*elasticache.CacheCluster, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			CacheClusterStatusCreating,
			CacheClusterStatusAvailable,
			CacheClusterStatusModifying,
			CacheClusterStatusDeleting,
			CacheClusterStatusIncompatibleNetwork,
			CacheClusterStatusRestoreFailed,
			CacheClusterStatusSnapshotting,
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
	stateConf := &resource.StateChangeConf{
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
	stateConf := &resource.StateChangeConf{
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
	stateConf := &resource.StateChangeConf{
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

// WaitUserActive waits for an ElastiCache user to reach an active state after modifications
func WaitUserActive(ctx context.Context, conn *elasticache.ElastiCache, userId string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{UserStatusModifying},
		Target:  []string{UserStatusActive},
		Refresh: StatusUser(ctx, conn, userId),
		Timeout: UserActiveTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

// WaitUserDeleted waits for an ElastiCache user to be deleted
func WaitUserDeleted(ctx context.Context, conn *elasticache.ElastiCache, userId string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{UserStatusDeleting},
		Target:  []string{},
		Refresh: StatusUser(ctx, conn, userId),
		Timeout: UserDeletedTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}
