package ec2

const (
	// https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_CreditSpecificationRequest.html#API_CreditSpecificationRequest_Contents
	CPUCreditsStandard  = "standard"
	CPUCreditsUnlimited = "unlimited"
)

func CPUCredits_Values() []string {
	return []string{
		CPUCreditsStandard,
		CPUCreditsUnlimited,
	}
}

const (
	// https://docs.aws.amazon.com/vpc/latest/privatelink/vpce-interface.html#vpce-interface-lifecycle
	VPCEndpointStateAvailable         = "available"
	VPCEndpointStateDeleted           = "deleted"
	VPCEndpointStateDeleting          = "deleting"
	VPCEndpointStateFailed            = "failed"
	VPCEndpointStatePending           = "pending"
	VPCEndpointStatePendingAcceptance = "pendingAcceptance"
	VPCEndpointStateRejected          = "rejected"
)

// See https://docs.aws.amazon.com/vm-import/latest/userguide/vmimport-image-import.html#check-import-task-status
const (
	EBSSnapshotImportStateActive     = "active"
	EBSSnapshotImportStateDeleting   = "deleting"
	EBSSnapshotImportStateDeleted    = "deleted"
	EBSSnapshotImportStateUpdating   = "updating"
	EBSSnapshotImportStateValidating = "validating"
	EBSSnapshotImportStateValidated  = "validated"
	EBSSnapshotImportStateConverting = "converting"
	EBSSnapshotImportStateCompleted  = "completed"
)
