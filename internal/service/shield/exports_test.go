// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package shield

// Exports for use in tests only.
var (
	ResourceDRTAccessRoleARNAssociation       = newDRTAccessRoleARNAssociationResource
	ResourceDRTAccessLogBucketAssociation     = newDRTAccessLogBucketAssociationResource
	ResourceApplicationLayerAutomaticResponse = newApplicationLayerAutomaticResponseResource
	ResourceProactiveEngagement               = newProactiveEngagementResource
	ResourceProtection                        = resourceProtection

	FindApplicationLayerAutomaticResponseByResourceARN = findApplicationLayerAutomaticResponseByResourceARN
	FindDRTLogBucketAssociation                        = findDRTLogBucketAssociation
	FindDRTRoleARNAssociation                          = findDRTRoleARNAssociation
	FindEmergencyContactSettings                       = findEmergencyContactSettings
	FindProtectionByID                                 = findProtectionByID
)
