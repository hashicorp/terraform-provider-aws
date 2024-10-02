// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package backup

// Exports for use in tests only.
var (
	ResourceLogicallyAirGappedVault = newLogicallyAirGappedVaultResource

	FindBackupVaultByName                   = findBackupVaultByName
	FindLogicallyAirGappedBackupVaultByName = findLogicallyAirGappedBackupVaultByName
	FindVaultAccessPolicyByName             = findVaultAccessPolicyByName
)
