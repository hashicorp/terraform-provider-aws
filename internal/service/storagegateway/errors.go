// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package storagegateway

import (
	awstypes "github.com/aws/aws-sdk-go-v2/service/storagegateway/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
)

// Operation error code constants missing from AWS Go SDK: https://docs.aws.amazon.com/sdk-for-go/api/service/storagegateway/#pkg-constants.
// See https://docs.aws.amazon.com/storagegateway/latest/userguide/AWSStorageGatewayAPI.html#APIOperationErrorCodes for details.
const (
	operationErrCodeFileShareNotFound             = "FileShareNotFound"
	operationErrCodeFileSystemAssociationNotFound = "FileSystemAssociationNotFound"
	operationErrCodeGatewayNotFound               = "GatewayNotFound"
)

// operationErrorCode returns the operation error code from the specified error:
//   - err is of type awserr.Error and represents a storagegateway.InternalServerError or storagegateway.InvalidGatewayRequestException
//   - Error_ is not nil
//
// See https://docs.aws.amazon.com/storagegateway/latest/userguide/AWSStorageGatewayAPI.html#APIErrorResponses for details.
func operationErrorCode(err error) string {
	if v, ok := errs.As[*awstypes.InternalServerError](err); ok && v.Error_ != nil {
		return string(v.Error_.ErrorCode)
	}

	if v, ok := errs.As[*awstypes.InvalidGatewayRequestException](err); ok && v.Error_ != nil {
		return string(v.Error_.ErrorCode)
	}

	return ""
}
