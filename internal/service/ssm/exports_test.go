// Copyright IBM Corp. 2014, 2025
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
	ResourceParameterVersionLabels  = resourceParameterVersionLabels
	ResourcePatchBaseline           = resourcePatchBaseline
	ResourcePatchGroup              = resourcePatchGroup
	ResourceResourceDataSync        = resourceResourceDataSync
	ResourceServiceSetting          = resourceServiceSetting

	FindActivationByID                                 = findActivationByID
	FindAssociationByID                                = findAssociationByID
	FindDefaultPatchBaselineByOperatingSystem          = findDefaultPatchBaselineByOperatingSystem
	FindDefaultDefaultPatchBaselineIDByOperatingSystem = findDefaultDefaultPatchBaselineIDByOperatingSystem
	FindDocumentByName                                 = findDocumentByName
	FindMaintenanceWindowByID                          = findMaintenanceWindowByID
	FindMaintenanceWindowTargetByTwoPartKey            = findMaintenanceWindowTargetByTwoPartKey
	FindMaintenanceWindowTaskByTwoPartKey              = findMaintenanceWindowTaskByTwoPartKey
	FindParameterByName                                = findParameterByName
	FindParameterVersionLabels                         = findParameterVersionLabels
	FindPatchBaselineByID                              = findPatchBaselineByID
	FindPatchGroupByTwoPartKey                         = findPatchGroupByTwoPartKey
	FindResourceDataSyncByName                         = findResourceDataSyncByName
	FindServiceSettingByID                             = findServiceSettingByID

	ParameterVersionLabelsParseResourceID = parameterVersionLabelsParseResourceID
)
