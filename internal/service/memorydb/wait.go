package memorydb

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/memorydb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	aclActiveTimeout  = 5 * time.Minute
	aclDeletedTimeout = 5 * time.Minute

	clusterAvailableTimeout = 120 * time.Minute
	clusterDeletedTimeout   = 120 * time.Minute

	clusterParameterGroupInSyncTimeout = 60 * time.Minute

	clusterSecurityGroupsActiveTimeout = 10 * time.Minute

	userActiveTimeout  = 5 * time.Minute
	userDeletedTimeout = 5 * time.Minute
)

// waitACLActive waits for MemoryDB ACL to reach an active state after modifications.
func waitACLActive(ctx context.Context, conn *memorydb.MemoryDB, aclId string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{aclStatusCreating, aclStatusModifying},
		Target:  []string{aclStatusActive},
		Refresh: statusACL(ctx, conn, aclId),
		Timeout: aclActiveTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

// waitACLDeleted waits for MemoryDB ACL to be deleted.
func waitACLDeleted(ctx context.Context, conn *memorydb.MemoryDB, aclId string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{aclStatusDeleting},
		Target:  []string{},
		Refresh: statusACL(ctx, conn, aclId),
		Timeout: aclDeletedTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

// waitClusterAvailable waits for MemoryDB Cluster to reach an active state after modifications.
func waitClusterAvailable(ctx context.Context, conn *memorydb.MemoryDB, clusterId string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{clusterStatusCreating, clusterStatusUpdating},
		Target:  []string{clusterStatusAvailable},
		Refresh: statusCluster(ctx, conn, clusterId),
		Timeout: clusterAvailableTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

// waitClusterDeleted waits for MemoryDB Cluster to be deleted.
func waitClusterDeleted(ctx context.Context, conn *memorydb.MemoryDB, clusterId string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{clusterStatusDeleting},
		Target:  []string{},
		Refresh: statusCluster(ctx, conn, clusterId),
		Timeout: clusterDeletedTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

// waitClusterParameterGroupInSync waits for MemoryDB Cluster to come in sync
// with a new parameter group.
func waitClusterParameterGroupInSync(ctx context.Context, conn *memorydb.MemoryDB, clusterId string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{clusterParameterGroupStatusApplying},
		Target:  []string{clusterParameterGroupStatusInSync},
		Refresh: statusClusterParameterGroup(ctx, conn, clusterId),
		Timeout: clusterParameterGroupInSyncTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

// waitClusterSecurityGroupsActive waits for MemoryDB Cluster to apply all
// security group-related changes.
func waitClusterSecurityGroupsActive(ctx context.Context, conn *memorydb.MemoryDB, clusterId string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{clusterSecurityGroupStatusModifying},
		Target:  []string{clusterSecurityGroupStatusActive},
		Refresh: statusClusterSecurityGroups(ctx, conn, clusterId),
		Timeout: clusterSecurityGroupsActiveTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

// waitUserActive waits for MemoryDB user to reach an active state after modifications.
func waitUserActive(ctx context.Context, conn *memorydb.MemoryDB, userId string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{userStatusModifying},
		Target:  []string{userStatusActive},
		Refresh: statusUser(ctx, conn, userId),
		Timeout: userActiveTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

// waitUserDeleted waits for MemoryDB user to be deleted.
func waitUserDeleted(ctx context.Context, conn *memorydb.MemoryDB, userId string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{userStatusDeleting},
		Target:  []string{},
		Refresh: statusUser(ctx, conn, userId),
		Timeout: userDeletedTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}
