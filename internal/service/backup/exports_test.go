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
	ResourceVault                   = resourceVault
	ResourceVaultLockConfiguration  = resourceVaultLockConfiguration
	ResourceVaultNotifications      = resourceVaultNotifications
	ResourceVaultPolicy             = resourceVaultPolicy

	FindBackupVaultByName                   = findBackupVaultByName // nosemgrep:ci.backup-in-var-name
	FindFrameworkByName                     = findFrameworkByName
	FindGlobalSettings                      = findGlobalSettings
	FindLogicallyAirGappedBackupVaultByName = findLogicallyAirGappedBackupVaultByName // nosemgrep:ci.backup-in-var-name
	FindPlanByID                            = findPlanByID
	FindRegionSettings                      = findRegionSettings
	FindReportPlanByName                    = findReportPlanByName
	FindSelectionByTwoPartKey               = findSelectionByTwoPartKey
	FindVaultAccessPolicyByName             = findVaultAccessPolicyByName
	FindVaultNotificationsByName            = findVaultNotificationsByName
)
