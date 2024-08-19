// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3control

// Error code constants missing from AWS Go SDK:
// https://docs.aws.amazon.com/sdk-for-go/api/service/s3control/#pkg-constants
const (
	errCodeAccessGrantsLocationNotEmptyError    = "AccessGrantsLocationNotEmptyError"
	errCodeInvalidBucketState                   = "InvalidBucketState"
	errCodeInvalidIAMRole                       = "InvalidIamRole"
	errCodeInvalidRequest                       = "InvalidRequest"
	errCodeNoSuchAccessPoint                    = "NoSuchAccessPoint"
	errCodeNoSuchAccessPointPolicy              = "NoSuchAccessPointPolicy"
	errCodeNoSuchAsyncRequest                   = "NoSuchAsyncRequest"
	errCodeNoSuchBucket                         = "NoSuchBucket"
	errCodeNoSuchBucketPolicy                   = "NoSuchBucketPolicy"
	errCodeNoSuchLifecycleConfiguration         = "NoSuchLifecycleConfiguration"
	errCodeNoSuchMultiRegionAccessPoint         = "NoSuchMultiRegionAccessPoint"
	errCodeNoSuchOutpost                        = "NoSuchOutpost"
	errCodeNoSuchPublicAccessBlockConfiguration = "NoSuchPublicAccessBlockConfiguration"
	errCodeNoSuchTagSet                         = "NoSuchTagSet"
)
