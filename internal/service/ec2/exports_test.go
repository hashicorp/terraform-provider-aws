// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

// Exports for use in tests only.
var (
	ResourceDefaultNetworkACL               = resourceDefaultNetworkACL
	ResourceDefaultRouteTable               = resourceDefaultRouteTable
	ResourceEBSFastSnapshotRestore          = newResourceEBSFastSnapshotRestore
	ResourceInstanceConnectEndpoint         = newResourceInstanceConnectEndpoint
	ResourceNetworkACL                      = resourceNetworkACL
	ResourceNetworkACLRule                  = resourceNetworkACLRule
	ResourceRoute                           = resourceRoute
	ResourceRouteTable                      = resourceRouteTable
	ResourceSecurityGroupEgressRule         = newSecurityGroupEgressRuleResource
	ResourceSecurityGroupIngressRule        = newSecurityGroupIngressRuleResource
	ResourceTag                             = resourceTag
	ResourceTransitGatewayPeeringAttachment = resourceTransitGatewayPeeringAttachment

	CustomFiltersSchema            = customFiltersSchema
	FindEBSFastSnapshotRestoreByID = findEBSFastSnapshotRestoreByID
	FindNetworkACLByIDV2           = findNetworkACLByIDV2
	NewAttributeFilterList         = newAttributeFilterList
	NewCustomFilterList            = newCustomFilterList
	NewTagFilterList               = newTagFilterList
	ProtocolForValue               = protocolForValue
	StopInstance                   = stopInstance
	UpdateTags                     = updateTags
	UpdateTagsV2                   = updateTagsV2
)

type NormalizeIPProtocol = normalizeIPProtocol
