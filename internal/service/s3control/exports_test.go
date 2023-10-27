// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3control

// Exports for use in tests only.
var (
	ResourceAccessGrant                   = newResourceAccessGrant
	ResourceAccessGrantsInstance          = newResourceAccessGrantsInstance
	ResourceAccessGrantsLocation          = newResourceAccessGrantsLocation
	ResourceAccessPoint                   = resourceAccessPoint
	ResourceAccessPointPolicy             = resourceAccessPointPolicy
	ResourceAccountPublicAccessBlock      = resourceAccountPublicAccessBlock
	ResourceBucket                        = resourceBucket
	ResourceBucketLifecycleConfiguration  = resourceBucketLifecycleConfiguration
	ResourceBucketPolicy                  = resourceBucketPolicy
	ResourceMultiRegionAccessPoint        = resourceMultiRegionAccessPoint
	ResourceMultiRegionAccessPointPolicy  = resourceMultiRegionAccessPointPolicy
	ResourceObjectLambdaAccessPoint       = resourceObjectLambdaAccessPoint
	ResourceObjectLambdaAccessPointPolicy = resourceObjectLambdaAccessPointPolicy
	ResourceStorageLensConfiguration      = resourceStorageLensConfiguration

	FindAccessGrantByTwoPartKey                            = findAccessGrantByTwoPartKey
	FindAccessGrantsInstance                               = findAccessGrantsInstance
	FindAccessGrantsLocationByTwoPartKey                   = findAccessGrantsLocationByTwoPartKey
	FindAccessPointByTwoPartKey                            = findAccessPointByTwoPartKey
	FindAccessPointPolicyAndStatusByTwoPartKey             = findAccessPointPolicyAndStatusByTwoPartKey
	FindBucketByTwoPartKey                                 = findBucketByTwoPartKey
	FindBucketLifecycleConfigurationByTwoPartKey           = findBucketLifecycleConfigurationByTwoPartKey
	FindBucketPolicyByTwoPartKey                           = findBucketPolicyByTwoPartKey
	FindMultiRegionAccessPointByTwoPartKey                 = findMultiRegionAccessPointByTwoPartKey
	FindMultiRegionAccessPointPolicyDocumentByTwoPartKey   = findMultiRegionAccessPointPolicyDocumentByTwoPartKey
	FindObjectLambdaAccessPointAliasByTwoPartKey           = findObjectLambdaAccessPointAliasByTwoPartKey
	FindObjectLambdaAccessPointConfigurationByTwoPartKey   = findObjectLambdaAccessPointConfigurationByTwoPartKey
	FindObjectLambdaAccessPointPolicyAndStatusByTwoPartKey = findObjectLambdaAccessPointPolicyAndStatusByTwoPartKey
	FindPublicAccessBlockByAccountID                       = findPublicAccessBlockByAccountID
	FindStorageLensConfigurationByAccountIDAndConfigID     = findStorageLensConfigurationByAccountIDAndConfigID
)
