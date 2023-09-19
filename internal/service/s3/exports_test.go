// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

// Exports for use in tests only.
var (
	DeleteAllObjectVersions           = deleteAllObjectVersions
	EmptyBucket                       = emptyBucket
	FindBucket                        = findBucket
	FindBucketAccelerateConfiguration = findBucketAccelerateConfiguration
	FindBucketPolicy                  = findBucketPolicy
	FindBucketVersioning              = findBucketVersioning
	FindBucketWebsite                 = findBucketWebsite
	FindObjectByBucketAndKey          = findObjectByBucketAndKey
	FindObjectLockConfiguration       = findObjectLockConfiguration
	SDKv1CompatibleCleanKey           = sdkv1CompatibleCleanKey
)
