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

	AliasARNToKeyARN          = aliasARNToKeyARN
	AliasNamePrefix           = aliasNamePrefix
	FindCustomKeyStoreByID    = findCustomKeyStoreByID
	FindGrantByTwoPartKey     = findGrantByTwoPartKey
	FindKeyPolicyByTwoPartKey = findKeyPolicyByTwoPartKey
	GrantParseResourceID      = grantParseResourceID
	KeyARNOrIDEqual           = keyARNOrIDEqual
	PropagationTimeout        = propagationTimeout
	PolicyNameDefault         = policyNameDefault
	SecretRemovedMessage      = secretRemovedMessage
)
