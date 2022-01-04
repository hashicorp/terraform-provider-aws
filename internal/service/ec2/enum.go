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
	EBSSnapshotImportStateValidd     = "validated"
	EBSSnapshotImportStateConverting = "converting"
	EBSSnapshotImportStateCompleted  = "completed"
)

// See https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_CreateNetworkInterface.html#API_CreateNetworkInterface_Example_2_Response.
const (
	NetworkInterfaceStatusPending = "pending"
)

// See https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeInternetGateways.html#API_DescribeInternetGateways_Example_1_Response.
const (
	InternetGatewayAttachmentStateAvailable = "available"
)
