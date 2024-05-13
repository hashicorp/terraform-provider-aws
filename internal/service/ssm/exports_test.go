// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssm

// Exports for use in tests only.
var (
	ResourceActivation              = resourceActivation
	ResourceAssociation             = resourceAssociation
	ResourceDefaultPatchBaseline    = resourceDefaultPatchBaseline
	ResourceDocument                = resourceDocument
	ResourceMaintenanceWindow       = resourceMaintenanceWindow
	ResourceMaintenanceWindowTarget = resourceMaintenanceWindowTarget
	ResourceMaintenanceWindowTask   = resourceMaintenanceWindowTask
	ResourceParameter               = resourceParameter
	ResourcePatchBaseline           = resourcePatchBaseline

	FindActivationByID                                 = findActivationByID
	FindAssociationByID                                = findAssociationByID
	FindDefaultPatchBaselineByOperatingSystem          = findDefaultPatchBaselineByOperatingSystem
	FindDefaultDefaultPatchBaselineIDByOperatingSystem = findDefaultDefaultPatchBaselineIDByOperatingSystem
	FindDocumentByName                                 = findDocumentByName
	FindMaintenanceWindowByID                          = findMaintenanceWindowByID
	FindMaintenanceWindowTargetByID                    = findMaintenanceWindowTargetByID
	FindMaintenanceWindowTaskByTwoPartKey              = findMaintenanceWindowTaskByTwoPartKey
	FindParameterByName                                = findParameterByName
	FindPatchBaselineByID                              = findPatchBaselineByID
)
