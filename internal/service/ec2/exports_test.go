// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

// Exports for use in tests only.
var (
	ResourceEBSFastSnapshotRestore   = newResourceEBSFastSnapshotRestore
	ResourceInstanceConnectEndpoint  = newResourceInstanceConnectEndpoint
	ResourceSecurityGroupEgressRule  = newResourceSecurityGroupEgressRule
	ResourceSecurityGroupIngressRule = newResourceSecurityGroupIngressRule
	ResourceTag                      = resourceTag

	FindEBSFastSnapshotRestoreByID = findEBSFastSnapshotRestoreByID

	UpdateTags   = updateTags
	UpdateTagsV2 = updateTagsV2

	StopInstance = stopInstance
)
