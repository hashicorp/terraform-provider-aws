// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kms

// Exports for use in tests only.
var (
	ResourceAlias          = resourceAlias
	ResourceCiphertext     = resourceCiphertext
	ResourceCustomKeyStore = resourceCustomKeyStore
	ResourceExternalKey    = resourceExternalKey
	ResourceGrant          = resourceGrant
	ResourceKey            = resourceKey

	FindAliasByName           = findAliasByName
	FindCustomKeyStoreByID    = findCustomKeyStoreByID
	FindGrantByTwoPartKey     = findGrantByTwoPartKey
	FindKeyByID               = findKeyByID
	FindKeyPolicyByTwoPartKey = findKeyPolicyByTwoPartKey
	GrantParseResourceID      = grantParseResourceID
)
