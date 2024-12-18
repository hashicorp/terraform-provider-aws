// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package networkmanager

// Exports for use in tests only.
var (
	ResourceAttachmentAccepter                   = resourceAttachmentAccepter
	ResourceConnectAttachment                    = resourceConnectAttachment
	ResourceConnection                           = resourceConnection
	ResourceCoreNetwork                          = resourceCoreNetwork
	ResourceCustomerGatewayAssociation           = resourceCustomerGatewayAssociation
	ResourceDevice                               = resourceDevice
	ResourceDirectConnectGatewayAttachment       = newDirectConnectGatewayAttachmentResource
	ResourceGlobalNetwork                        = resourceGlobalNetwork
	ResourceLink                                 = resourceLink
	ResourceLinkAssociation                      = resourceLinkAssociation
	ResourceSite                                 = resourceSite
	ResourceSiteToSiteVPNAttachment              = resourceSiteToSiteVPNAttachment
	ResourceTransitGatewayConnectPeerAssociation = resourceTransitGatewayConnectPeerAssociation
	ResourceTransitGatewayPeering                = resourceTransitGatewayPeering
	ResourceTransitGatewayRegistration           = resourceTransitGatewayRegistration
	ResourceTransitGatewayRouteTableAttachment   = resourceTransitGatewayRouteTableAttachment
	ResourceVPCAttachment                        = resourceVPCAttachment

	FindConnectAttachmentByID                            = findConnectAttachmentByID
	FindConnectionByTwoPartKey                           = findConnectionByTwoPartKey
	FindConnectPeerByID                                  = findConnectPeerByID
	FindCoreNetworkByID                                  = findCoreNetworkByID
	FindCoreNetworkPolicyByTwoPartKey                    = findCoreNetworkPolicyByTwoPartKey
	FindCustomerGatewayAssociationByTwoPartKey           = findCustomerGatewayAssociationByTwoPartKey
	FindDeviceByTwoPartKey                               = findDeviceByTwoPartKey
	FindDirectConnectGatewayAttachmentByID               = findDirectConnectGatewayAttachmentByID
	FindGlobalNetworkByID                                = findGlobalNetworkByID
	FindLinkAssociationByThreePartKey                    = findLinkAssociationByThreePartKey
	FindLinkByTwoPartKey                                 = findLinkByTwoPartKey
	FindSiteByTwoPartKey                                 = findSiteByTwoPartKey
	FindSiteToSiteVPNAttachmentByID                      = findSiteToSiteVPNAttachmentByID
	FindTransitGatewayConnectPeerAssociationByTwoPartKey = findTransitGatewayConnectPeerAssociationByTwoPartKey
	FindTransitGatewayPeeringByID                        = findTransitGatewayPeeringByID
	FindTransitGatewayRegistrationByTwoPartKey           = findTransitGatewayRegistrationByTwoPartKey
	FindTransitGatewayRouteTableAttachmentByID           = findTransitGatewayRouteTableAttachmentByID
	FindVPCAttachmentByID                                = findVPCAttachmentByID

	CustomerGatewayAssociationParseResourceID           = customerGatewayAssociationParseResourceID
	LinkAssociationParseResourceID                      = linkAssociationParseResourceID
	TransitGatewayConnectPeerAssociationParseResourceID = transitGatewayConnectPeerAssociationParseResourceID
	TransitGatewayRegistrationParseResourceID           = transitGatewayRegistrationParseResourceID
)
