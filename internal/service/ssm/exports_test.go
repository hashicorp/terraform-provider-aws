// Copyright IBM Corp. 2014, 2026
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
	ResourcePatchGroup              = resourcePatchGroup
	ResourceResourceDataSync        = resourceResourceDataSync
	ResourceResourcePolicy          = newResourcePolicyResource
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
	FindPatchBaselineByID                              = findPatchBaselineByID
	FindPatchGroupByTwoPartKey                         = findPatchGroupByTwoPartKey
	FindResourceDataSyncByName                         = findResourceDataSyncByName
	FindResourcePolicyByTwoPartKey                     = findResourcePolicyByTwoPartKey
	FindServiceSettingByID                             = findServiceSettingByID
)
