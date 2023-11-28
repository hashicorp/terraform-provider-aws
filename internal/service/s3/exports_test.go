// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

// Exports for use in tests only.
var (
	ResourceDirectoryBucket = newDirectoryBucketResource

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
	FindLifecycleRules                    = findLifecycleRules
	FindLoggingEnabled                    = findLoggingEnabled
	FindMetricsConfiguration              = findMetricsConfiguration
	FindObjectByBucketAndKey              = findObjectByBucketAndKey
	FindObjectLockConfiguration           = findObjectLockConfiguration
	FindOwnershipControls                 = findOwnershipControls
	FindPublicAccessBlockConfiguration    = findPublicAccessBlockConfiguration
	FindReplicationConfiguration          = findReplicationConfiguration
	FindServerSideEncryptionConfiguration = findServerSideEncryptionConfiguration
	SDKv1CompatibleCleanKey               = sdkv1CompatibleCleanKey

	ErrCodeNoSuchCORSConfiguration = errCodeNoSuchCORSConfiguration
	LifecycleRuleStatusDisabled    = lifecycleRuleStatusDisabled
	LifecycleRuleStatusEnabled     = lifecycleRuleStatusEnabled
)
