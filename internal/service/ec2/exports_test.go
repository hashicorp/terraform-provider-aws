// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

// Exports for use in tests only.
var (
	ResourceDefaultNetworkACL        = resourceDefaultNetworkACL
	ResourceEBSFastSnapshotRestore   = newResourceEBSFastSnapshotRestore
	ResourceInstanceConnectEndpoint  = newResourceInstanceConnectEndpoint
	ResourceNetworkACL               = resourceNetworkACL
	ResourceSecurityGroupEgressRule  = newResourceSecurityGroupEgressRule
	ResourceSecurityGroupIngressRule = newResourceSecurityGroupIngressRule
	ResourceTag                      = resourceTag

	CustomFiltersSchema            = customFiltersSchema
	FindEBSFastSnapshotRestoreByID = findEBSFastSnapshotRestoreByID
	FindNetworkACLByIDV2           = findNetworkACLByIDV2
	NewCustomFilterList            = newCustomFilterList
	NewTagFilterList               = newTagFilterList
	StopInstance                   = stopInstance
	UpdateTags                     = updateTags
	UpdateTagsV2                   = updateTagsV2
)
