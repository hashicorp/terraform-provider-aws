// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package computeoptimizer

// Exports for use in tests only.
var (
	ResourceEnrollmentStatus          = newEnrollmentStatusResource
	ResourceRecommendationPreferences = newRecommendationPreferencesResource

	FindEnrollmentStatus                        = findEnrollmentStatus
	FindRecommendationPreferencesByThreePartKey = findRecommendationPreferencesByThreePartKey
)
