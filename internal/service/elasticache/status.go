// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticache

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	ReplicationGroupStatusCreating     = "creating"
	ReplicationGroupStatusAvailable    = "available"
	ReplicationGroupStatusModifying    = "modifying"
	ReplicationGroupStatusDeleting     = "deleting"
	ReplicationGroupStatusCreateFailed = "create-failed"
	ReplicationGroupStatusSnapshotting = "snapshotting"
)

// StatusReplicationGroup fetches the Replication Group and its Status
func StatusReplicationGroup(ctx context.Context, conn *elasticache.ElastiCache, replicationGroupID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		rg, err := FindReplicationGroupByID(ctx, conn, replicationGroupID)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}

		return rg, aws.StringValue(rg.Status), nil
	}
}

// StatusReplicationGroupMemberClusters fetches the Replication Group's Member Clusters and either "available" or the first non-"available" status.
// NOTE: This function assumes that the intended end-state is to have all member clusters in "available" status.
func StatusReplicationGroupMemberClusters(ctx context.Context, conn *elasticache.ElastiCache, replicationGroupID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		clusters, err := FindReplicationGroupMemberClustersByID(ctx, conn, replicationGroupID)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}

		status := CacheClusterStatusAvailable
		for _, v := range clusters {
			clusterStatus := aws.StringValue(v.CacheClusterStatus)
			if clusterStatus != CacheClusterStatusAvailable {
				status = clusterStatus
				break
			}
		}
		return clusters, status, nil
	}
}

const (
	CacheClusterStatusAvailable             = "available"
	CacheClusterStatusCreating              = "creating"
	CacheClusterStatusDeleted               = "deleted"
	CacheClusterStatusDeleting              = "deleting"
	CacheClusterStatusIncompatibleNetwork   = "incompatible-network"
	CacheClusterStatusModifying             = "modifying"
	CacheClusterStatusRebootingClusterNodes = "rebooting cluster nodes"
	CacheClusterStatusRestoreFailed         = "restore-failed"
	CacheClusterStatusSnapshotting          = "snapshotting"
)

// StatusCacheCluster fetches the Cache Cluster and its Status
func StatusCacheCluster(ctx context.Context, conn *elasticache.ElastiCache, cacheClusterID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		c, err := FindCacheClusterByID(ctx, conn, cacheClusterID)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}

		return c, aws.StringValue(c.CacheClusterStatus), nil
	}
}

const (
	GlobalReplicationGroupStatusAvailable   = "available"
	GlobalReplicationGroupStatusCreating    = "creating"
	GlobalReplicationGroupStatusModifying   = "modifying"
	GlobalReplicationGroupStatusPrimaryOnly = "primary-only"
	GlobalReplicationGroupStatusDeleting    = "deleting"
	GlobalReplicationGroupStatusDeleted     = "deleted"
)

// statusGlobalReplicationGroup fetches the Global Replication Group and its Status
func statusGlobalReplicationGroup(ctx context.Context, conn *elasticache.ElastiCache, globalReplicationGroupID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		grg, err := FindGlobalReplicationGroupByID(ctx, conn, globalReplicationGroupID)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}

		return grg, aws.StringValue(grg.Status), nil
	}
}

const (
	GlobalReplicationGroupMemberStatusAssociated = "associated"
)

// statusGlobalReplicationGroupMember fetches a Global Replication Group Member and its Status
func statusGlobalReplicationGroupMember(ctx context.Context, conn *elasticache.ElastiCache, globalReplicationGroupID, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		member, err := FindGlobalReplicationGroupMemberByID(ctx, conn, globalReplicationGroupID, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}

		return member, aws.StringValue(member.Status), nil
	}
}
