package ec2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	multierror "github.com/hashicorp/go-multierror"
)

const (
	ErrCodeAuthFailure                                    = "AuthFailure"
	ErrCodeClientInvalidHostIDNotFound                    = "Client.InvalidHostID.NotFound"
	ErrCodeDefaultSubnetAlreadyExistsInAvailabilityZone   = "DefaultSubnetAlreadyExistsInAvailabilityZone"
	ErrCodeDependencyViolation                            = "DependencyViolation"
	ErrCodeGatewayNotAttached                             = "Gateway.NotAttached"
	ErrCodeIncorrectState                                 = "IncorrectState"
	ErrCodeInvalidAMIIDNotFound                           = "InvalidAMIID.NotFound"
	ErrCodeInvalidAMIIDUnavailable                        = "InvalidAMIID.Unavailable"
	ErrCodeInvalidAddressNotFound                         = "InvalidAddress.NotFound"
	ErrCodeInvalidAllocationIDNotFound                    = "InvalidAllocationID.NotFound"
	ErrCodeInvalidAssociationIDNotFound                   = "InvalidAssociationID.NotFound"
	ErrCodeInvalidAttachmentIDNotFound                    = "InvalidAttachmentID.NotFound"
	ErrCodeInvalidCapacityReservationIdNotFound           = "InvalidCapacityReservationId.NotFound'"
	ErrCodeInvalidCarrierGatewayIDNotFound                = "InvalidCarrierGatewayID.NotFound"
	ErrCodeInvalidClientVpnActiveAssociationNotFound      = "InvalidClientVpnActiveAssociationNotFound"
	ErrCodeInvalidClientVpnAssociationIdNotFound          = "InvalidClientVpnAssociationIdNotFound"
	ErrCodeInvalidClientVpnAuthorizationRuleNotFound      = "InvalidClientVpnEndpointAuthorizationRuleNotFound"
	ErrCodeInvalidClientVpnEndpointIdNotFound             = "InvalidClientVpnEndpointId.NotFound"
	ErrCodeInvalidClientVpnRouteNotFound                  = "InvalidClientVpnRouteNotFound"
	ErrCodeInvalidConnectionNotification                  = "InvalidConnectionNotification"
	ErrCodeInvalidCustomerGatewayIDNotFound               = "InvalidCustomerGatewayID.NotFound"
	ErrCodeInvalidDhcpOptionIDNotFound                    = "InvalidDhcpOptionID.NotFound"
	ErrCodeInvalidFlowLogIdNotFound                       = "InvalidFlowLogId.NotFound"
	ErrCodeInvalidGatewayIDNotFound                       = "InvalidGatewayID.NotFound"
	ErrCodeInvalidGroupNotFound                           = "InvalidGroup.NotFound"
	ErrCodeInvalidHostIDNotFound                          = "InvalidHostID.NotFound"
	ErrCodeInvalidInstanceIDNotFound                      = "InvalidInstanceID.NotFound"
	ErrCodeInvalidInternetGatewayIDNotFound               = "InvalidInternetGatewayID.NotFound"
	ErrCodeInvalidKeyPairNotFound                         = "InvalidKeyPair.NotFound"
	ErrCodeInvalidLaunchTemplateIdMalformed               = "InvalidLaunchTemplateId.Malformed"
	ErrCodeInvalidLaunchTemplateIdNotFound                = "InvalidLaunchTemplateId.NotFound"
	ErrCodeInvalidLaunchTemplateIdVersionNotFound         = "InvalidLaunchTemplateId.VersionNotFound"
	ErrCodeInvalidLaunchTemplateNameNotFoundException     = "InvalidLaunchTemplateName.NotFoundException"
	ErrCodeInvalidNetworkAclEntryNotFound                 = "InvalidNetworkAclEntry.NotFound"
	ErrCodeInvalidNetworkAclIDNotFound                    = "InvalidNetworkAclID.NotFound"
	ErrCodeInvalidNetworkInterfaceIDNotFound              = "InvalidNetworkInterfaceID.NotFound"
	ErrCodeInvalidNetworkInsightsPathIdNotFound           = "InvalidNetworkInsightsPathId.NotFound"
	ErrCodeInvalidParameter                               = "InvalidParameter"
	ErrCodeInvalidParameterException                      = "InvalidParameterException"
	ErrCodeInvalidParameterValue                          = "InvalidParameterValue"
	ErrCodeInvalidPermissionDuplicate                     = "InvalidPermission.Duplicate"
	ErrCodeInvalidPermissionMalformed                     = "InvalidPermission.Malformed"
	ErrCodeInvalidPermissionNotFound                      = "InvalidPermission.NotFound"
	ErrCodeInvalidPlacementGroupUnknown                   = "InvalidPlacementGroup.Unknown"
	ErrCodeInvalidPoolIDNotFound                          = "InvalidPoolID.NotFound"
	ErrCodeInvalidPrefixListIDNotFound                    = "InvalidPrefixListID.NotFound"
	ErrCodeInvalidRouteNotFound                           = "InvalidRoute.NotFound"
	ErrCodeInvalidRouteTableIDNotFound                    = "InvalidRouteTableID.NotFound"
	ErrCodeInvalidRouteTableIdNotFound                    = "InvalidRouteTableId.NotFound"
	ErrCodeInvalidSecurityGroupIDNotFound                 = "InvalidSecurityGroupID.NotFound"
	ErrCodeInvalidSnapshotInUse                           = "InvalidSnapshot.InUse"
	ErrCodeInvalidSnapshotNotFound                        = "InvalidSnapshot.NotFound"
	ErrCodeInvalidSpotDatafeedNotFound                    = "InvalidSpotDatafeed.NotFound"
	ErrCodeInvalidSpotFleetRequestConfig                  = "InvalidSpotFleetRequestConfig"
	ErrCodeInvalidSpotFleetRequestIdNotFound              = "InvalidSpotFleetRequestId.NotFound"
	ErrCodeInvalidSpotInstanceRequestIDNotFound           = "InvalidSpotInstanceRequestID.NotFound"
	ErrCodeInvalidSubnetCIDRReservationIDNotFound         = "InvalidSubnetCidrReservationID.NotFound"
	ErrCodeInvalidSubnetIDNotFound                        = "InvalidSubnetID.NotFound"
	ErrCodeInvalidSubnetIdNotFound                        = "InvalidSubnetId.NotFound"
	ErrCodeInvalidTransitGatewayAttachmentIDNotFound      = "InvalidTransitGatewayAttachmentID.NotFound"
	ErrCodeInvalidTransitGatewayConnectPeerIDNotFound     = "InvalidTransitGatewayConnectPeerID.NotFound"
	ErrCodeInvalidTransitGatewayIDNotFound                = "InvalidTransitGatewayID.NotFound"
	ErrCodeInvalidTransitGatewayMulticastDomainIdNotFound = "InvalidTransitGatewayMulticastDomainId.NotFound"
	ErrCodeInvalidVolumeNotFound                          = "InvalidVolume.NotFound"
	ErrCodeInvalidVpcCidrBlockAssociationIDNotFound       = "InvalidVpcCidrBlockAssociationID.NotFound"
	ErrCodeInvalidVpcEndpointIdNotFound                   = "InvalidVpcEndpointId.NotFound"
	ErrCodeInvalidVpcEndpointNotFound                     = "InvalidVpcEndpoint.NotFound"
	ErrCodeInvalidVpcEndpointServiceIdNotFound            = "InvalidVpcEndpointServiceId.NotFound"
	ErrCodeInvalidVpcIDNotFound                           = "InvalidVpcID.NotFound"
	ErrCodeInvalidVpcPeeringConnectionIDNotFound          = "InvalidVpcPeeringConnectionID.NotFound"
	ErrCodeInvalidVpnConnectionIDNotFound                 = "InvalidVpnConnectionID.NotFound"
	ErrCodeInvalidVpnGatewayAttachmentNotFound            = "InvalidVpnGatewayAttachment.NotFound"
	ErrCodeInvalidVpnGatewayIDNotFound                    = "InvalidVpnGatewayID.NotFound"
	ErrCodeNatGatewayNotFound                             = "NatGatewayNotFound"
	ErrCodeUnsupportedOperation                           = "UnsupportedOperation"
	ErrCodeVolumeInUse                                    = "VolumeInUse"
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
