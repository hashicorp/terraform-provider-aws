package ec2

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	multierror "github.com/hashicorp/go-multierror"
)

const (
	ErrCodeGatewayNotAttached           = "Gateway.NotAttached"
	ErrCodeInvalidAssociationIDNotFound = "InvalidAssociationID.NotFound"
	ErrCodeInvalidAttachmentIDNotFound  = "InvalidAttachmentID.NotFound"
	ErrCodeInvalidParameter             = "InvalidParameter"
	ErrCodeInvalidParameterException    = "InvalidParameterException"
	ErrCodeInvalidParameterValue        = "InvalidParameterValue"
)

const (
	ErrCodeInvalidCarrierGatewayIDNotFound = "InvalidCarrierGatewayID.NotFound"
)

const (
	ErrCodeClientInvalidHostIDNotFound = "Client.InvalidHostID.NotFound"
	ErrCodeInvalidHostIDNotFound       = "InvalidHostID.NotFound"
)

const (
	ErrCodeInvalidNetworkInterfaceIDNotFound = "InvalidNetworkInterfaceID.NotFound"
)

const (
	ErrCodeInvalidPrefixListIDNotFound = "InvalidPrefixListID.NotFound"
)

const (
	ErrCodeInvalidRouteNotFound        = "InvalidRoute.NotFound"
	ErrCodeInvalidRouteTableIdNotFound = "InvalidRouteTableId.NotFound"
	ErrCodeInvalidRouteTableIDNotFound = "InvalidRouteTableID.NotFound"
)

const (
	ErrCodeInvalidTransitGatewayIDNotFound = "InvalidTransitGatewayID.NotFound"
)

const (
	ErrCodeClientVPNEndpointIdNotFound        = "InvalidClientVpnEndpointId.NotFound"
	ErrCodeClientVPNAuthorizationRuleNotFound = "InvalidClientVpnEndpointAuthorizationRuleNotFound"
	ErrCodeClientVPNAssociationIdNotFound     = "InvalidClientVpnAssociationId.NotFound"
	ErrCodeClientVPNRouteNotFound             = "InvalidClientVpnRouteNotFound"
)

const (
	ErrCodeInvalidInstanceIDNotFound = "InvalidInstanceID.NotFound"
)

const (
	InvalidSecurityGroupIDNotFound = "InvalidSecurityGroupID.NotFound"
	InvalidGroupNotFound           = "InvalidGroup.NotFound"
)

const (
	ErrCodeInvalidSpotInstanceRequestIDNotFound = "InvalidSpotInstanceRequestID.NotFound"
)

const (
	ErrCodeInvalidSubnetIdNotFound = "InvalidSubnetId.NotFound"
	ErrCodeInvalidSubnetIDNotFound = "InvalidSubnetID.NotFound"
)

const (
	ErrCodeInvalidVPCIDNotFound = "InvalidVpcID.NotFound"
)

const (
	ErrCodeInvalidVPCEndpointIdNotFound        = "InvalidVpcEndpointId.NotFound"
	ErrCodeInvalidVPCEndpointNotFound          = "InvalidVpcEndpoint.NotFound"
	ErrCodeInvalidVPCEndpointServiceIdNotFound = "InvalidVpcEndpointServiceId.NotFound"
)

const (
	ErrCodeInvalidVPCPeeringConnectionIDNotFound = "InvalidVpcPeeringConnectionID.NotFound"
)

const (
	InvalidVPNGatewayAttachmentNotFound = "InvalidVpnGatewayAttachment.NotFound"
	InvalidVPNGatewayIDNotFound         = "InvalidVpnGatewayID.NotFound"
)

const (
	ErrCodeInvalidPermissionDuplicate = "InvalidPermission.Duplicate"
	ErrCodeInvalidPermissionMalformed = "InvalidPermission.Malformed"
	ErrCodeInvalidPermissionNotFound  = "InvalidPermission.NotFound"
)

const (
	ErrCodeInvalidFlowLogIdNotFound = "InvalidFlowLogId.NotFound"
)

const (
	ErrCodeInvalidPlacementGroupUnknown = "InvalidPlacementGroup.Unknown"
)

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
