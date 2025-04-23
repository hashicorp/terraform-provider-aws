// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package guardduty

// Exports for use in tests only.
var (
	ResourceFilter                = resourceFilter
	ResourceInviteAccepter        = resourceInviteAccepter
	ResourceMalwareProtectionPlan = newResourceMalwareProtectionPlan
	ResourcePublishingDestination = resourcePublishingDestination

	FindMemberDetectorFeatureByThreePartKey = findMemberDetectorFeatureByThreePartKey

	GetOrganizationAdminAccount = getOrganizationAdminAccount
)
