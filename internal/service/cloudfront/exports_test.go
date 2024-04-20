// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront

// Exports for use in tests only.
var (
	ResourceCachePolicy                = resourceCachePolicy
	ResourceContinuousDeploymentPolicy = newContinuousDeploymentPolicyResource
	ResourceFieldLevelEncryptionConfig = resourceFieldLevelEncryptionConfig
	ResourceFunction                   = resourceFunction
	ResourceKeyValueStore              = newKeyValueStoreResource

	FindCachePolicyByID                = findCachePolicyByID
	FindFieldLevelEncryptionConfigByID = findFieldLevelEncryptionConfigByID
	FindFunctionByTwoPartKey           = findFunctionByTwoPartKey
	FindKeyValueStoreByName            = findKeyValueStoreByName
	FindPublicKeyByID                  = findPublicKeyByID
)
