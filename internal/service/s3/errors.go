// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"fmt"
)

// Error code constants missing from AWS Go SDK:
// https://docs.aws.amazon.com/sdk-for-go/api/service/s3/#pkg-constants

const (
	errCodeAccessDenied                         = "AccessDenied"
	errCodeBucketNotEmpty                       = "BucketNotEmpty"
	errCodeInvalidArgument                      = "InvalidArgument"
	errCodeInvalidBucketState                   = "InvalidBucketState"
	errCodeInvalidRequest                       = "InvalidRequest"
	errCodeMalformedPolicy                      = "MalformedPolicy"
	errCodeMethodNotAllowed                     = "MethodNotAllowed"
	errCodeNoSuchBucket                         = "NoSuchBucket"
	errCodeNoSuchBucketPolicy                   = "NoSuchBucketPolicy"
	errCodeNoSuchConfiguration                  = "NoSuchConfiguration"
	errCodeNoSuchCORSConfiguration              = "NoSuchCORSConfiguration"
	errCodeNoSuchLifecycleConfiguration         = "NoSuchLifecycleConfiguration"
	errCodeNoSuchKey                            = "NoSuchKey"
	errCodeNoSuchPublicAccessBlockConfiguration = "NoSuchPublicAccessBlockConfiguration"
	errCodeNoSuchTagSet                         = "NoSuchTagSet"
	errCodeNoSuchTagSetError                    = "NoSuchTagSetError"
	errCodeNoSuchWebsiteConfiguration           = "NoSuchWebsiteConfiguration"
	errCodeNotImplemented                       = "NotImplemented"
	// errCodeObjectLockConfigurationNotFound should be used with tfawserr.ErrCodeContains, not tfawserr.ErrCodeEquals.
	// Reference: https://github.com/hashicorp/terraform-provider-aws/pull/26317
	errCodeObjectLockConfigurationNotFound           = "ObjectLockConfigurationNotFound"
	errCodeObjectLockConfigurationNotFoundError      = "ObjectLockConfigurationNotFoundError"
	errCodeOperationAborted                          = "OperationAborted"
	errCodeOwnershipControlsNotFoundError            = "OwnershipControlsNotFoundError"
	errCodeReplicationConfigurationNotFound          = "ReplicationConfigurationNotFoundError"
	errCodeServerSideEncryptionConfigurationNotFound = "ServerSideEncryptionConfigurationNotFoundError"
	errCodeUnsupportedArgument                       = "UnsupportedArgument"
	// errCodeXNotImplemented is returned from Third Party S3 implementations
	// and so far has been noticed with calls to GetBucketWebsite.
	// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/14645
	errCodeXNotImplemented = "XNotImplemented"
)

const (
	ErrMessageBucketAlreadyExists = "bucket already exists"
)

func errDirectoryBucket(err error) error {
	return fmt.Errorf("directory buckets are not supported: %w", err)
}
