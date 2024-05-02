// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kms

// Exports for use in tests only.
var (
	ResourceAlias          = resourceAlias
	ResourceCiphertext     = resourceCiphertext
	ResourceCustomKeyStore = resourceCustomKeyStore

	FindAliasByName        = findAliasByName
	FindCustomKeyStoreByID = findCustomKeyStoreByID
)
