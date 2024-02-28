// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package shield

// Exports for use in tests only.
var (
	ResourceDRTAccessRoleARNAssociation       = newDRTAccessRoleARNAssociationResource
	ResourceDRTAccessLogBucketAssociation     = newDRTAccessLogBucketAssociationResource
	ResourceApplicationLayerAutomaticResponse = newResourceApplicationLayerAutomaticResponse
	ResourceProactiveEngagement               = newProactiveEngagementResource

	FindDRTLogBucketAssociation  = findDRTLogBucketAssociation
	FindDRTRoleARNAssociation    = findDRTRoleARNAssociation
	FindEmergencyContactSettings = findEmergencyContactSettings
)
