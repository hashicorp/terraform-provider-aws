// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glacier

// Exports for use in tests only.
var (
	ResourceVault     = resourceVault
	ResourceVaultLock = resourceVaultLock

	FindVaultByName     = findVaultByName
	FindVaultLockByName = findVaultLockByName
)
