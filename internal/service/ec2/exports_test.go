// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

// Exports for use in tests only.
var (
	ResourceCustomerGateway                 = resourceCustomerGateway
	ResourceDefaultNetworkACL               = resourceDefaultNetworkACL
	ResourceDefaultRouteTable               = resourceDefaultRouteTable
	ResourceEBSFastSnapshotRestore          = newResourceEBSFastSnapshotRestore
	ResourceInstanceConnectEndpoint         = newResourceInstanceConnectEndpoint
	ResourceInstanceMetadataDefaults        = newInstanceMetadataDefaultsResource
	ResourceNetworkACL                      = resourceNetworkACL
	ResourceNetworkACLRule                  = resourceNetworkACLRule
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

	CustomFiltersSchema            = customFiltersSchema
	FindEBSFastSnapshotRestoreByID = findEBSFastSnapshotRestoreByID
	FindInstanceMetadataDefaults   = findInstanceMetadataDefaults
	FindNetworkACLByIDV2           = findNetworkACLByIDV2
	NewAttributeFilterList         = newAttributeFilterList
	NewCustomFilterList            = newCustomFilterList
	NewTagFilterList               = newTagFilterList
	ProtocolForValue               = protocolForValue
	StopInstance                   = stopInstance
	UpdateTags                     = updateTags
	UpdateTagsV2                   = updateTagsV2
)

type (
	IPProtocol = ipProtocol
)
