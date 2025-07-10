// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshift

import (
	"time"
)

const (
	clusterAvailabilityStatusAvailable   = "Available"
	clusterAvailabilityStatusFailed      = "Failed"
	clusterAvailabilityStatusMaintenance = "Maintenance"
	clusterAvailabilityStatusModifying   = "Modifying"
	clusterAvailabilityStatusUnavailable = "Unavailable"
)

// https://docs.aws.amazon.com/redshift/latest/mgmt/working-with-clusters.html#rs-mgmt-cluster-status.

const (
	clusterStatusAvailable = "available"
	clusterStatusModifying = "modifying"
	clusterStatusRebooting = "rebooting"
)

const (
	clusterSnapshotStatusAvailable = "available"
	clusterSnapshotStatusCreating  = "creating"
	clusterSnapshotStatusDeleted   = "deleted"
)

const (
	clusterRestoreStatusStarting  = "starting"
	clusterRestoreStatusRestoring = "restoring"
	clusterRestoreStatusCompleted = "completed"
	clusterRestoreStatusFailed    = "failed"
)

const (
	clusterTypeMultiNode  = "multi-node"
	clusterTypeSingleNode = "single-node"
)

const (
	clusterAvailabilityZoneRelocationStatusEnabled          = "enabled"
	clusterAvailabilityZoneRelocationStatusDisabled         = "disabled"
	clusterAvailabilityZoneRelocationStatusPendingEnabling  = "pending_enabling"
	clusterAvailabilityZoneRelocationStatusPendingDisabling = "pending_disabling"
)

func clusterAvailabilityZoneRelocationStatus_TerminalValues() []string {
	return []string{
		clusterAvailabilityZoneRelocationStatusEnabled,
		clusterAvailabilityZoneRelocationStatusDisabled,
	}
}
func clusterAvailabilityZoneRelocationStatus_PendingValues() []string {
	return []string{
		clusterAvailabilityZoneRelocationStatusPendingEnabling,
		clusterAvailabilityZoneRelocationStatusPendingDisabling,
	}
}

const (
	endpointAccessStatusActive    = "active"
	endpointAccessStatusCreating  = "creating"
	endpointAccessStatusDeleting  = "deleting"
	endpointAccessStatusModifying = "modifying"
)

const (
	propagationTimeout = 2 * time.Minute
)
