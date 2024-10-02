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

	FindBackupVaultByName                   = findBackupVaultByName
	FindFrameworkByName                     = findFrameworkByName
	FindGlobalSettings                      = findGlobalSettings
	FindLogicallyAirGappedBackupVaultByName = findLogicallyAirGappedBackupVaultByName
	FindPlanByID                            = findPlanByID
	FindRegionSettings                      = findRegionSettings
	FindVaultAccessPolicyByName             = findVaultAccessPolicyByName
)
