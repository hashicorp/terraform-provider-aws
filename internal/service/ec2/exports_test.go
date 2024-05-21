// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

// Exports for use in tests only.
var (
	ResourceCarrierGateway                  = resourceCarrierGateway
	ResourceClientVPNAuthorizationRule      = resourceClientVPNAuthorizationRule
	ResourceClientVPNEndpoint               = resourceClientVPNEndpoint
	ResourceClientVPNNetworkAssociation     = resourceClientVPNNetworkAssociation
	ResourceClientVPNRoute                  = resourceClientVPNRoute
	ResourceCustomerGateway                 = resourceCustomerGateway
	ResourceDefaultNetworkACL               = resourceDefaultNetworkACL
	ResourceDefaultRouteTable               = resourceDefaultRouteTable
	ResourceEBSFastSnapshotRestore          = newEBSFastSnapshotRestoreResource
	ResourceEIP                             = resourceEIP
	ResourceEIPAssociation                  = resourceEIPAssociation
	ResourceEIPDomainName                   = newEIPDomainNameResource
	ResourceInstanceConnectEndpoint         = newInstanceConnectEndpointResource
	ResourceInstanceMetadataDefaults        = newInstanceMetadataDefaultsResource
	ResourceIPAMOrganizationAdminAccount    = resourceIPAMOrganizationAdminAccount
	ResourceKeyPair                         = resourceKeyPair
	ResourceMainRouteTableAssociation       = resourceMainRouteTableAssociation
	ResourceNetworkACL                      = resourceNetworkACL
	ResourceNetworkACLRule                  = resourceNetworkACLRule
	ResourceNetworkInterface                = resourceNetworkInterface
	ResourceRoute                           = resourceRoute
	ResourceRouteTable                      = resourceRouteTable
	ResourceSecurityGroupEgressRule         = newSecurityGroupEgressRuleResource
	ResourceSecurityGroupIngressRule        = newSecurityGroupIngressRuleResource
	ResourceTag                             = resourceTag
	ResourceTransitGatewayPeeringAttachment = resourceTransitGatewayPeeringAttachment
	ResourceVPNConnection                   = resourceVPNConnection
	ResourceVPNConnectionRoute              = resourceVPNConnectionRoute
	ResourceVPNGateway                      = resourceVPNGateway
	ResourceVPNGatewayAttachment            = resourceVPNGatewayAttachment
	ResourceVPNGatewayRoutePropagation      = resourceVPNGatewayRoutePropagation

	CustomFiltersSchema                                    = customFiltersSchema
	FindAvailabilityZonesV2                                = findAvailabilityZonesV2
	FindCarrierGatewayByID                                 = findCarrierGatewayByID
	FindClientVPNAuthorizationRuleByThreePartKey           = findClientVPNAuthorizationRuleByThreePartKey
	FindClientVPNEndpointByID                              = findClientVPNEndpointByID
	FindClientVPNNetworkAssociationByTwoPartKey            = findClientVPNNetworkAssociationByTwoPartKey
	FindClientVPNRouteByThreePartKey                       = findClientVPNRouteByThreePartKey
	FindCustomerGatewayByID                                = findCustomerGatewayByID
	FindEBSVolumeAttachment                                = findVolumeAttachment
	FindEIPByAllocationID                                  = findEIPByAllocationID
	FindEIPByAssociationID                                 = findEIPByAssociationID
	FindEIPDomainNameAttributeByAllocationID               = findEIPDomainNameAttributeByAllocationID
	FindFastSnapshotRestoreByTwoPartKey                    = findFastSnapshotRestoreByTwoPartKey
	FindInstanceMetadataDefaults                           = findInstanceMetadataDefaults
	FindIPAMByID                                           = findIPAMByID
	FindIPAMPoolAllocationByTwoPartKey                     = findIPAMPoolAllocationByTwoPartKey
	FindIPAMPoolByID                                       = findIPAMPoolByID
	FindIPAMPoolCIDRByTwoPartKey                           = findIPAMPoolCIDRByTwoPartKey
	FindIPAMResourceDiscoveryAssociationByID               = findIPAMResourceDiscoveryAssociationByID
	FindIPAMResourceDiscoveryByID                          = findIPAMResourceDiscoveryByID
	FindIPAMScopeByID                                      = findIPAMScopeByID
	FindKeyPairByName                                      = findKeyPairByName
	FindMainRouteTableAssociationByID                      = findMainRouteTableAssociationByID
	FindNetworkACLByIDV2                                   = findNetworkACLByIDV2
	FindNetworkInterfaceByIDV2                             = findNetworkInterfaceByIDV2
	FindRouteByIPv4DestinationV2                           = findRouteByIPv4Destination
	FindRouteByIPv6DestinationV2                           = findRouteByIPv6Destination
	FindRouteByPrefixListIDDestinationV2                   = findRouteByPrefixListIDDestination
	FindRouteTableAssociationByIDV2                        = findRouteTableAssociationByID
	FindRouteTableByIDV2                                   = findRouteTableByID
	FindVolumeAttachmentInstanceByID                       = findVolumeAttachmentInstanceByID
	FindVPCEndpointByIDV2                                  = findVPCEndpointByIDV2
	FindVPCEndpointConnectionByServiceIDAndVPCEndpointIDV2 = findVPCEndpointConnectionByServiceIDAndVPCEndpointIDV2
	FindVPCEndpointConnectionNotificationByIDV2            = findVPCEndpointConnectionNotificationByIDV2
	FindVPCEndpointRouteTableAssociationExistsV2           = findVPCEndpointRouteTableAssociationExistsV2
	FindVPCEndpointSecurityGroupAssociationExistsV2        = findVPCEndpointSecurityGroupAssociationExistsV2
	FindVPCEndpointServiceConfigurationByIDV2              = findVPCEndpointServiceConfigurationByIDV2
	FindVPCEndpointServicePermissionV2                     = findVPCEndpointServicePermissionV2
	FindVPCEndpointSubnetAssociationExistsV2               = findVPCEndpointSubnetAssociationExistsV2
	FindVPCIPv6CIDRBlockAssociationByIDV2                  = findVPCIPv6CIDRBlockAssociationByIDV2
	FindVPNConnectionByID                                  = findVPNConnectionByID
	FindVPNConnectionRouteByTwoPartKey                     = findVPNConnectionRouteByTwoPartKey
	FindVPNGatewayByID                                     = findVPNGatewayByID
	FindVPNGatewayVPCAttachmentByTwoPartKey                = findVPNGatewayVPCAttachmentByTwoPartKey
	FindVPNGatewayRoutePropagationExistsV2                 = findVPNGatewayRoutePropagationExists
	FlattenNetworkInterfacePrivateIPAddresses              = flattenNetworkInterfacePrivateIPAddresses
	IPAMServicePrincipal                                   = ipamServicePrincipal
	NewAttributeFilterList                                 = newAttributeFilterList
	NewAttributeFilterListV2                               = newAttributeFilterListV2
	NewCustomFilterList                                    = newCustomFilterList
	NewTagFilterList                                       = newTagFilterList
	ProtocolForValue                                       = protocolForValue
	StopInstance                                           = stopInstance
	StopEBSVolumeAttachmentInstance                        = stopVolumeAttachmentInstance
	UpdateTags                                             = updateTags
	UpdateTagsV2                                           = updateTagsV2
)

type (
	IPProtocol = ipProtocol
)
