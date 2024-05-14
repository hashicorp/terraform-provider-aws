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
	ResourceEIPDomainName                   = newEIPDomainNameResource
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
	FindEBSVolumeAttachment                   = findVolumeAttachment
	FindEIPByAllocationID                     = findEIPByAllocationID
	FindEIPByAssociationID                    = findEIPByAssociationID
	FindEIPDomainNameAttributeByAllocationID  = findEIPDomainNameAttributeByAllocationID
	FindFastSnapshotRestoreByTwoPartKey       = findFastSnapshotRestoreByTwoPartKey
	FindInstanceMetadataDefaults              = findInstanceMetadataDefaults
	FindKeyPairByName                         = findKeyPairByName
	FindNetworkACLByIDV2                      = findNetworkACLByIDV2
	FindNetworkInterfaceByIDV2                = findNetworkInterfaceByIDV2
	FindVolumeAttachmentInstanceByID          = findVolumeAttachmentInstanceByID
	FlattenNetworkInterfacePrivateIPAddresses = flattenNetworkInterfacePrivateIPAddresses
	NewAttributeFilterList                    = newAttributeFilterList
	NewAttributeFilterListV2                  = newAttributeFilterListV2
	NewCustomFilterList                       = newCustomFilterList
	NewTagFilterList                          = newTagFilterList
	ProtocolForValue                          = protocolForValue
	StopInstance                              = stopInstance
	StopEBSVolumeAttachmentInstance           = stopVolumeAttachmentInstance
	UpdateTags                                = updateTags
	UpdateTagsV2                              = updateTagsV2
)

type (
	IPProtocol = ipProtocol
)
