// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

// Exports for use in tests only.
var (
	ResourceCustomerGateway                 = resourceCustomerGateway
	ResourceDefaultNetworkACL               = resourceDefaultNetworkACL
	ResourceDefaultRouteTable               = resourceDefaultRouteTable
	ResourceEBSFastSnapshotRestore          = newEBSFastSnapshotRestoreResource
	ResourceEIP                             = resourceEIP
	ResourceEIPAssociation                  = resourceEIPAssociation
	ResourceInstanceConnectEndpoint         = newInstanceConnectEndpointResource
	ResourceInstanceMetadataDefaults        = newInstanceMetadataDefaultsResource
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

	CustomFiltersSchema                       = customFiltersSchema
	FindEIPByAllocationID                     = findEIPByAllocationID
	FindEIPByAssociationID                    = findEIPByAssociationID
	FindFastSnapshotRestoreByTwoPartKey       = findFastSnapshotRestoreByTwoPartKey
	FindInstanceMetadataDefaults              = findInstanceMetadataDefaults
	FindKeyPairByName                         = findKeyPairByName
	FindNetworkACLByIDV2                      = findNetworkACLByIDV2
	FindNetworkInterfaceByIDV2                = findNetworkInterfaceByIDV2
	FlattenNetworkInterfacePrivateIPAddresses = flattenNetworkInterfacePrivateIPAddresses
	NewAttributeFilterList                    = newAttributeFilterList
	NewCustomFilterList                       = newCustomFilterList
	NewTagFilterList                          = newTagFilterList
	ProtocolForValue                          = protocolForValue
	StopInstance                              = stopInstance
	UpdateTags                                = updateTags
	UpdateTagsV2                              = updateTagsV2
)

type (
	IPProtocol = ipProtocol
)
