// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package guardduty

// Exports for use in tests only.
var (
	ResourceFilter                = resourceFilter
	ResourceInviteAccepter        = resourceInviteAccepter
	ResourceMalwareProtectionPlan = newMalwareProtectionPlanResource
	ResourcePublishingDestination = resourcePublishingDestination

	FindDetectorByID                        = findDetectorByID
	FindDetectorFeatureByTwoPartKey         = findDetectorFeatureByTwoPartKey
	FindMalwareProtectionPlanByID           = findMalwareProtectionPlanByID
	FindMemberDetectorFeatureByThreePartKey = findMemberDetectorFeatureByThreePartKey

	GetOrganizationAdminAccount = getOrganizationAdminAccount
)
