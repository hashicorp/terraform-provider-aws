// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

// Exports for use in tests only.
var (
	ResourceBucketAccelerateConfiguration           = resourceBucketAccelerateConfiguration
	ResourceBucketACL                               = resourceBucketACL
	ResourceBucketAnalyticsConfiguration            = resourceBucketAnalyticsConfiguration
	ResourceBucketCorsConfiguration                 = resourceBucketCorsConfiguration
	ResourceBucketIntelligentTieringConfiguration   = resourceBucketIntelligentTieringConfiguration
	ResourceBucketInventory                         = resourceBucketInventory
	ResourceBucketLifecycleConfiguration            = resourceBucketLifecycleConfiguration
	ResourceBucketLogging                           = resourceBucketLogging
	ResourceBucketMetric                            = resourceBucketMetric
	ResourceBucketNotification                      = resourceBucketNotification
	ResourceBucketObjectLockConfiguration           = resourceBucketObjectLockConfiguration
	ResourceBucketObject                            = resourceBucketObject
	ResourceBucketOwnershipControls                 = resourceBucketOwnershipControls
	ResourceBucketPolicy                            = resourceBucketPolicy
	ResourceBucketPublicAccessBlock                 = resourceBucketPublicAccessBlock
	ResourceBucketReplicationConfiguration          = resourceBucketReplicationConfiguration
	ResourceBucketRequestPaymentConfiguration       = resourceBucketRequestPaymentConfiguration
	ResourceBucketServerSideEncryptionConfiguration = resourceBucketServerSideEncryptionConfiguration
	ResourceBucketVersioning                        = resourceBucketVersioning
	ResourceBucketWebsiteConfiguration              = resourceBucketWebsiteConfiguration
	ResourceDirectoryBucket                         = newDirectoryBucketResource
	ResourceObjectCopy                              = resourceObjectCopy

	BucketUpdateTags                      = bucketUpdateTags
	BucketRegionalDomainName              = bucketRegionalDomainName
	BucketWebsiteEndpointAndDomain        = bucketWebsiteEndpointAndDomain
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
	HostedZoneIDForRegion                 = hostedZoneIDForRegion
	IsDirectoryBucket                     = isDirectoryBucket
	ObjectListTags                        = objectListTags
	ObjectUpdateTags                      = objectUpdateTags
	SDKv1CompatibleCleanKey               = sdkv1CompatibleCleanKey
	ValidBucketName                       = validBucketName

	BucketPropagationTimeout       = bucketPropagationTimeout
	BucketVersioningStatusDisabled = bucketVersioningStatusDisabled
	ErrCodeBucketAlreadyExists     = errCodeBucketAlreadyExists
	ErrCodeBucketAlreadyOwnedByYou = errCodeBucketAlreadyOwnedByYou
	ErrCodeNoSuchCORSConfiguration = errCodeNoSuchCORSConfiguration
	LifecycleRuleStatusDisabled    = lifecycleRuleStatusDisabled
	LifecycleRuleStatusEnabled     = lifecycleRuleStatusEnabled
)
