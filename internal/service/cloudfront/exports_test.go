// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront

// Exports for use in tests only.
var (
	ResourceContinuousDeploymentPolicy = newResourceContinuousDeploymentPolicy
	ResourceKeyValueStore              = newKeyValueStoreResource

	FindKeyValueStoreByName = findKeyValueStoreByName
	FindPublicKeyByID       = findPublicKeyByID
)
