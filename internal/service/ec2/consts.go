// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	// https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_CreditSpecificationRequest.html#API_CreditSpecificationRequest_Contents
	cpuCreditsStandard  = "standard"
	cpuCreditsUnlimited = "unlimited"
)

func cpuCredits_Values() []string {
	return []string{
		cpuCreditsStandard,
		cpuCreditsUnlimited,
	}
}

const (
	// The AWS SDK constant ec2.fleetOnDemandAllocationStrategyLowestPrice is incorrect.
	fleetOnDemandAllocationStrategyLowestPrice = "lowestPrice"
)

func fleetOnDemandAllocationStrategy_Values() []string {
	return append(
		tfslices.RemoveAll(enum.Values[awstypes.FleetOnDemandAllocationStrategy](), string(awstypes.FleetOnDemandAllocationStrategyLowestPrice)),
		fleetOnDemandAllocationStrategyLowestPrice,
	)
}

const (
	// The AWS SDK constant ec2.spotAllocationStrategyLowestPrice is incorrect.
	spotAllocationStrategyLowestPrice = "lowestPrice"
)

func spotAllocationStrategy_Values() []string {
	return append(
		tfslices.RemoveAll(enum.Values[awstypes.SpotAllocationStrategy](), string(awstypes.SpotAllocationStrategyLowestPrice)),
		spotAllocationStrategyLowestPrice,
	)
}

const (
	// https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/spot-request-status.html#spot-instance-request-status-understand
	spotInstanceRequestStatusCodeFulfilled          = "fulfilled"
	spotInstanceRequestStatusCodePendingEvaluation  = "pending-evaluation"
	spotInstanceRequestStatusCodePendingFulfillment = "pending-fulfillment"
)

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
	ebsSnapshotImportStateActive     = "active"
	ebsSnapshotImportStateDeleting   = "deleting"
	ebsSnapshotImportStateDeleted    = "deleted"
	ebsSnapshotImportStateUpdating   = "updating"
	ebsSnapshotImportStateValidating = "validating"
	ebsSnapshotImportStateValidated  = "validated"
	ebsSnapshotImportStateConverting = "converting"
	ebsSnapshotImportStateCompleted  = "completed"
)

// See https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_CreateNetworkInterface.html#API_CreateNetworkInterface_Example_2_Response.
const (
	networkInterfaceStatusPending = "pending"
)

// See https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_DescribeInternetGateways.html#API_DescribeInternetGateways_Example_1_Response.
const (
	internetGatewayAttachmentStateAvailable = "available"
)

// See https://docs.aws.amazon.com/AWSEC2/latest/APIReference/API_CustomerGateway.html#API_CustomerGateway_Contents.
const (
	customerGatewayStateAvailable = "available"
	customerGatewayStateDeleted   = "deleted"
	customerGatewayStateDeleting  = "deleting"
	customerGatewayStatePending   = "pending"
)

// See https://docs.aws.amazon.com/cli/latest/reference/ec2/modify-address-attribute.html#examples.
const (
	ptrUpdateStatusPending = "PENDING"
)

const (
	managedPrefixListAddressFamilyIPv4 = "IPv4"
	managedPrefixListAddressFamilyIPv6 = "IPv6"
)

func managedPrefixListAddressFamily_Values() []string {
	return []string{
		managedPrefixListAddressFamilyIPv4,
		managedPrefixListAddressFamilyIPv6,
	}
}

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
	vpnTunnelCloudWatchLogOutputFormatJSON = names.AttrJSON
	vpnTunnelCloudWatchLogOutputFormatText = "text"
)

func vpnTunnelCloudWatchLogOutputFormat_Values() []string {
	return []string{
		names.AttrJSON,
		vpnTunnelCloudWatchLogOutputFormatText,
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
	amazonIPv6PoolID      = "Amazon"
	ipamManagedIPv6PoolID = "IPAM Managed"
)

const (
	defaultDHCPOptionsID = "default"
)

const (
	defaultSecurityGroupName = "default"
)

const (
	defaultSnapshotImportRoleName = "vmimport"
)

const (
	launchTemplateVersionDefault = "$Default"
	launchTemplateVersionLatest  = "$Latest"
)

const (
	sriovNetSupportSimple = "simple"
)

const (
	targetStorageTierStandard awstypes.TargetStorageTier = "standard"
)

const (
	outsideIPAddressTypePrivateIPv4 = "PrivateIpv4"
	outsideIPAddressTypePublicIPv4  = "PublicIpv4"
)

func outsideIPAddressType_Values() []string {
	return []string{
		outsideIPAddressTypePrivateIPv4,
		outsideIPAddressTypePublicIPv4,
	}
}

type securityGroupRuleType string

const (
	securityGroupRuleTypeEgress  securityGroupRuleType = "egress"
	securityGroupRuleTypeIngress securityGroupRuleType = "ingress"
)

func (securityGroupRuleType) Values() []securityGroupRuleType {
	return []securityGroupRuleType{
		securityGroupRuleTypeEgress,
		securityGroupRuleTypeIngress,
	}
}

const (
	gatewayIDLocal      = "local"
	gatewayIDVPCLattice = "VpcLattice"
)

const (
	verifiedAccessAttachmentTypeVPC = "vpc"
)

func verifiedAccessAttachmentType_Values() []string {
	return []string{
		verifiedAccessAttachmentTypeVPC,
	}
}

const (
	verifiedAccessEndpointTypeLoadBalancer     = "load-balancer"
	verifiedAccessEndpointTypeNetworkInterface = "network-interface"
)

func verifiedAccessEndpointType_Values() []string {
	return []string{
		verifiedAccessEndpointTypeLoadBalancer,
		verifiedAccessEndpointTypeNetworkInterface,
	}
}

const (
	verifiedAccessEndpointProtocolHTTP  = "http"
	verifiedAccessEndpointProtocolHTTPS = "https"
)

func verifiedAccessEndpointProtocol_Values() []string {
	return []string{
		verifiedAccessEndpointProtocolHTTP,
		verifiedAccessEndpointProtocolHTTPS,
	}
}
