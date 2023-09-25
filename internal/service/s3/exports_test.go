// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

// Exports for use in tests only.
var (
	DeleteAllObjectVersions               = deleteAllObjectVersions
	EmptyBucket                           = emptyBucket
	FindAnalyticsConfiguration            = findAnalyticsConfiguration
	FindBucket                            = findBucket
	FindBucketACL                         = findBucketACL
	FindBucketAccelerateConfiguration     = findBucketAccelerateConfiguration
	FindBucketPolicy                      = findBucketPolicy
	FindBucketVersioning                  = findBucketVersioning
	FindBucketWebsite                     = findBucketWebsite
	FindCORSRules                         = findCORSRules
	FindObjectByBucketAndKey              = findObjectByBucketAndKey
	FindObjectLockConfiguration           = findObjectLockConfiguration
	FindServerSideEncryptionConfiguration = findServerSideEncryptionConfiguration
	SDKv1CompatibleCleanKey               = sdkv1CompatibleCleanKey

	ErrCodeNoSuchCORSConfiguration = errCodeNoSuchCORSConfiguration
)
