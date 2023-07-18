// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3

// Error code constants missing from AWS Go SDK:
// https://docs.aws.amazon.com/sdk-for-go/api/service/s3/#pkg-constants

const (
	errCodeAccessDenied                         = "AccessDenied"
	errCodeBucketNotEmpty                       = "BucketNotEmpty"
	errCodeInvalidBucketState                   = "InvalidBucketState"
	errCodeInvalidRequest                       = "InvalidRequest"
	errCodeMalformedPolicy                      = "MalformedPolicy"
	errCodeMethodNotAllowed                     = "MethodNotAllowed"
	ErrCodeNoSuchBucketPolicy                   = "NoSuchBucketPolicy"
	errCodeNoSuchConfiguration                  = "NoSuchConfiguration"
	ErrCodeNoSuchCORSConfiguration              = "NoSuchCORSConfiguration"
	ErrCodeNoSuchLifecycleConfiguration         = "NoSuchLifecycleConfiguration"
	ErrCodeNoSuchPublicAccessBlockConfiguration = "NoSuchPublicAccessBlockConfiguration"
	errCodeNoSuchTagSet                         = "NoSuchTagSet"
	errCodeNoSuchTagSetError                    = "NoSuchTagSetError"
	ErrCodeNoSuchWebsiteConfiguration           = "NoSuchWebsiteConfiguration"
	errCodeNotImplemented                       = "NotImplemented"
	// errCodeObjectLockConfigurationNotFound should be used with tfawserr.ErrCodeContains, not tfawserr.ErrCodeEquals.
	// Reference: https://github.com/hashicorp/terraform-provider-aws/pull/26317
	errCodeObjectLockConfigurationNotFound           = "ObjectLockConfigurationNotFound"
	errCodeOperationAborted                          = "OperationAborted"
	ErrCodeReplicationConfigurationNotFound          = "ReplicationConfigurationNotFoundError"
	ErrCodeServerSideEncryptionConfigurationNotFound = "ServerSideEncryptionConfigurationNotFoundError"
	errCodeUnsupportedArgument                       = "UnsupportedArgument"
	// errCodeXNotImplemented is returned from Third Party S3 implementations
	// and so far has been noticed with calls to GetBucketWebsite.
	// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/14645
	errCodeXNotImplemented = "XNotImplemented"
)

const (
	ErrMessageBucketAlreadyExists = "bucket already exists"
)
