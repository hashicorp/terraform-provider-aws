// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package guardduty

// Exports for use in tests only.
var (
	ResourceFilter                = resourceFilter
	ResourceInviteAccepter        = resourceInviteAccepter
	ResourceMalwareProtectionPlan = newMalwareProtectionPlanResource
	ResourcePublishingDestination = resourcePublishingDestination
	ResourceThreatIntelSet        = resourceThreatIntelSet

	FindDetectorByID                                 = findDetectorByID
	FindDetectorID                                   = findDetectorID
	FindDetectorFeatureByTwoPartKey                  = findDetectorFeatureByTwoPartKey
	FindFilterByTwoPartKey                           = findFilterByTwoPartKey
	FindIPSetByTwoPartKey                            = findIPSetByTwoPartKey
	FindMalwareProtectionPlanByID                    = findMalwareProtectionPlanByID
	FindMemberDetectorFeatureByThreePartKey          = findMemberDetectorFeatureByThreePartKey
	FindOrganizationConfigurationByID                = findOrganizationConfigurationByID
	FindOrganizationConfigurationFeatureByTwoPartKey = findOrganizationConfigurationFeatureByTwoPartKey
	FindPublishingDestinationByTwoPartKey            = findPublishingDestinationByTwoPartKey
	FindThreatIntelSetByTwoPartKey                   = findThreatIntelSetByTwoPartKey

	GetOrganizationAdminAccount = getOrganizationAdminAccount
)
