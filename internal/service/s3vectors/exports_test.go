// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3vectors

// Exports for use in tests only.
var (
	ResourceIndex              = newIndexResource
	ResourceVectorBucket       = newVectorBucketResource
	ResourceVectorBucketPolicy = newVectorBucketPolicyResource

	FindIndexByARN              = findIndexByARN
	FindVectorBucketByARN       = findVectorBucketByARN
	FindVectorBucketPolicyByARN = findVectorBucketPolicyByARN
)
