// Copyright IBM Corp. 2014, 2026
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
