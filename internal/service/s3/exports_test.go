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
	FindBucketNotificationConfiguration   = findBucketNotificationConfiguration
	FindBucketPolicy                      = findBucketPolicy
	FindBucketRequestPayment              = findBucketRequestPayment
	FindBucketVersioning                  = findBucketVersioning
	FindBucketWebsite                     = findBucketWebsite
	FindCORSRules                         = findCORSRules
	FindIntelligentTieringConfiguration   = findIntelligentTieringConfiguration
	FindInventoryConfiguration            = findInventoryConfiguration
	FindLoggingEnabled                    = findLoggingEnabled
	FindMetricsConfiguration              = findMetricsConfiguration
	FindObjectByBucketAndKey              = findObjectByBucketAndKey
	FindObjectLockConfiguration           = findObjectLockConfiguration
	FindOwnershipControls                 = findOwnershipControls
	FindServerSideEncryptionConfiguration = findServerSideEncryptionConfiguration
	SDKv1CompatibleCleanKey               = sdkv1CompatibleCleanKey

	ErrCodeNoSuchCORSConfiguration = errCodeNoSuchCORSConfiguration
)
