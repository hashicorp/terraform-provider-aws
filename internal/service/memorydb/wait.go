// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package memorydb

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/memorydb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
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

	snapshotAvailableTimeout = 120 * time.Minute
	snapshotDeletedTimeout   = 120 * time.Minute
)

// waitACLActive waits for MemoryDB ACL to reach an active state after modifications.
func waitACLActive(ctx context.Context, conn *memorydb.MemoryDB, aclId string) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ACLStatusCreating, ACLStatusModifying},
		Target:  []string{ACLStatusActive},
		Refresh: statusACL(ctx, conn, aclId),
		Timeout: aclActiveTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

// waitACLDeleted waits for MemoryDB ACL to be deleted.
func waitACLDeleted(ctx context.Context, conn *memorydb.MemoryDB, aclId string) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ACLStatusDeleting},
		Target:  []string{},
		Refresh: statusACL(ctx, conn, aclId),
		Timeout: aclDeletedTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

// waitClusterAvailable waits for MemoryDB Cluster to reach an active state after modifications.
func waitClusterAvailable(ctx context.Context, conn *memorydb.MemoryDB, clusterId string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ClusterStatusCreating, ClusterStatusUpdating},
		Target:  []string{ClusterStatusAvailable},
		Refresh: statusCluster(ctx, conn, clusterId),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

// waitClusterDeleted waits for MemoryDB Cluster to be deleted.
func waitClusterDeleted(ctx context.Context, conn *memorydb.MemoryDB, clusterId string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ClusterStatusDeleting},
		Target:  []string{},
		Refresh: statusCluster(ctx, conn, clusterId),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

// waitClusterParameterGroupInSync waits for MemoryDB Cluster to come in sync
// with a new parameter group.
func waitClusterParameterGroupInSync(ctx context.Context, conn *memorydb.MemoryDB, clusterId string) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ClusterParameterGroupStatusApplying},
		Target:  []string{ClusterParameterGroupStatusInSync},
		Refresh: statusClusterParameterGroup(ctx, conn, clusterId),
		Timeout: clusterParameterGroupInSyncTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

// waitClusterSecurityGroupsActive waits for MemoryDB Cluster to apply all
// security group-related changes.
func waitClusterSecurityGroupsActive(ctx context.Context, conn *memorydb.MemoryDB, clusterId string) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ClusterSecurityGroupStatusModifying},
		Target:  []string{ClusterSecurityGroupStatusActive},
		Refresh: statusClusterSecurityGroups(ctx, conn, clusterId),
		Timeout: clusterSecurityGroupsActiveTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

// waitUserActive waits for MemoryDB user to reach an active state after modifications.
func waitUserActive(ctx context.Context, conn *memorydb.MemoryDB, userId string) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{UserStatusModifying},
		Target:  []string{UserStatusActive},
		Refresh: statusUser(ctx, conn, userId),
		Timeout: userActiveTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

// waitUserDeleted waits for MemoryDB user to be deleted.
func waitUserDeleted(ctx context.Context, conn *memorydb.MemoryDB, userId string) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{UserStatusDeleting},
		Target:  []string{},
		Refresh: statusUser(ctx, conn, userId),
		Timeout: userDeletedTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

// waitSnapshotAvailable waits for MemoryDB snapshot to reach the available state.
func waitSnapshotAvailable(ctx context.Context, conn *memorydb.MemoryDB, snapshotId string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{SnapshotStatusCreating},
		Target:  []string{SnapshotStatusAvailable},
		Refresh: statusSnapshot(ctx, conn, snapshotId),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

// waitSnapshotDeleted waits for MemoryDB snapshot to be deleted.
func waitSnapshotDeleted(ctx context.Context, conn *memorydb.MemoryDB, snapshotId string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{SnapshotStatusDeleting},
		Target:  []string{},
		Refresh: statusSnapshot(ctx, conn, snapshotId),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}
