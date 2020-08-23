package ec2

// Common.
const (
	ErrCodeInvalidParameterException = "InvalidParameterException"
	ErrCodeInvalidParameterValue     = "InvalidParameterValue"
)

// Client VPN.
const (
	ErrCodeClientVpnEndpointIdNotFound        = "InvalidClientVpnEndpointId.NotFound"
	ErrCodeClientVpnAuthorizationRuleNotFound = "InvalidClientVpnEndpointAuthorizationRuleNotFound"
	ErrCodeClientVpnAssociationIdNotFound     = "InvalidClientVpnAssociationId.NotFound"
	ErrCodeClientVpnRouteNotFound             = "InvalidClientVpnRouteNotFound"
)

// Security Group.
const (
	InvalidSecurityGroupIDNotFound = "InvalidSecurityGroupID.NotFound"
	InvalidGroupNotFound           = "InvalidGroup.NotFound"
)

// Route and Route Table.
const (
	ErrCodeInvalidRouteNotFound        = "InvalidRoute.NotFound"
	ErrCodeInvalidRouteTableIDNotFound = "InvalidRouteTableID.NotFound"
)

// Transit Gateway.
const (
	ErrCodeInvalidTransitGatewayIDNotFound = "InvalidTransitGatewayID.NotFound"
)

// VPN Gateway.
const (
	InvalidVpnGatewayAttachmentNotFound = "InvalidVpnGatewayAttachment.NotFound"
	InvalidVpnGatewayIDNotFound         = "InvalidVpnGatewayID.NotFound"
)
