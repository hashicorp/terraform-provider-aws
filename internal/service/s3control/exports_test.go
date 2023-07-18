// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3control

// Exports for use in tests only.
var (
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
)
