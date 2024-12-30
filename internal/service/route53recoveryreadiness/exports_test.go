// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53recoveryreadiness

// Exports for use in tests only.
var (
	ResourceCell           = resourceCell
	ResourceReadinessCheck = resourceReadinessCheck
	ResourceRecoveryGroup  = resourceRecoveryGroup
	ResourceResourceSet    = resourceResourceSet

	FindCellByName           = findCellByName
	FindReadinessCheckByName = findReadinessCheckByName
	FindRecoveryGroupByName  = findRecoveryGroupByName
	FindResourceSetByName    = findResourceSetByName
)
