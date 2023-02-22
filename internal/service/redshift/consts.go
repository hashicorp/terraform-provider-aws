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
