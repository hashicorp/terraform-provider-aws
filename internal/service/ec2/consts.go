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

const (
	VpnStateModifying = "modifying"
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

// See https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_CreateNetworkInterface.html#API_CreateNetworkInterface_Example_2_Response.
const (
	NetworkInterfaceStatusPending = "pending"
)

// See https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeInternetGateways.html#API_DescribeInternetGateways_Example_1_Response.
const (
	InternetGatewayAttachmentStateAvailable = "available"
)

// See https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_CustomerGateway.html#API_CustomerGateway_Contents.
const (
	CustomerGatewayStateAvailable = "available"
	CustomerGatewayStateDeleted   = "deleted"
	CustomerGatewayStateDeleting  = "deleting"
	CustomerGatewayStatePending   = "pending"
)

const (
	VpnTunnelOptionsDPDTimeoutActionClear   = "clear"
	VpnTunnelOptionsDPDTimeoutActionNone    = "none"
	VpnTunnelOptionsDPDTimeoutActionRestart = "restart"
)

func VpnTunnelOptionsDPDTimeoutAction_Values() []string {
	return []string{
		VpnTunnelOptionsDPDTimeoutActionClear,
		VpnTunnelOptionsDPDTimeoutActionNone,
		VpnTunnelOptionsDPDTimeoutActionRestart,
	}
}

const (
	VpnTunnelOptionsIKEVersion1 = "ikev1"
	VpnTunnelOptionsIKEVersion2 = "ikev2"
)

func VpnTunnelOptionsIKEVersion_Values() []string {
	return []string{
		VpnTunnelOptionsIKEVersion1,
		VpnTunnelOptionsIKEVersion2,
	}
}

const (
	VpnTunnelOptionsPhase1EncryptionAlgorithmAES128        = "AES128"
	VpnTunnelOptionsPhase1EncryptionAlgorithmAES256        = "AES256"
	VpnTunnelOptionsPhase1EncryptionAlgorithmAES128_GCM_16 = "AES128-GCM-16"
	VpnTunnelOptionsPhase1EncryptionAlgorithmAES256_GCM_16 = "AES256-GCM-16"
)

func VpnTunnelOptionsPhase1EncryptionAlgorithm_Values() []string {
	return []string{
		VpnTunnelOptionsPhase1EncryptionAlgorithmAES128,
		VpnTunnelOptionsPhase1EncryptionAlgorithmAES256,
		VpnTunnelOptionsPhase1EncryptionAlgorithmAES128_GCM_16,
		VpnTunnelOptionsPhase1EncryptionAlgorithmAES256_GCM_16,
	}
}

const (
	VpnTunnelOptionsPhase1IntegrityAlgorithmSHA1     = "SHA1"
	VpnTunnelOptionsPhase1IntegrityAlgorithmSHA2_256 = "SHA2-256"
	VpnTunnelOptionsPhase1IntegrityAlgorithmSHA2_384 = "SHA2-384"
	VpnTunnelOptionsPhase1IntegrityAlgorithmSHA2_512 = "SHA2-512"
)

func VpnTunnelOptionsPhase1IntegrityAlgorithm_Values() []string {
	return []string{
		VpnTunnelOptionsPhase1IntegrityAlgorithmSHA1,
		VpnTunnelOptionsPhase1IntegrityAlgorithmSHA2_256,
		VpnTunnelOptionsPhase1IntegrityAlgorithmSHA2_384,
		VpnTunnelOptionsPhase1IntegrityAlgorithmSHA2_512,
	}
}

const (
	VpnTunnelOptionsPhase2EncryptionAlgorithmAES128        = "AES128"
	VpnTunnelOptionsPhase2EncryptionAlgorithmAES256        = "AES256"
	VpnTunnelOptionsPhase2EncryptionAlgorithmAES128_GCM_16 = "AES128-GCM-16"
	VpnTunnelOptionsPhase2EncryptionAlgorithmAES256_GCM_16 = "AES256-GCM-16"
)

func VpnTunnelOptionsPhase2EncryptionAlgorithm_Values() []string {
	return []string{
		VpnTunnelOptionsPhase2EncryptionAlgorithmAES128,
		VpnTunnelOptionsPhase2EncryptionAlgorithmAES256,
		VpnTunnelOptionsPhase2EncryptionAlgorithmAES128_GCM_16,
		VpnTunnelOptionsPhase2EncryptionAlgorithmAES256_GCM_16,
	}
}

const (
	VpnTunnelOptionsPhase2IntegrityAlgorithmSHA1     = "SHA1"
	VpnTunnelOptionsPhase2IntegrityAlgorithmSHA2_256 = "SHA2-256"
	VpnTunnelOptionsPhase2IntegrityAlgorithmSHA2_384 = "SHA2-384"
	VpnTunnelOptionsPhase2IntegrityAlgorithmSHA2_512 = "SHA2-512"
)

func VpnTunnelOptionsPhase2IntegrityAlgorithm_Values() []string {
	return []string{
		VpnTunnelOptionsPhase2IntegrityAlgorithmSHA1,
		VpnTunnelOptionsPhase2IntegrityAlgorithmSHA2_256,
		VpnTunnelOptionsPhase2IntegrityAlgorithmSHA2_384,
		VpnTunnelOptionsPhase2IntegrityAlgorithmSHA2_512,
	}
}

const (
	VpnTunnelOptionsStartupActionAdd   = "add"
	VpnTunnelOptionsStartupActionStart = "start"
)

func VpnTunnelOptionsStartupAction_Values() []string {
	return []string{
		VpnTunnelOptionsStartupActionAdd,
		VpnTunnelOptionsStartupActionStart,
	}
}

const (
	VpnConnectionTypeIpsec1        = "ipsec.1"
	VpnConnectionTypeIpsec1_AES256 = "ipsec.1-aes256" // https://github.com/hashicorp/terraform-provider-aws/issues/23105.
)

func VpnConnectionType_Values() []string {
	return []string{
		VpnConnectionTypeIpsec1,
		VpnConnectionTypeIpsec1_AES256,
	}
}

const (
	AmazonIPv6PoolID = "Amazon"
)

const (
	DefaultDHCPOptionsID = "default"
)

const (
	DefaultSecurityGroupName = "default"
)

const (
	LaunchTemplateVersionDefault = "$Default"
	LaunchTemplateVersionLatest  = "$Latest"
)
