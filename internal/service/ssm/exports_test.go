// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssm

// Exports for use in tests only.
var (
	ResourceActivation              = resourceActivation
	ResourceAssociation             = resourceAssociation
	ResourceDefaultPatchBaseline    = resourceDefaultPatchBaseline
	ResourceDocument                = resourceDocument
	ResourceMaintenanceWindowTarget = resourceMaintenanceWindowTarget
	ResourceMaintenanceWindowTask   = resourceMaintenanceWindowTask
	ResourcePatchBaseline           = resourcePatchBaseline

	FindActivationByID                    = findActivationByID
	FindAssociationByID                   = findAssociationByID
	FindDocumentByName                    = findDocumentByName
	FindMaintenanceWindowTargetByID       = findMaintenanceWindowTargetByID
	FindMaintenanceWindowTaskByTwoPartKey = findMaintenanceWindowTaskByTwoPartKey
	FindPatchBaselineByID                 = findPatchBaselineByID
)
