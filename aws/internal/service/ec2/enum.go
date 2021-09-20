package ec2

const (
	// https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_CreditSpecificationRequest.html#API_CreditSpecificationRequest_Contents
	CpuCreditsStandard  = "standard"
	CpuCreditsUnlimited = "unlimited"
)

func CpuCredits_Values() []string {
	return []string{
		CpuCreditsStandard,
		CpuCreditsUnlimited,
	}
}

const (
	// https://docs.aws.amazon.com/vpc/latest/privatelink/vpce-interface.html#vpce-interface-lifecycle
	VpcEndpointStateAvailable         = "available"
	VpcEndpointStateDeleted           = "deleted"
	VpcEndpointStateDeleting          = "deleting"
	VpcEndpointStateFailed            = "failed"
	VpcEndpointStatePending           = "pending"
	VpcEndpointStatePendingAcceptance = "pendingAcceptance"
	VpcEndpointStateRejected          = "rejected"
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
