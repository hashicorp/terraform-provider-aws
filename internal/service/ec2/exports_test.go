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
	FindAvailabilityZones                                = findAvailabilityZones
	FindCarrierGatewayByID                                 = findCarrierGatewayByID
	FindClientVPNAuthorizationRuleByThreePartKey           = findClientVPNAuthorizationRuleByThreePartKey
	FindClientVPNEndpointByID                              = findClientVPNEndpointByID
	FindClientVPNNetworkAssociationByTwoPartKey            = findClientVPNNetworkAssociationByTwoPartKey
	FindClientVPNRouteByThreePartKey                       = findClientVPNRouteByThreePartKey
	FindCapacityReservationByID                            = findCapacityReservationByID
	FindEBSVolumeAttachment                                = findVolumeAttachment
	FindEIPByAllocationID                                  = findEIPByAllocationID
	FindEIPByAssociationID                                 = findEIPByAssociationID
	FindEIPDomainNameAttributeByAllocationID               = findEIPDomainNameAttributeByAllocationID
	FindFastSnapshotRestoreByTwoPartKey                    = findFastSnapshotRestoreByTwoPartKey
	FindFleetByID                                          = findFleetByID
	FindHostByID                                           = findHostByID
	FindInstanceMetadataDefaults                           = findInstanceMetadataDefaults
	FindInstanceStateByID                                  = findInstanceStateByID
	FindKeyPairByName                                      = findKeyPairByName
	FindLaunchTemplateByID                                 = findLaunchTemplateByID
	FindNetworkACLByIDV2                                   = findNetworkACLByIDV2
	FindNetworkInterfaceByIDV2                             = findNetworkInterfaceByIDV2
	FindPlacementGroupByName                               = findPlacementGroupByName
	FindPublicIPv4Pools                                    = findPublicIPv4Pools
	FindRouteByIPv4DestinationV2                           = findRouteByIPv4DestinationV2
	FindRouteByIPv6DestinationV2                           = findRouteByIPv6DestinationV2
	FindRouteByPrefixListIDDestinationV2                   = findRouteByPrefixListIDDestinationV2
	FindRouteTableAssociationByIDV2                        = findRouteTableAssociationByIDV2
	FindRouteTableByIDV2                                   = findRouteTableByIDV2
	FindSpotDatafeedSubscription                           = findSpotDatafeedSubscription
	FindSpotFleetRequestByID                               = findSpotFleetRequestByID
	FindSpotFleetRequests                                  = findSpotFleetRequests
	FindSpotInstanceRequestByID                            = findSpotInstanceRequestByID
	FindSubnetsV2                                          = findSubnetsV2
	FindVolumeAttachmentInstanceByID                       = findVolumeAttachmentInstanceByID
	FindVPCEndpointByIDV2                                  = findVPCEndpointByIDV2
	FindVPCEndpointConnectionByServiceIDAndVPCEndpointIDV2 = findVPCEndpointConnectionByServiceIDAndVPCEndpointIDV2
	FindVPCEndpointConnectionNotificationByIDV2            = findVPCEndpointConnectionNotificationByIDV2
	FindVPCEndpointRouteTableAssociationExistsV2           = findVPCEndpointRouteTableAssociationExistsV2
	FindVPCEndpointSecurityGroupAssociationExistsV2        = findVPCEndpointSecurityGroupAssociationExistsV2
	FindVPCEndpointServiceConfigurationByIDV2              = findVPCEndpointServiceConfigurationByIDV2
	FindVPCEndpointServicePermissionV2                     = findVPCEndpointServicePermissionV2
	FindVPCEndpointSubnetAssociationExistsV2               = findVPCEndpointSubnetAssociationExistsV2
	FindVPNGatewayRoutePropagationExistsV2                 = findVPNGatewayRoutePropagationExistsV2
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

	WaitVolumeAttachmentCreated = waitVolumeAttachmentCreated
)

type (
	IPProtocol = ipProtocol
)
