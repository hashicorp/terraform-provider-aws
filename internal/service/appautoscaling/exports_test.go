// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appautoscaling

// Exports for use in tests only.
var (
	ResourcePolicy          = resourcePolicy
	ResourceScheduledAction = resourceScheduledAction
	ResourceTarget          = resourceTarget

	FindScalingPolicyByFourPartKey   = findScalingPolicyByFourPartKey
	FindScheduledActionByFourPartKey = findScheduledActionByFourPartKey
	ValidPolicyImportInput           = validPolicyImportInput
)
