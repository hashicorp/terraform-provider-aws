package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/elasticache"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/elasticache/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	ReplicationGroupStatusCreating     = "creating"
	ReplicationGroupStatusAvailable    = "available"
	ReplicationGroupStatusModifying    = "modifying"
	ReplicationGroupStatusDeleting     = "deleting"
	ReplicationGroupStatusCreateFailed = "create-failed"
	ReplicationGroupStatusSnapshotting = "snapshotting"

	UserStatusActive    = "active"
	UserStatusDeleting  = "deleting"
	UserStatusModifying = "modifying"
)

// StatusReplicationGroup fetches the Replication Group and its Status
func StatusReplicationGroup(conn *elasticache.ElastiCache, replicationGroupID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		rg, err := finder.FindReplicationGroupByID(conn, replicationGroupID)
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
func StatusReplicationGroupMemberClusters(conn *elasticache.ElastiCache, replicationGroupID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		clusters, err := finder.FindReplicationGroupMemberClustersByID(conn, replicationGroupID)
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
func StatusCacheCluster(conn *elasticache.ElastiCache, cacheClusterID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		c, err := finder.FindCacheClusterByID(conn, cacheClusterID)
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

// StatusGlobalReplicationGroup fetches the Global Replication Group and its Status
func StatusGlobalReplicationGroup(conn *elasticache.ElastiCache, globalReplicationGroupID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		grg, err := finder.FindGlobalReplicationGroupByID(conn, globalReplicationGroupID)
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

// StatusGlobalReplicationGroup fetches the Global Replication Group and its Status
func StatusGlobalReplicationGroupMember(conn *elasticache.ElastiCache, globalReplicationGroupID, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		member, err := finder.FindGlobalReplicationGroupMemberByID(conn, globalReplicationGroupID, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}

		return member, aws.StringValue(member.Status), nil
	}
}

// StatusUser fetches the ElastiCache user and its Status
func StatusUser(conn *elasticache.ElastiCache, userId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		user, err := finder.FindElastiCacheUserByID(conn, userId)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return user, aws.StringValue(user.Status), nil
	}
}
