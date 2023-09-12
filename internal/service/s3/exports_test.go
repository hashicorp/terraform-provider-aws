// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

// Exports for use in tests only.
var (
	DeleteAllObjectVersions  = deleteAllObjectVersions
	FindBucketPolicy         = findBucketPolicy
	FindObjectByBucketAndKey = findObjectByBucketAndKey
	SDKv1CompatibleCleanKey  = sdkv1CompatibleCleanKey
)
