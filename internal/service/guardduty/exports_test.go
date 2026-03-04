// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package guardduty

// Exports for use in tests only.
var (
	ResourceDetector                 = resourceDetector
	ResourceFilter                   = resourceFilter
	ResourceInviteAccepter           = resourceInviteAccepter
	ResourceIPSet                    = resourceIPSet
	ResourceMalwareProtectionPlan    = newMalwareProtectionPlanResource
	ResourceOrganizationAdminAccount = resourceOrganizationAdminAccount
	ResourcePublishingDestination    = resourcePublishingDestination
	ResourceThreatIntelSet           = resourceThreatIntelSet

	FindDetectorByID                                 = findDetectorByID
	FindDetectorID                                   = findDetectorID
	FindDetectorFeatureByTwoPartKey                  = findDetectorFeatureByTwoPartKey
	FindFilterByTwoPartKey                           = findFilterByTwoPartKey
	FindIPSetByTwoPartKey                            = findIPSetByTwoPartKey
	FindMalwareProtectionPlanByID                    = findMalwareProtectionPlanByID
	FindMemberDetectorFeatureByThreePartKey          = findMemberDetectorFeatureByThreePartKey
	FindOrganizationAdminAccountByID                 = findOrganizationAdminAccountByID
	FindOrganizationConfigurationByID                = findOrganizationConfigurationByID
	FindOrganizationConfigurationFeatureByTwoPartKey = findOrganizationConfigurationFeatureByTwoPartKey
	FindPublishingDestinationByTwoPartKey            = findPublishingDestinationByTwoPartKey
	FindThreatIntelSetByTwoPartKey                   = findThreatIntelSetByTwoPartKey
)
