// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package storagegateway

import (
	awstypes "github.com/aws/aws-sdk-go-v2/service/storagegateway/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
)

// Operation error code constants missing from AWS Go SDK: https://docs.aws.amazon.com/sdk-for-go/api/service/storagegateway/#pkg-constants.
// See e.g. https://docs.aws.amazon.com/filegateway/latest/files3/AWSStorageGatewayAPI.html#APIOperationErrorCodes for details.
const (
	operationErrCodeFileShareNotFound             awstypes.ErrorCode = "FileShareNotFound"
	operationErrCodeFileSystemAssociationNotFound awstypes.ErrorCode = "FileSystemAssociationNotFound"
	operationErrCodeGatewayNotFound               awstypes.ErrorCode = "GatewayNotFound"
)

// operationErrorCode returns the operation error code from the specified error:
//   - err represents an InternalServerError or InvalidGatewayRequestException
//   - Error_ is not nil
//
// See https://docs.aws.amazon.com/storagegateway/latest/userguide/AWSStorageGatewayAPI.html#APIErrorResponses for details.
func operationErrorCode(err error) awstypes.ErrorCode {
	if v, ok := errs.As[*awstypes.InternalServerError](err); ok && v.Error_ != nil {
		return v.Error_.ErrorCode
	}

	if v, ok := errs.As[*awstypes.InvalidGatewayRequestException](err); ok && v.Error_ != nil {
		return v.Error_.ErrorCode
	}

	return ""
}

// The API returns multiple responses for a missing gateway.
func isGatewayNotFoundErr(err error) bool {
	if operationErrorCode(err) == operationErrCodeGatewayNotFound {
		return true
	}

	if tfawserr.ErrCodeEquals(err, string(awstypes.ErrorCodeGatewayNotFound)) {
		return true
	}

	if errs.IsAErrorMessageContains[*awstypes.InvalidGatewayRequestException](err, "The specified gateway was not found") {
		return true
	}

	return false
}

// The API returns multiple responses for a missing volume.
func isVolumeNotFoundErr(err error) bool {
	if tfawserr.ErrCodeEquals(err, string(awstypes.ErrorCodeVolumeNotFound)) {
		return true
	}

	if errs.IsAErrorMessageContains[*awstypes.InvalidGatewayRequestException](err, "The specified volume was not found") {
		return true
	}

	return false
}
