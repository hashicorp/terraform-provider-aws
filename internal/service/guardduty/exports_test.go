// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package guardduty

// Exports for use in tests only.
var (
	ResourceInviteAccepter        = resourceInviteAccepter
	ResourceMalwareProtectionPlan = newResourceMalwareProtectionPlan

	FindMemberDetectorFeatureByThreePartKey = findMemberDetectorFeatureByThreePartKey
)
