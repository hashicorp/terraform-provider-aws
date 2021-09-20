package storagegateway

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/storagegateway"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	operationErrCodeFileShareNotFound = "FileShareNotFound"
	fileSystemAssociationNotFound     = "fileSystemAssociationNotFound"
)

// operationErrorCode returns the operation error code from the specified error:
//  * err is of type awserr.Error and represents a storagegateway.InternalServerError or storagegateway.InvalidGatewayRequestException
//  * Error_ is not nil
// See https://docs.aws.amazon.com/storagegateway/latest/userguide/AWSStorageGatewayAPI.html#APIErrorResponses for details.
func operationErrorCode(err error) string {
	if inner := (*storagegateway.InternalServerError)(nil); errors.As(err, &inner) && inner.Error_ != nil {
		return aws.StringValue(inner.Error_.ErrorCode)
	}

	if inner := (*storagegateway.InvalidGatewayRequestException)(nil); errors.As(err, &inner) && inner.Error_ != nil {
		return aws.StringValue(inner.Error_.ErrorCode)
	}

	return ""
}

// Error code constants missing from AWS Go SDK:
// https://docs.aws.amazon.com/sdk-for-go/api/service/storagegateway/#pkg-constants

func invalidGatewayRequestErrCodeEquals(err error, errCode string) bool {
	var igrex *storagegateway.InvalidGatewayRequestException
	if errors.As(err, &igrex) {
		if err := igrex.Error_; err != nil {
			return aws.StringValue(err.ErrorCode) == errCode
		}
	}
	return false
}
