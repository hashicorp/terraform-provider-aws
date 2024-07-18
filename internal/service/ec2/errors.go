// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"errors"
	"fmt"

	aws_sdkv2 "github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
)

const (
	errCodeAnalysisExistsForNetworkInsightsPath                    = "AnalysisExistsForNetworkInsightsPath"
	errCodeAuthFailure                                             = "AuthFailure"
	errCodeClientInvalidHostIDNotFound                             = "Client.InvalidHostID.NotFound"
	errCodeConcurrentMutationLimitExceeded                         = "ConcurrentMutationLimitExceeded"
	errCodeDefaultSubnetAlreadyExistsInAvailabilityZone            = "DefaultSubnetAlreadyExistsInAvailabilityZone"
	errCodeDependencyViolation                                     = "DependencyViolation"
	errCodeGatewayNotAttached                                      = "Gateway.NotAttached"
	errCodeIPAMOrganizationAccountNotRegistered                    = "IpamOrganizationAccountNotRegistered"
	errCodeIncorrectState                                          = "IncorrectState"
	errCodeInsufficientInstanceCapacity                            = "InsufficientInstanceCapacity"
	errCodeInvalidAMIIDNotFound                                    = "InvalidAMIID.NotFound"
	errCodeInvalidAMIIDUnavailable                                 = "InvalidAMIID.Unavailable"
	errCodeInvalidAddressNotFound                                  = "InvalidAddress.NotFound"
	errCodeInvalidAllocationIDNotFound                             = "InvalidAllocationID.NotFound"
	errCodeInvalidAssociationIDNotFound                            = "InvalidAssociationID.NotFound"
	errCodeInvalidAssociationNotFound                              = "InvalidAssociation.NotFound"
	errCodeInvalidAttachmentIDNotFound                             = "InvalidAttachmentID.NotFound"
	errCodeInvalidCapacityReservationIdNotFound                    = "InvalidCapacityReservationId.NotFound"
	errCodeInvalidCarrierGatewayIDNotFound                         = "InvalidCarrierGatewayID.NotFound"
	errCodeInvalidClientVPNActiveAssociationNotFound               = "InvalidClientVpnActiveAssociationNotFound"
	errCodeInvalidClientVPNAssociationIdNotFound                   = "InvalidClientVpnAssociationIdNotFound"
	errCodeInvalidClientVPNAuthorizationRuleNotFound               = "InvalidClientVpnEndpointAuthorizationRuleNotFound"
	errCodeInvalidClientVPNEndpointIdNotFound                      = "InvalidClientVpnEndpointId.NotFound"
	errCodeInvalidClientVPNRouteNotFound                           = "InvalidClientVpnRouteNotFound"
	errCodeInvalidConnectionNotification                           = "InvalidConnectionNotification"
	errCodeInvalidConversionTaskIdMalformed                        = "InvalidConversionTaskId.Malformed"
	errCodeInvalidCustomerGatewayIDNotFound                        = "InvalidCustomerGatewayID.NotFound"
	errCodeInvalidDHCPOptionIDNotFound                             = "InvalidDhcpOptionID.NotFound"
	errCodeInvalidDHCPOptionsIDNotFound                            = "InvalidDhcpOptionsID.NotFound"
	errCodeInvalidFleetIdNotFound                                  = "InvalidFleetId.NotFound"
	errCodeInvalidFlowLogIdNotFound                                = "InvalidFlowLogId.NotFound"
	errCodeInvalidGatewayIDNotFound                                = "InvalidGatewayID.NotFound"
	errCodeInvalidGroupInUse                                       = "InvalidGroup.InUse"
	errCodeInvalidGroupNotFound                                    = "InvalidGroup.NotFound"
	errCodeInvalidHostIDNotFound                                   = "InvalidHostID.NotFound"
	errCodeInvalidIPAMIdNotFound                                   = "InvalidIpamId.NotFound"
	errCodeInvalidIPAMPoolAllocationIdNotFound                     = "InvalidIpamPoolAllocationId.NotFound"
	errCodeInvalidIPAMPoolIdNotFound                               = "InvalidIpamPoolId.NotFound"
	errCodeInvalidIPAMResourceDiscoveryAssociationIdNotFound       = "InvalidIpamResourceDiscoveryAssociationId.NotFound"
	errCodeInvalidIPAMResourceDiscoveryIdNotFound                  = "InvalidIpamResourceDiscoveryId.NotFound"
	errCodeInvalidIPAMScopeIdNotFound                              = "InvalidIpamScopeId.NotFound"
	errCodeInvalidInstanceConnectEndpointIdNotFound                = "InvalidInstanceConnectEndpointId.NotFound"
	errCodeInvalidInstanceID                                       = "InvalidInstanceID"
	errCodeInvalidInstanceIDNotFound                               = "InvalidInstanceID.NotFound"
	errCodeInvalidInternetGatewayIDNotFound                        = "InvalidInternetGatewayID.NotFound"
	errCodeInvalidKeyPairNotFound                                  = "InvalidKeyPair.NotFound"
	errCodeInvalidLaunchTemplateIdMalformed                        = "InvalidLaunchTemplateId.Malformed"
	errCodeInvalidLaunchTemplateIdNotFound                         = "InvalidLaunchTemplateId.NotFound"
	errCodeInvalidLaunchTemplateIdVersionNotFound                  = "InvalidLaunchTemplateId.VersionNotFound"
	errCodeInvalidLaunchTemplateNameNotFoundException              = "InvalidLaunchTemplateName.NotFoundException"
	errCodeInvalidLocalGatewayRouteTableIDNotFound                 = "InvalidLocalGatewayRouteTableID.NotFound"
	errCodeInvalidLocalGatewayRouteTableVPCAssociationIDNotFound   = "InvalidLocalGatewayRouteTableVpcAssociationID.NotFound"
	errCodeInvalidNetworkACLEntryNotFound                          = "InvalidNetworkAclEntry.NotFound"
	errCodeInvalidNetworkACLIDNotFound                             = "InvalidNetworkAclID.NotFound"
	errCodeInvalidNetworkInsightsAnalysisIdNotFound                = "InvalidNetworkInsightsAnalysisId.NotFound"
	errCodeInvalidNetworkInsightsPathIdNotFound                    = "InvalidNetworkInsightsPathId.NotFound"
	errCodeInvalidNetworkInterfaceIDNotFound                       = "InvalidNetworkInterfaceID.NotFound"
	errCodeInvalidParameter                                        = "InvalidParameter"
	errCodeInvalidParameterCombination                             = "InvalidParameterCombination"
	errCodeInvalidParameterException                               = "InvalidParameterException"
	errCodeInvalidParameterValue                                   = "InvalidParameterValue"
	errCodeInvalidPermissionDuplicate                              = "InvalidPermission.Duplicate"
	errCodeInvalidPermissionNotFound                               = "InvalidPermission.NotFound"
	errCodeInvalidPlacementGroupUnknown                            = "InvalidPlacementGroup.Unknown"
	errCodeInvalidPoolIDNotFound                                   = "InvalidPoolID.NotFound"
	errCodeInvalidPrefixListIDNotFound                             = "InvalidPrefixListID.NotFound"
	errCodeInvalidPrefixListIdNotFound                             = "InvalidPrefixListId.NotFound"
	errCodeInvalidPublicIpv4PoolIDNotFound                         = "InvalidPublicIpv4PoolID.NotFound" // nosemgrep:ci.caps5-in-const-name,ci.caps5-in-var-name
	errCodeInvalidReservationNotFound                              = "InvalidReservationID.NotFound"
	errCodeInvalidRouteNotFound                                    = "InvalidRoute.NotFound"
	errCodeInvalidRouteTableIDNotFound                             = "InvalidRouteTableID.NotFound"
	errCodeInvalidRouteTableIdNotFound                             = "InvalidRouteTableId.NotFound"
	errCodeInvalidSecurityGroupIDNotFound                          = "InvalidSecurityGroupID.NotFound"
	errCodeInvalidSecurityGroupRuleIdNotFound                      = "InvalidSecurityGroupRuleId.NotFound"
	errCodeInvalidServiceName                                      = "InvalidServiceName"
	errCodeInvalidSnapshotInUse                                    = "InvalidSnapshot.InUse"
	errCodeInvalidSnapshotNotFound                                 = "InvalidSnapshot.NotFound"
	errCodeInvalidSpotDatafeedNotFound                             = "InvalidSpotDatafeed.NotFound"
	errCodeInvalidSpotFleetRequestConfig                           = "InvalidSpotFleetRequestConfig"
	errCodeInvalidSpotFleetRequestIdNotFound                       = "InvalidSpotFleetRequestId.NotFound"
	errCodeInvalidSpotInstanceRequestIDNotFound                    = "InvalidSpotInstanceRequestID.NotFound"
	errCodeInvalidSubnetCIDRReservationIDNotFound                  = "InvalidSubnetCidrReservationID.NotFound"
	errCodeInvalidSubnetIDNotFound                                 = "InvalidSubnetID.NotFound"
	errCodeInvalidSubnetIdNotFound                                 = "InvalidSubnetId.NotFound"
	errCodeInvalidTrafficMirrorFilterIdNotFound                    = "InvalidTrafficMirrorFilterId.NotFound"
	errCodeInvalidTrafficMirrorFilterRuleIdNotFound                = "InvalidTrafficMirrorFilterRuleId.NotFound"
	errCodeInvalidTrafficMirrorSessionIdNotFound                   = "InvalidTrafficMirrorSessionId.NotFound"
	errCodeInvalidTrafficMirrorTargetIdNotFound                    = "InvalidTrafficMirrorTargetId.NotFound"
	errCodeInvalidTransitGatewayAttachmentIDNotFound               = "InvalidTransitGatewayAttachmentID.NotFound"
	errCodeInvalidTransitGatewayConnectPeerIDNotFound              = "InvalidTransitGatewayConnectPeerID.NotFound"
	errCodeInvalidTransitGatewayIDNotFound                         = "InvalidTransitGatewayID.NotFound"
	errCodeInvalidTransitGatewayMulticastDomainAssociationNotFound = "InvalidTransitGatewayMulticastDomainAssociation.NotFound"
	errCodeInvalidTransitGatewayMulticastDomainIdNotFound          = "InvalidTransitGatewayMulticastDomainId.NotFound"
	errCodeInvalidTransitGatewayPolicyTableAssociationNotFound     = "InvalidTransitGatewayPolicyTableAssociation.NotFound"
	errCodeInvalidTransitGatewayPolicyTableIdNotFound              = "InvalidTransitGatewayPolicyTableId.NotFound"
	errCodeInvalidVPCCIDRBlockAssociationIDNotFound                = "InvalidVpcCidrBlockAssociationID.NotFound"
	errCodeInvalidVPCEndpointIdNotFound                            = "InvalidVpcEndpointId.NotFound"
	errCodeInvalidVPCEndpointNotFound                              = "InvalidVpcEndpoint.NotFound"
	errCodeInvalidVPCEndpointServiceIdNotFound                     = "InvalidVpcEndpointServiceId.NotFound"
	errCodeInvalidVPCEndpointServiceNotFound                       = "InvalidVpcEndpointService.NotFound"
	errCodeInvalidVPCIDNotFound                                    = "InvalidVpcID.NotFound"
	errCodeInvalidVPCPeeringConnectionIDNotFound                   = "InvalidVpcPeeringConnectionID.NotFound"
	errCodeInvalidVPNConnectionIDNotFound                          = "InvalidVpnConnectionID.NotFound"
	errCodeInvalidVPNGatewayAttachmentNotFound                     = "InvalidVpnGatewayAttachment.NotFound"
	errCodeInvalidVPNGatewayIDNotFound                             = "InvalidVpnGatewayID.NotFound"
	errCodeInvalidVerifiedAccessEndpointIdNotFound                 = "InvalidVerifiedAccessEndpointId.NotFound"
	errCodeInvalidVerifiedAccessGroupIdNotFound                    = "InvalidVerifiedAccessGroupId.NotFound"
	errCodeInvalidVerifiedAccessInstanceIdNotFound                 = "InvalidVerifiedAccessInstanceId.NotFound"
	errCodeInvalidVerifiedAccessTrustProviderIdNotFound            = "InvalidVerifiedAccessTrustProviderId.NotFound"
	errCodeInvalidVolumeNotFound                                   = "InvalidVolume.NotFound"
	errCodeNatGatewayNotFound                                      = "NatGatewayNotFound"
	errCodeNetworkACLEntryAlreadyExists                            = "NetworkAclEntryAlreadyExists"
	errCodeOperationNotPermitted                                   = "OperationNotPermitted"
	errCodePrefixListVersionMismatch                               = "PrefixListVersionMismatch"
	errCodeResourceNotReady                                        = "ResourceNotReady"
	errCodeRouteAlreadyExists                                      = "RouteAlreadyExists"
	errCodeSnapshotCreationPerVolumeRateExceeded                   = "SnapshotCreationPerVolumeRateExceeded"
	errCodeTransitGatewayMulticastGroupMemberNotFound              = "TransitGatewayMulticastGroupMember.NotFound"
	errCodeTransitGatewayMulticastGroupSourceNotFound              = "TransitGatewayMulticastGroupSource.NotFound"
	errCodeTransitGatewayRouteTablePropagationNotFound             = "TransitGatewayRouteTablePropagation.NotFound"
	errCodeUnsupportedOperation                                    = "UnsupportedOperation"
	errCodeVPNConnectionLimitExceeded                              = "VpnConnectionLimitExceeded"
	errCodeVPNGatewayLimitExceeded                                 = "VpnGatewayLimitExceeded"
	errCodeVolumeInUse                                             = "VolumeInUse"
)

func cancelSpotFleetRequestError(apiObject *awstypes.CancelSpotFleetRequestsError) error {
	if apiObject == nil {
		return nil
	}

	return errs.APIError(apiObject.Code, aws_sdkv2.ToString(apiObject.Message))
}

func cancelSpotFleetRequestsError(apiObjects []awstypes.CancelSpotFleetRequestsErrorItem) error {
	var errs []error

	for _, apiObject := range apiObjects {
		if err := cancelSpotFleetRequestError(apiObject.Error); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", aws_sdkv2.ToString(apiObject.SpotFleetRequestId), err))
		}
	}

	return errors.Join(errs...)
}

func deleteFleetError(apiObject *awstypes.DeleteFleetError) error {
	if apiObject == nil {
		return nil
	}

	return errs.APIError(apiObject.Code, aws_sdkv2.ToString(apiObject.Message))
}

func deleteFleetsError(apiObjects []awstypes.DeleteFleetErrorItem) error {
	var errs []error

	for _, apiObject := range apiObjects {
		if err := deleteFleetError(apiObject.Error); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", aws_sdkv2.ToString(apiObject.FleetId), err))
		}
	}

	return errors.Join(errs...)
}

func UnsuccessfulItemError(apiObject *ec2.UnsuccessfulItemError) error {
	if apiObject == nil {
		return nil
	}

	return awserr.New(aws.StringValue(apiObject.Code), aws.StringValue(apiObject.Message), nil)
}

func UnsuccessfulItemsError(apiObjects []*ec2.UnsuccessfulItem) error {
	var errs []error

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		if err := UnsuccessfulItemError(apiObject.Error); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", aws.StringValue(apiObject.ResourceId), err))
		}
	}

	return errors.Join(errs...)
}

func unsuccessfulItemErrorV2(apiObject *awstypes.UnsuccessfulItemError) error {
	if apiObject == nil {
		return nil
	}

	return errs.APIError(aws_sdkv2.ToString(apiObject.Code), aws_sdkv2.ToString(apiObject.Message))
}

func unsuccessfulItemsErrorV2(apiObjects []awstypes.UnsuccessfulItem) error {
	var errs []error

	for _, apiObject := range apiObjects {
		if err := unsuccessfulItemErrorV2(apiObject.Error); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", aws_sdkv2.ToString(apiObject.ResourceId), err))
		}
	}

	return errors.Join(errs...)
}

func enableFastSnapshotRestoreStateItemError(apiObject *awstypes.EnableFastSnapshotRestoreStateError) error {
	if apiObject == nil {
		return nil
	}

	return errs.APIError(aws_sdkv2.ToString(apiObject.Code), aws_sdkv2.ToString(apiObject.Message))
}

func enableFastSnapshotRestoreStateItemsError(apiObjects []awstypes.EnableFastSnapshotRestoreStateErrorItem) error {
	var errs []error

	for _, apiObject := range apiObjects {
		if err := enableFastSnapshotRestoreStateItemError(apiObject.Error); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", aws_sdkv2.ToString(apiObject.AvailabilityZone), err))
		}
	}

	return errors.Join(errs...)
}

func enableFastSnapshotRestoreItemsError(apiObjects []awstypes.EnableFastSnapshotRestoreErrorItem) error {
	var errs []error

	for _, apiObject := range apiObjects {
		if err := enableFastSnapshotRestoreStateItemsError(apiObject.FastSnapshotRestoreStateErrors); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", aws_sdkv2.ToString(apiObject.SnapshotId), err))
		}
	}

	return errors.Join(errs...)
}

func networkACLEntryAlreadyExistsError(naclID string, egress bool, ruleNumber int) error {
	return awserr.New(errCodeNetworkACLEntryAlreadyExists, fmt.Sprintf("EC2 Network ACL (%s) Rule (egress: %t)(%d) already exists", naclID, egress, ruleNumber), nil)
}

func routeAlreadyExistsError(routeTableID, destination string) error {
	return errs.APIError(errCodeRouteAlreadyExists, fmt.Sprintf("Route in Route Table (%s) with destination (%s) already exists", routeTableID, destination))
}
