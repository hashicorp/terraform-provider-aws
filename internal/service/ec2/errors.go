// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	multierror "github.com/hashicorp/go-multierror"
)

const (
	errCodeAuthFailure                                       = "AuthFailure"
	errCodeClientInvalidHostIDNotFound                       = "Client.InvalidHostID.NotFound"
	errCodeConcurrentMutationLimitExceeded                   = "ConcurrentMutationLimitExceeded"
	ErrCodeDefaultSubnetAlreadyExistsInAvailabilityZone      = "DefaultSubnetAlreadyExistsInAvailabilityZone"
	errCodeDependencyViolation                               = "DependencyViolation"
	errCodeGatewayNotAttached                                = "Gateway.NotAttached"
	errCodeIncorrectState                                    = "IncorrectState"
	errCodeInsufficientInstanceCapacity                      = "InsufficientInstanceCapacity"
	errCodeInvalidAMIIDNotFound                              = "InvalidAMIID.NotFound"
	errCodeInvalidAMIIDUnavailable                           = "InvalidAMIID.Unavailable"
	errCodeInvalidAddressNotFound                            = "InvalidAddress.NotFound"
	errCodeInvalidAllocationIDNotFound                       = "InvalidAllocationID.NotFound"
	errCodeInvalidAssociationIDNotFound                      = "InvalidAssociationID.NotFound"
	errCodeInvalidAttachmentIDNotFound                       = "InvalidAttachmentID.NotFound"
	errCodeInvalidCapacityReservationIdNotFound              = "InvalidCapacityReservationId.NotFound'"
	errCodeInvalidCarrierGatewayIDNotFound                   = "InvalidCarrierGatewayID.NotFound"
	errCodeInvalidClientVPNActiveAssociationNotFound         = "InvalidClientVpnActiveAssociationNotFound"
	errCodeInvalidClientVPNAssociationIdNotFound             = "InvalidClientVpnAssociationIdNotFound"
	errCodeInvalidClientVPNAuthorizationRuleNotFound         = "InvalidClientVpnEndpointAuthorizationRuleNotFound"
	errCodeInvalidClientVPNEndpointIdNotFound                = "InvalidClientVpnEndpointId.NotFound"
	errCodeInvalidClientVPNRouteNotFound                     = "InvalidClientVpnRouteNotFound"
	errCodeInvalidConnectionNotification                     = "InvalidConnectionNotification"
	errCodeInvalidConversionTaskIdMalformed                  = "InvalidConversionTaskId.Malformed"
	errCodeInvalidCustomerGatewayIDNotFound                  = "InvalidCustomerGatewayID.NotFound"
	errCodeInvalidDHCPOptionIDNotFound                       = "InvalidDhcpOptionID.NotFound"
	errCodeInvalidFleetIdNotFound                            = "InvalidFleetId.NotFound"
	errCodeInvalidFlowLogIdNotFound                          = "InvalidFlowLogId.NotFound"
	errCodeInvalidGatewayIDNotFound                          = "InvalidGatewayID.NotFound"
	errCodeInvalidGroupInUse                                 = "InvalidGroup.InUse"
	errCodeInvalidGroupNotFound                              = "InvalidGroup.NotFound"
	errCodeInvalidHostIDNotFound                             = "InvalidHostID.NotFound"
	errCodeInvalidInstanceConnectEndpointIdNotFound          = "InvalidInstanceConnectEndpointId.NotFound"
	errCodeInvalidInstanceID                                 = "InvalidInstanceID"
	errCodeInvalidInstanceIDNotFound                         = "InvalidInstanceID.NotFound"
	errCodeInvalidInternetGatewayIDNotFound                  = "InvalidInternetGatewayID.NotFound"
	errCodeInvalidIPAMIdNotFound                             = "InvalidIpamId.NotFound"
	errCodeInvalidIPAMPoolAllocationIdNotFound               = "InvalidIpamPoolAllocationId.NotFound"
	errCodeInvalidIPAMPoolIdNotFound                         = "InvalidIpamPoolId.NotFound"
	errCodeInvalidIPAMResourceDiscoveryIdNotFound            = "InvalidIpamResourceDiscoveryId.NotFound"
	errCodeInvalidIPAMResourceDiscoveryAssociationIdNotFound = "InvalidIpamResourceDiscoveryAssociationId.NotFound"
	errCodeInvalidIPAMScopeIdNotFound                        = "InvalidIpamScopeId.NotFound"
	errCodeInvalidKeyPairNotFound                            = "InvalidKeyPair.NotFound"
	errCodeInvalidLaunchTemplateIdMalformed                  = "InvalidLaunchTemplateId.Malformed"
	errCodeInvalidLaunchTemplateIdNotFound                   = "InvalidLaunchTemplateId.NotFound"
	errCodeInvalidLaunchTemplateIdVersionNotFound            = "InvalidLaunchTemplateId.VersionNotFound"
	errCodeInvalidLaunchTemplateNameNotFoundException        = "InvalidLaunchTemplateName.NotFoundException"
	errCodeInvalidNetworkACLEntryNotFound                    = "InvalidNetworkAclEntry.NotFound"
	errCodeInvalidNetworkACLIDNotFound                       = "InvalidNetworkAclID.NotFound"
	errCodeInvalidNetworkInterfaceIDNotFound                 = "InvalidNetworkInterfaceID.NotFound"
	errCodeInvalidNetworkInsightsAnalysisIdNotFound          = "InvalidNetworkInsightsAnalysisId.NotFound"
	errCodeInvalidNetworkInsightsPathIdNotFound              = "InvalidNetworkInsightsPathId.NotFound"
	errCodeInvalidParameter                                  = "InvalidParameter"
	errCodeInvalidParameterCombination                       = "InvalidParameterCombination"
	errCodeInvalidParameterException                         = "InvalidParameterException"
	errCodeInvalidParameterValue                             = "InvalidParameterValue"
	errCodeInvalidPermissionDuplicate                        = "InvalidPermission.Duplicate"
	errCodeInvalidPermissionNotFound                         = "InvalidPermission.NotFound"
	errCodeInvalidPlacementGroupUnknown                      = "InvalidPlacementGroup.Unknown"
	errCodeInvalidPoolIDNotFound                             = "InvalidPoolID.NotFound"
	errCodeInvalidPrefixListIDNotFound                       = "InvalidPrefixListID.NotFound"
	errCodeInvalidPrefixListIdNotFound                       = "InvalidPrefixListId.NotFound"
	errCodeInvalidPublicIpv4PoolIDNotFound                   = "InvalidPublicIpv4PoolID.NotFound" // nosemgrep:ci.caps5-in-const-name,ci.caps5-in-var-name
	errCodeInvalidRouteNotFound                              = "InvalidRoute.NotFound"
	errCodeInvalidRouteTableIDNotFound                       = "InvalidRouteTableID.NotFound"
	errCodeInvalidRouteTableIdNotFound                       = "InvalidRouteTableId.NotFound"
	errCodeInvalidSecurityGroupIDNotFound                    = "InvalidSecurityGroupID.NotFound"
	errCodeInvalidSecurityGroupRuleIdNotFound                = "InvalidSecurityGroupRuleId.NotFound"
	errCodeInvalidServiceName                                = "InvalidServiceName"
	errCodeInvalidSnapshotInUse                              = "InvalidSnapshot.InUse"
	errCodeInvalidSnapshotNotFound                           = "InvalidSnapshot.NotFound"
	ErrCodeInvalidSpotDatafeedNotFound                       = "InvalidSpotDatafeed.NotFound"
	errCodeInvalidSpotFleetRequestConfig                     = "InvalidSpotFleetRequestConfig"
	errCodeInvalidSpotFleetRequestIdNotFound                 = "InvalidSpotFleetRequestId.NotFound"
	errCodeInvalidSpotInstanceRequestIDNotFound              = "InvalidSpotInstanceRequestID.NotFound"
	errCodeInvalidSubnetCIDRReservationIDNotFound            = "InvalidSubnetCidrReservationID.NotFound"
	errCodeInvalidSubnetIDNotFound                           = "InvalidSubnetID.NotFound"
	errCodeInvalidSubnetIdNotFound                           = "InvalidSubnetId.NotFound"
	errCodeInvalidTrafficMirrorFilterIdNotFound              = "InvalidTrafficMirrorFilterId.NotFound"
	errCodeInvalidTrafficMirrorFilterRuleIdNotFound          = "InvalidTrafficMirrorFilterRuleId.NotFound"
	errCodeInvalidTrafficMirrorSessionIdNotFound             = "InvalidTrafficMirrorSessionId.NotFound"
	errCodeInvalidTrafficMirrorTargetIdNotFound              = "InvalidTrafficMirrorTargetId.NotFound"
	errCodeInvalidTransitGatewayAttachmentIDNotFound         = "InvalidTransitGatewayAttachmentID.NotFound"
	errCodeInvalidTransitGatewayConnectPeerIDNotFound        = "InvalidTransitGatewayConnectPeerID.NotFound"
	errCodeInvalidTransitGatewayPolicyTableIdNotFound        = "InvalidTransitGatewayPolicyTableId.NotFound"
	errCodeInvalidTransitGatewayIDNotFound                   = "InvalidTransitGatewayID.NotFound"
	errCodeInvalidTransitGatewayMulticastDomainIdNotFound    = "InvalidTransitGatewayMulticastDomainId.NotFound"
	errCodeInvalidVolumeNotFound                             = "InvalidVolume.NotFound"
	errCodeInvalidVPCCIDRBlockAssociationIDNotFound          = "InvalidVpcCidrBlockAssociationID.NotFound"
	errCodeInvalidVPCEndpointIdNotFound                      = "InvalidVpcEndpointId.NotFound"
	errCodeInvalidVPCEndpointNotFound                        = "InvalidVpcEndpoint.NotFound"
	errCodeInvalidVPCEndpointServiceIdNotFound               = "InvalidVpcEndpointServiceId.NotFound"
	errCodeInvalidVPCIDNotFound                              = "InvalidVpcID.NotFound"
	errCodeInvalidVPCPeeringConnectionIDNotFound             = "InvalidVpcPeeringConnectionID.NotFound"
	errCodeInvalidVPNConnectionIDNotFound                    = "InvalidVpnConnectionID.NotFound"
	errCodeInvalidVPNGatewayAttachmentNotFound               = "InvalidVpnGatewayAttachment.NotFound"
	errCodeInvalidVPNGatewayIDNotFound                       = "InvalidVpnGatewayID.NotFound"
	errCodeNatGatewayNotFound                                = "NatGatewayNotFound"
	errCodeOperationNotPermitted                             = "OperationNotPermitted"
	errCodePrefixListVersionMismatch                         = "PrefixListVersionMismatch"
	errCodeResourceNotReady                                  = "ResourceNotReady"
	errCodeSnapshotCreationPerVolumeRateExceeded             = "SnapshotCreationPerVolumeRateExceeded"
	errCodeUnsupportedOperation                              = "UnsupportedOperation"
	errCodeVolumeInUse                                       = "VolumeInUse"
	errCodeVPNConnectionLimitExceeded                        = "VpnConnectionLimitExceeded"
	errCodeVPNGatewayLimitExceeded                           = "VpnGatewayLimitExceeded"
)

func CancelSpotFleetRequestError(apiObject *ec2.CancelSpotFleetRequestsErrorItem) error {
	if apiObject == nil || apiObject.Error == nil {
		return nil
	}

	return awserr.New(aws.StringValue(apiObject.Error.Code), aws.StringValue(apiObject.Error.Message), nil)
}

func CancelSpotFleetRequestsError(apiObjects []*ec2.CancelSpotFleetRequestsErrorItem) error {
	var errors *multierror.Error

	for _, apiObject := range apiObjects {
		if err := CancelSpotFleetRequestError(apiObject); err != nil {
			errors = multierror.Append(errors, fmt.Errorf("%s: %w", aws.StringValue(apiObject.SpotFleetRequestId), err))
		}
	}

	return errors.ErrorOrNil()
}

func DeleteFleetError(apiObject *ec2.DeleteFleetErrorItem) error {
	if apiObject == nil || apiObject.Error == nil {
		return nil
	}

	return awserr.New(aws.StringValue(apiObject.Error.Code), aws.StringValue(apiObject.Error.Message), nil)
}

func DeleteFleetsError(apiObjects []*ec2.DeleteFleetErrorItem) error {
	var errors *multierror.Error

	for _, apiObject := range apiObjects {
		if err := DeleteFleetError(apiObject); err != nil {
			errors = multierror.Append(errors, fmt.Errorf("%s: %w", aws.StringValue(apiObject.FleetId), err))
		}
	}

	return errors.ErrorOrNil()
}

func UnsuccessfulItemError(apiObject *ec2.UnsuccessfulItemError) error {
	if apiObject == nil {
		return nil
	}

	return awserr.New(aws.StringValue(apiObject.Code), aws.StringValue(apiObject.Message), nil)
}

func UnsuccessfulItemsError(apiObjects []*ec2.UnsuccessfulItem) error {
	var errors *multierror.Error

	for _, apiObject := range apiObjects {
		if apiObject == nil {
			continue
		}

		err := UnsuccessfulItemError(apiObject.Error)

		if err != nil {
			errors = multierror.Append(errors, fmt.Errorf("%s: %w", aws.StringValue(apiObject.ResourceId), err))
		}
	}

	return errors.ErrorOrNil()
}
