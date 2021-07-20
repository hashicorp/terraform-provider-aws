package storagegateway

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/storagegateway"
)

// Error code constants missing from AWS Go SDK:
// https://docs.aws.amazon.com/sdk-for-go/api/service/storagegateway/#pkg-constants

const (
	FileSystemAssociationNotFound = "FileSystemAssociationNotFound"
)

func InvalidGatewayRequestErrCodeEquals(err error, errCode string) bool {
	var igrex *storagegateway.InvalidGatewayRequestException
	if errors.As(err, &igrex) {
		if err := igrex.Error_; err != nil {
			return aws.StringValue(err.ErrorCode) == errCode
		}
	}
	return false
}
