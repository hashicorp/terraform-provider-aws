// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package s3

import (
	"fmt"
	"net/http"

	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
)

// Error code constants missing from AWS Go SDK:
// https://docs.aws.amazon.com/sdk-for-go/api/service/s3/#pkg-constants

const (
	errCodeAccessDenied                         = "AccessDenied"
	errCodeAuthorizationHeaderMalformed         = "AuthorizationHeaderMalformed"
	errCodeBucketAlreadyExists                  = "BucketAlreadyExists"
	errCodeBucketAlreadyOwnedByYou              = "BucketAlreadyOwnedByYou"
	errCodeBucketNotEmpty                       = "BucketNotEmpty"
	errCodeIllegalLocationConstraintException   = "IllegalLocationConstraintException"
	errCodeInvalidArgument                      = "InvalidArgument"
	errCodeInvalidBucketState                   = "InvalidBucketState"
	errCodeInvalidRequest                       = "InvalidRequest"
	errCodeMalformedPolicy                      = "MalformedPolicy"
	errCodeMalformedXML                         = "MalformedXML"
	errCodeMetadataConfigurationNotFound        = "MetadataConfigurationNotFound"
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
	// Reference: https://github.com/hashicorp/terraform-provider-aws/pull/26317.
	errCodeObjectLockConfigurationNotFound           = "ObjectLockConfigurationNotFound"
	errCodeObjectLockConfigurationNotFoundError      = "ObjectLockConfigurationNotFoundError"
	errCodeOperationAborted                          = "OperationAborted"
	errCodeOwnershipControlsNotFoundError            = "OwnershipControlsNotFoundError"
	errCodePermanentRedirect                         = "PermanentRedirect"
	errCodeReplicationConfigurationNotFound          = "ReplicationConfigurationNotFoundError"
	errCodeServerSideEncryptionConfigurationNotFound = "ServerSideEncryptionConfigurationNotFoundError"
	errCodeUnsupportedArgument                       = "UnsupportedArgument"
	// errCodeXNotImplemented, errCodeUnsupportedOperation are returned from third-party S3 API implementations.
	// References:
	//   https://github.com/hashicorp/terraform-provider-aws/issues/14645.
	//   https://github.com/hashicorp/terraform-provider-aws/pull/37801.
	errCodeXNotImplemented      = "XNotImplemented"
	errCodeUnsupportedOperation = "UnsupportedOperation"
)

func errDirectoryBucket(err error) error {
	return fmt.Errorf("directory buckets are not supported: %w", err)
}

func errBucketRegionMismatch(bucket string, err error) error {
	if err == nil {
		return nil
	}

	if tfawserr.ErrHTTPStatusCodeEquals(err, http.StatusMovedPermanently) ||
		tfawserr.ErrCodeEquals(err, errCodePermanentRedirect) ||
		tfawserr.ErrCodeEquals(err, errCodeAuthorizationHeaderMalformed) {
		return fmt.Errorf("S3 Bucket (%s) was redirected to another Region; verify that the bucket name is the actual S3 bucket name and that the resource's `region` argument matches the bucket Region: %w", bucket, err)
	}

	return err
}
