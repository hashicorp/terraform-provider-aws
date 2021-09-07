package redshift

// https://docs.aws.amazon.com/redshift/latest/mgmt/working-with-clusters.html#rs-mgmt-cluster-status.
const (
	ClusterStatusAvailable              = "available"
	ClusterStatusAvailablePrepForResize = "available, prep-for-resize"
	ClusterStatusAvailableResizeCleanup = "available, resize-cleanup"
	ClusterStatusCancellingResize       = "cancelling-resize"
	ClusterStatusCreating               = "creating"
	ClusterStatusDeleting               = "deleting"
	ClusterStatusFinalSnapshot          = "final-snapshot"
	ClusterStatusHardwareFailure        = "hardware-failure"
	ClusterStatusIncompatibleHsm        = "incompatible-hsm"
	ClusterStatusIncompatibleNetwork    = "incompatible-network"
	ClusterStatusIncompatibleParameters = "incompatible-parameters"
	ClusterStatusIncompatibleRestore    = "incompatible-restore"
	ClusterStatusModifying              = "modifying"
	ClusterStatusPaused                 = "paused"
	ClusterStatusRebooting              = "rebooting"
	ClusterStatusRenaming               = "renaming"
	ClusterStatusResizing               = "resizing"
	ClusterStatusRotatingKeys           = "rotating-keys"
	ClusterStatusStorageFull            = "storage-full"
	ClusterStatusUpdatingHsm            = "updating-hsm"
)

const (
	ClusterTypeMultiNode  = "multi-node"
	ClusterTypeSingleNode = "single-node"
)

func ClusterType_Values() []string {
	return []string{
		ClusterTypeMultiNode,
		ClusterTypeSingleNode,
	}
}
