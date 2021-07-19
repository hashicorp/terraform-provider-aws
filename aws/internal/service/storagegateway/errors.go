package storagegateway

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/storagegateway"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
)

// OperationErrCodeEquals returns true if the error matches all these conditions:
//  * err is of type awserr.Error and represents storagegateway.InternalServerError or storagegateway.InvalidGatewayRequestException
//  * Error_.ErrorCode equals one of the passed codes
// See https://docs.aws.amazon.com/storagegateway/latest/userguide/AWSStorageGatewayAPI.html#APIErrorResponses for details.
func OperationErrCodeEquals(err error, codes ...string) bool {
	if tfawserr.ErrCodeEquals(err, storagegateway.ErrCodeInternalServerError) {
		var inner *storagegateway.InternalServerError

		if ok := errors.As(err, &inner); ok {
			if err := inner.Error_; err != nil {
				for _, code := range codes {
					if aws.StringValue(err.ErrorCode) == code {
						return true
					}
				}
			}
		}
	} else if tfawserr.ErrCodeEquals(err, storagegateway.ErrCodeInvalidGatewayRequestException) {
		var inner *storagegateway.InvalidGatewayRequestException

		if ok := errors.As(err, &inner); ok {
			if err := inner.Error_; err != nil {
				for _, code := range codes {
					if aws.StringValue(err.ErrorCode) == code {
						return true
					}
				}
			}
		}
	}

	return false
}
