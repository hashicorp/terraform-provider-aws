// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3

// Exports for use in tests only.
var (
	ResourceBucketABAC                              = newResourceBucketABAC
	ResourceBucketAccelerateConfiguration           = resourceBucketAccelerateConfiguration
	ResourceBucketACL                               = resourceBucketACL
	ResourceBucketAnalyticsConfiguration            = resourceBucketAnalyticsConfiguration
	ResourceBucketCorsConfiguration                 = resourceBucketCorsConfiguration
	ResourceBucketIntelligentTieringConfiguration   = resourceBucketIntelligentTieringConfiguration
	ResourceBucketInventory                         = resourceBucketInventory
	ResourceBucketLifecycleConfiguration            = newBucketLifecycleConfigurationResource
	ResourceBucketLogging                           = resourceBucketLogging
	ResourceBucketMetadataConfiguration             = newBucketMetadataConfigurationResource
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

	BucketUpdateTags                            = bucketUpdateTags
	BucketRegionalDomainName                    = bucketRegionalDomainName
	BucketWebsiteEndpointAndDomain              = bucketWebsiteEndpointAndDomain
	DeleteAllObjectVersions                     = deleteAllObjectVersions
	EmptyBucket                                 = emptyBucket
	FindAnalyticsConfiguration                  = findAnalyticsConfiguration
	FindBucket                                  = findBucket
	FindBucketABAC                              = findBucketABAC
	FindBucketACL                               = findBucketACL
	FindBucketAccelerateConfiguration           = findBucketAccelerateConfiguration
	FindBucketLifecycleConfiguration            = findBucketLifecycleConfiguration
	FindBucketMetadataConfigurationByTwoPartKey = findBucketMetadataConfigurationByTwoPartKey
	FindBucketNotificationConfiguration         = findBucketNotificationConfiguration
	FindBucketPolicy                            = findBucketPolicy
	FindBucketRequestPayment                    = findBucketRequestPayment
	FindBucketVersioning                        = findBucketVersioning
	FindBucketWebsite                           = findBucketWebsite
	FindCORSRules                               = findCORSRules
	FindIntelligentTieringConfiguration         = findIntelligentTieringConfiguration
	FindInventoryConfiguration                  = findInventoryConfiguration
	FindLoggingEnabled                          = findLoggingEnabled
	FindMetricsConfiguration                    = findMetricsConfiguration
	FindObjectByBucketAndKey                    = findObjectByBucketAndKey
	FindObjectLockConfiguration                 = findObjectLockConfiguration
	FindOwnershipControls                       = findOwnershipControls
	FindPublicAccessBlockConfiguration          = findPublicAccessBlockConfiguration
	FindReplicationConfiguration                = findReplicationConfiguration
	FindServerSideEncryptionConfiguration       = findServerSideEncryptionConfiguration
	HostedZoneIDForRegion                       = hostedZoneIDForRegion
	IsDirectoryBucket                           = isDirectoryBucket
	ObjectListTags                              = objectListTags
	ObjectUpdateTags                            = objectUpdateTags
	SDKv1CompatibleCleanKey                     = sdkv1CompatibleCleanKey
	ValidBucketName                             = validBucketName

	BucketPropagationTimeout       = bucketPropagationTimeout
	BucketVersioningStatusDisabled = bucketVersioningStatusDisabled
	ErrCodeBucketAlreadyExists     = errCodeBucketAlreadyExists
	ErrCodeBucketAlreadyOwnedByYou = errCodeBucketAlreadyOwnedByYou
	ErrCodeNoSuchCORSConfiguration = errCodeNoSuchCORSConfiguration
	LifecycleRuleStatusDisabled    = lifecycleRuleStatusDisabled
	LifecycleRuleStatusEnabled     = lifecycleRuleStatusEnabled

	NewObjectARN   = newObjectARN
	ParseObjectARN = parseObjectARN

	CreateResourceID          = createResourceID
	ParseResourceID           = parseResourceID
	CreateBucketACLResourceID = createBucketACLResourceID
	ParseBucketACLResourceID  = parseBucketACLResourceID

	DirectoryBucketNameSuffixRegexPattern = directoryBucketNameSuffixRegexPattern

	LifecycleConfigEqual = lifecycleConfigEqual
)

type (
	ObjectARN = objectARN
)
