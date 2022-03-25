package redshift

// https://docs.aws.amazon.com/redshift/latest/mgmt/working-with-clusters.html#rs-mgmt-cluster-status.
//nolint:deadcode,varcheck // These constants are missing from the AWS SDK
const (
	clusterStatusAvailable              = "available"
	clusterStatusAvailablePrepForResize = "available, prep-for-resize"
	clusterStatusAvailableResizeCleanup = "available, resize-cleanup"
	clusterStatusCancellingResize       = "cancelling-resize"
	clusterStatusCreating               = "creating"
	clusterStatusDeleting               = "deleting"
	clusterStatusFinalSnapshot          = "final-snapshot"
	clusterStatusHardwareFailure        = "hardware-failure"
	clusterStatusIncompatibleHSM        = "incompatible-hsm"
	clusterStatusIncompatibleNetwork    = "incompatible-network"
	clusterStatusIncompatibleParameters = "incompatible-parameters"
	clusterStatusIncompatibleRestore    = "incompatible-restore"
	clusterStatusModifying              = "modifying"
	clusterStatusPaused                 = "paused"
	clusterStatusRebooting              = "rebooting"
	clusterStatusRenaming               = "renaming"
	clusterStatusResizing               = "resizing"
	clusterStatusRotatingKeys           = "rotating-keys"
	clusterStatusStorageFull            = "storage-full"
	clusterStatusUpdatingHSM            = "updating-hsm"
)

const (
	clusterTypeMultiNode  = "multi-node"
	clusterTypeSingleNode = "single-node"
)

//nolint:deadcode // These constants are missing from the AWS SDK
func clusterType_Values() []string {
	return []string{
		clusterTypeMultiNode,
		clusterTypeSingleNode,
	}
}

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
