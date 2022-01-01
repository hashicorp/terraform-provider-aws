package memorydb

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/memorydb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	aclStatusActive    = "active"
	aclStatusCreating  = "creating"
	aclStatusDeleting  = "deleting"
	aclStatusModifying = "modifying"

	clusterStatusAvailable = "available"
	clusterStatusCreating  = "creating"
	clusterStatusDeleting  = "deleting"
	clusterStatusUpdating  = "updating"

	clusterParameterGroupStatusApplying = "applying"
	clusterParameterGroupStatusInSync   = "in-sync"

	clusterShardStatusAvailable = "available"
	clusterShardStatusUpdating  = "updating"

	clusterSnsTopicStatusActive   = "ACTIVE"
	clusterSnsTopicStatusInactive = "INACTIVE"

	userStatusActive    = "active"
	userStatusDeleting  = "deleting"
	userStatusModifying = "modifying"
)

// statusACL fetches the MemoryDB ACL and its status.
func statusACL(ctx context.Context, conn *memorydb.MemoryDB, aclName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		acl, err := FindACLByName(ctx, conn, aclName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return acl, aws.StringValue(acl.Status), nil
	}
}

// statusCluster fetches the MemoryDB Cluster and its status.
func statusCluster(ctx context.Context, conn *memorydb.MemoryDB, clusterName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		Cluster, err := FindClusterByName(ctx, conn, clusterName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return Cluster, aws.StringValue(Cluster.Status), nil
	}
}

// statusClusterParameterGroup fetches the MemoryDB Cluster and its parameter group status.
func statusClusterParameterGroup(ctx context.Context, conn *memorydb.MemoryDB, clusterName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		Cluster, err := FindClusterByName(ctx, conn, clusterName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return Cluster, aws.StringValue(Cluster.ParameterGroupStatus), nil
	}
}

// statusUser fetches the MemoryDB user and its status.
func statusUser(ctx context.Context, conn *memorydb.MemoryDB, userName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		user, err := FindUserByName(ctx, conn, userName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return user, aws.StringValue(user.Status), nil
	}
}
