// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kms

// Exports for use in tests only.
var (
	ResourceAlias              = resourceAlias
	ResourceCiphertext         = resourceCiphertext
	ResourceCustomKeyStore     = resourceCustomKeyStore
	ResourceExternalKey        = resourceExternalKey
	ResourceGrant              = resourceGrant
	ResourceKey                = resourceKey
	ResourceKeyPolicy          = resourceKeyPolicy
	ResourceReplicaExternalKey = resourceReplicaExternalKey
	ResourceReplicaKey         = resourceReplicaKey

	AliasNamePrefix           = aliasNamePrefix
	FindCustomKeyStoreByID    = findCustomKeyStoreByID
	FindGrantByTwoPartKey     = findGrantByTwoPartKey
	FindKeyPolicyByTwoPartKey = findKeyPolicyByTwoPartKey
	GrantParseResourceID      = grantParseResourceID
	KMSPropagationTimeout     = kmsPropagationTimeout // nosemgrep:ci.kms-in-var-name
	PolicyNameDefault         = policyNameDefault
	SecretRemovedMessage      = secretRemovedMessage
)
