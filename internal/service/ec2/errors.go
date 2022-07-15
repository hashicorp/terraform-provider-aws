package ec2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	multierror "github.com/hashicorp/go-multierror"
)

const (
	errCodeAuthFailure                                    = "AuthFailure"
	errCodeClientInvalidHostIDNotFound                    = "Client.InvalidHostID.NotFound"
	ErrCodeDefaultSubnetAlreadyExistsInAvailabilityZone   = "DefaultSubnetAlreadyExistsInAvailabilityZone"
	errCodeDependencyViolation                            = "DependencyViolation"
	errCodeGatewayNotAttached                             = "Gateway.NotAttached"
	errCodeIncorrectState                                 = "IncorrectState"
	errCodeInvalidAMIIDNotFound                           = "InvalidAMIID.NotFound"
	errCodeInvalidAMIIDUnavailable                        = "InvalidAMIID.Unavailable"
	errCodeInvalidAddressNotFound                         = "InvalidAddress.NotFound"
	errCodeInvalidAllocationIDNotFound                    = "InvalidAllocationID.NotFound"
	errCodeInvalidAssociationIDNotFound                   = "InvalidAssociationID.NotFound"
	errCodeInvalidAttachmentIDNotFound                    = "InvalidAttachmentID.NotFound"
	errCodeInvalidCapacityReservationIdNotFound           = "InvalidCapacityReservationId.NotFound'"
	ErrCodeInvalidCarrierGatewayIDNotFound                = "InvalidCarrierGatewayID.NotFound"
	errCodeInvalidClientVPNActiveAssociationNotFound      = "InvalidClientVpnActiveAssociationNotFound"
	errCodeInvalidClientVPNAssociationIdNotFound          = "InvalidClientVpnAssociationIdNotFound"
	errCodeInvalidClientVPNAuthorizationRuleNotFound      = "InvalidClientVpnEndpointAuthorizationRuleNotFound"
	errCodeInvalidClientVPNEndpointIdNotFound             = "InvalidClientVpnEndpointId.NotFound"
	errCodeInvalidClientVPNRouteNotFound                  = "InvalidClientVpnRouteNotFound"
	ErrCodeInvalidConnectionNotification                  = "InvalidConnectionNotification"
	errCodeInvalidConversionTaskIdMalformed               = "InvalidConversionTaskId.Malformed"
	errCodeInvalidCustomerGatewayIDNotFound               = "InvalidCustomerGatewayID.NotFound"
	errCodeInvalidDHCPOptionIDNotFound                    = "InvalidDhcpOptionID.NotFound"
	errCodeInvalidFleetIdNotFound                         = "InvalidFleetId.NotFound"
	errCodeInvalidFlowLogIdNotFound                       = "InvalidFlowLogId.NotFound"
	errCodeInvalidGatewayIDNotFound                       = "InvalidGatewayID.NotFound"
	errCodeInvalidGroupNotFound                           = "InvalidGroup.NotFound"
	errCodeInvalidHostIDNotFound                          = "InvalidHostID.NotFound"
	errCodeInvalidInstanceIDNotFound                      = "InvalidInstanceID.NotFound"
	errCodeInvalidInternetGatewayIDNotFound               = "InvalidInternetGatewayID.NotFound"
	errCodeInvalidKeyPairNotFound                         = "InvalidKeyPair.NotFound"
	errCodeInvalidLaunchTemplateIdMalformed               = "InvalidLaunchTemplateId.Malformed"
	errCodeInvalidLaunchTemplateIdNotFound                = "InvalidLaunchTemplateId.NotFound"
	errCodeInvalidLaunchTemplateIdVersionNotFound         = "InvalidLaunchTemplateId.VersionNotFound"
	errCodeInvalidLaunchTemplateNameNotFoundException     = "InvalidLaunchTemplateName.NotFoundException"
	errCodeInvalidNetworkACLEntryNotFound                 = "InvalidNetworkAclEntry.NotFound"
	errCodeInvalidNetworkACLIDNotFound                    = "InvalidNetworkAclID.NotFound"
	errCodeInvalidNetworkInterfaceIDNotFound              = "InvalidNetworkInterfaceID.NotFound"
	errCodeInvalidNetworkInsightsPathIdNotFound           = "InvalidNetworkInsightsPathId.NotFound"
	errCodeInvalidParameter                               = "InvalidParameter"
	errCodeInvalidParameterException                      = "InvalidParameterException"
	errCodeInvalidParameterValue                          = "InvalidParameterValue"
	errCodeInvalidPermissionDuplicate                     = "InvalidPermission.Duplicate"
	errCodeInvalidPermissionNotFound                      = "InvalidPermission.NotFound"
	errCodeInvalidPlacementGroupUnknown                   = "InvalidPlacementGroup.Unknown"
	errCodeInvalidPoolIDNotFound                          = "InvalidPoolID.NotFound"
	errCodeInvalidPrefixListIDNotFound                    = "InvalidPrefixListID.NotFound"
	errCodeInvalidPrefixListIdNotFound                    = "InvalidPrefixListId.NotFound"
	errCodeInvalidRouteNotFound                           = "InvalidRoute.NotFound"
	errCodeInvalidRouteTableIDNotFound                    = "InvalidRouteTableID.NotFound"
	errCodeInvalidRouteTableIdNotFound                    = "InvalidRouteTableId.NotFound"
	errCodeInvalidSecurityGroupIDNotFound                 = "InvalidSecurityGroupID.NotFound"
	errCodeInvalidServiceName                             = "InvalidServiceName"
	errCodeInvalidSnapshotInUse                           = "InvalidSnapshot.InUse"
	errCodeInvalidSnapshotNotFound                        = "InvalidSnapshot.NotFound"
	ErrCodeInvalidSpotDatafeedNotFound                    = "InvalidSpotDatafeed.NotFound"
	errCodeInvalidSpotFleetRequestConfig                  = "InvalidSpotFleetRequestConfig"
	errCodeInvalidSpotFleetRequestIdNotFound              = "InvalidSpotFleetRequestId.NotFound"
	errCodeInvalidSpotInstanceRequestIDNotFound           = "InvalidSpotInstanceRequestID.NotFound"
	errCodeInvalidSubnetCIDRReservationIDNotFound         = "InvalidSubnetCidrReservationID.NotFound"
	errCodeInvalidSubnetIDNotFound                        = "InvalidSubnetID.NotFound"
	errCodeInvalidSubnetIdNotFound                        = "InvalidSubnetId.NotFound"
	errCodeInvalidTransitGatewayAttachmentIDNotFound      = "InvalidTransitGatewayAttachmentID.NotFound"
	errCodeInvalidTransitGatewayConnectPeerIDNotFound     = "InvalidTransitGatewayConnectPeerID.NotFound"
	errCodeInvalidTransitGatewayIDNotFound                = "InvalidTransitGatewayID.NotFound"
	errCodeInvalidTransitGatewayMulticastDomainIdNotFound = "InvalidTransitGatewayMulticastDomainId.NotFound"
	errCodeInvalidVolumeNotFound                          = "InvalidVolume.NotFound"
	errCodeInvalidVPCCIDRBlockAssociationIDNotFound       = "InvalidVpcCidrBlockAssociationID.NotFound"
	errCodeInvalidVPCEndpointIdNotFound                   = "InvalidVpcEndpointId.NotFound"
	errCodeInvalidVPCEndpointNotFound                     = "InvalidVpcEndpoint.NotFound"
	errCodeInvalidVPCEndpointServiceIdNotFound            = "InvalidVpcEndpointServiceId.NotFound"
	errCodeInvalidVPCIDNotFound                           = "InvalidVpcID.NotFound"
	errCodeInvalidVPCPeeringConnectionIDNotFound          = "InvalidVpcPeeringConnectionID.NotFound"
	errCodeInvalidVPNConnectionIDNotFound                 = "InvalidVpnConnectionID.NotFound"
	errCodeInvalidVPNGatewayAttachmentNotFound            = "InvalidVpnGatewayAttachment.NotFound"
	errCodeInvalidVPNGatewayIDNotFound                    = "InvalidVpnGatewayID.NotFound"
	errCodeNatGatewayNotFound                             = "NatGatewayNotFound"
	errCodeResourceNotReady                               = "ResourceNotReady"
	errCodeSnapshotCreationPerVolumeRateExceeded          = "SnapshotCreationPerVolumeRateExceeded"
	errCodeUnsupportedOperation                           = "UnsupportedOperation"
	errCodeVolumeInUse                                    = "VolumeInUse"
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
