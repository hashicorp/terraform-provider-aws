// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package backup

// Exports for use in tests only.
var (
	ResourceFramework               = resourceFramework
	ResourceGlobalSettings          = resourceGlobalSettings
	ResourceLogicallyAirGappedVault = newLogicallyAirGappedVaultResource
	ResourcePlan                    = resourcePlan
	ResourceRegionSettings          = resourceRegionSettings
	ResourceReportPlan              = resourceReportPlan
	ResourceSelection               = resourceSelection
	ResourceVaultLockConfiguration  = resourceVaultLockConfiguration

	FindBackupVaultByName                   = findBackupVaultByName
	FindFrameworkByName                     = findFrameworkByName
	FindGlobalSettings                      = findGlobalSettings
	FindLogicallyAirGappedBackupVaultByName = findLogicallyAirGappedBackupVaultByName
	FindPlanByID                            = findPlanByID
	FindRegionSettings                      = findRegionSettings
	FindReportPlanByName                    = findReportPlanByName
	FindSelectionByTwoPartKey               = findSelectionByTwoPartKey
	FindVaultAccessPolicyByName             = findVaultAccessPolicyByName
)
