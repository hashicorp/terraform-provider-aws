package ec2

import (
	"github.com/aws/aws-sdk-go/service/ec2"
)

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
	// The AWS SDK constant ec2.FleetOnDemandAllocationStrategyLowestPrice is incorrect.
	FleetOnDemandAllocationStrategyLowestPrice = "lowestPrice"
)

func FleetOnDemandAllocationStrategy_Values() []string {
	return append(
		removeFirstOccurrenceFromStringSlice(ec2.FleetOnDemandAllocationStrategy_Values(), ec2.FleetOnDemandAllocationStrategyLowestPrice),
		FleetOnDemandAllocationStrategyLowestPrice,
	)
}

const (
	// The AWS SDK constant ec2.SpotAllocationStrategyLowestPrice is incorrect.
	SpotAllocationStrategyLowestPrice = "lowestPrice"
)

func SpotAllocationStrategy_Values() []string {
	return append(
		removeFirstOccurrenceFromStringSlice(ec2.SpotAllocationStrategy_Values(), ec2.SpotAllocationStrategyLowestPrice),
		SpotAllocationStrategyLowestPrice,
	)
}

const (
	// https://docs.aws.amazon.com/vpc/latest/privatelink/vpce-interface.html#vpce-interface-lifecycle
	vpcEndpointStateAvailable         = "available"
	vpcEndpointStateDeleted           = "deleted"
	vpcEndpointStateDeleting          = "deleting"
	vpcEndpointStateFailed            = "failed"
	vpcEndpointStatePending           = "pending"
	vpcEndpointStatePendingAcceptance = "pendingAcceptance"
)

const (
	vpnStateModifying = "modifying"
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
	vpnTunnelOptionsDPDTimeoutActionClear   = "clear"
	vpnTunnelOptionsDPDTimeoutActionNone    = "none"
	vpnTunnelOptionsDPDTimeoutActionRestart = "restart"
)

func vpnTunnelOptionsDPDTimeoutAction_Values() []string {
	return []string{
		vpnTunnelOptionsDPDTimeoutActionClear,
		vpnTunnelOptionsDPDTimeoutActionNone,
		vpnTunnelOptionsDPDTimeoutActionRestart,
	}
}

const (
	vpnTunnelOptionsIKEVersion1 = "ikev1"
	vpnTunnelOptionsIKEVersion2 = "ikev2"
)

func vpnTunnelOptionsIKEVersion_Values() []string {
	return []string{
		vpnTunnelOptionsIKEVersion1,
		vpnTunnelOptionsIKEVersion2,
	}
}

const (
	vpnTunnelOptionsPhase1EncryptionAlgorithmAES128        = "AES128"
	vpnTunnelOptionsPhase1EncryptionAlgorithmAES256        = "AES256"
	vpnTunnelOptionsPhase1EncryptionAlgorithmAES128_GCM_16 = "AES128-GCM-16"
	vpnTunnelOptionsPhase1EncryptionAlgorithmAES256_GCM_16 = "AES256-GCM-16"
)

func vpnTunnelOptionsPhase1EncryptionAlgorithm_Values() []string {
	return []string{
		vpnTunnelOptionsPhase1EncryptionAlgorithmAES128,
		vpnTunnelOptionsPhase1EncryptionAlgorithmAES256,
		vpnTunnelOptionsPhase1EncryptionAlgorithmAES128_GCM_16,
		vpnTunnelOptionsPhase1EncryptionAlgorithmAES256_GCM_16,
	}
}

const (
	vpnTunnelOptionsPhase1IntegrityAlgorithmSHA1     = "SHA1"
	vpnTunnelOptionsPhase1IntegrityAlgorithmSHA2_256 = "SHA2-256"
	vpnTunnelOptionsPhase1IntegrityAlgorithmSHA2_384 = "SHA2-384"
	vpnTunnelOptionsPhase1IntegrityAlgorithmSHA2_512 = "SHA2-512"
)

func vpnTunnelOptionsPhase1IntegrityAlgorithm_Values() []string {
	return []string{
		vpnTunnelOptionsPhase1IntegrityAlgorithmSHA1,
		vpnTunnelOptionsPhase1IntegrityAlgorithmSHA2_256,
		vpnTunnelOptionsPhase1IntegrityAlgorithmSHA2_384,
		vpnTunnelOptionsPhase1IntegrityAlgorithmSHA2_512,
	}
}

const (
	vpnTunnelOptionsPhase2EncryptionAlgorithmAES128        = "AES128"
	vpnTunnelOptionsPhase2EncryptionAlgorithmAES256        = "AES256"
	vpnTunnelOptionsPhase2EncryptionAlgorithmAES128_GCM_16 = "AES128-GCM-16"
	vpnTunnelOptionsPhase2EncryptionAlgorithmAES256_GCM_16 = "AES256-GCM-16"
)

func vpnTunnelOptionsPhase2EncryptionAlgorithm_Values() []string {
	return []string{
		vpnTunnelOptionsPhase2EncryptionAlgorithmAES128,
		vpnTunnelOptionsPhase2EncryptionAlgorithmAES256,
		vpnTunnelOptionsPhase2EncryptionAlgorithmAES128_GCM_16,
		vpnTunnelOptionsPhase2EncryptionAlgorithmAES256_GCM_16,
	}
}

const (
	vpnTunnelOptionsPhase2IntegrityAlgorithmSHA1     = "SHA1"
	vpnTunnelOptionsPhase2IntegrityAlgorithmSHA2_256 = "SHA2-256"
	vpnTunnelOptionsPhase2IntegrityAlgorithmSHA2_384 = "SHA2-384"
	vpnTunnelOptionsPhase2IntegrityAlgorithmSHA2_512 = "SHA2-512"
)

func vpnTunnelOptionsPhase2IntegrityAlgorithm_Values() []string {
	return []string{
		vpnTunnelOptionsPhase2IntegrityAlgorithmSHA1,
		vpnTunnelOptionsPhase2IntegrityAlgorithmSHA2_256,
		vpnTunnelOptionsPhase2IntegrityAlgorithmSHA2_384,
		vpnTunnelOptionsPhase2IntegrityAlgorithmSHA2_512,
	}
}

const (
	vpnTunnelOptionsStartupActionAdd   = "add"
	vpnTunnelOptionsStartupActionStart = "start"
)

func vpnTunnelOptionsStartupAction_Values() []string {
	return []string{
		vpnTunnelOptionsStartupActionAdd,
		vpnTunnelOptionsStartupActionStart,
	}
}

const (
	vpnConnectionTypeIPsec1        = "ipsec.1"
	vpnConnectionTypeIPsec1_AES256 = "ipsec.1-aes256" // https://github.com/hashicorp/terraform-provider-aws/issues/23105.
)

func vpnConnectionType_Values() []string {
	return []string{
		vpnConnectionTypeIPsec1,
		vpnConnectionTypeIPsec1_AES256,
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

func removeFirstOccurrenceFromStringSlice(slice []string, s string) []string {
	for i, v := range slice {
		if v == s {
			return append(slice[:i], slice[i+1:]...)
		}
	}

	return slice
}
