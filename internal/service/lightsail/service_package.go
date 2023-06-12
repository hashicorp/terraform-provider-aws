package lightsail

import (
	"context"

	aws_sdkv1 "github.com/aws/aws-sdk-go/aws"
	request_sdkv1 "github.com/aws/aws-sdk-go/aws/request"
	lightsail_sdkv1 "github.com/aws/aws-sdk-go/service/lightsail"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
)

// CustomizeConn customizes a new AWS SDK for Go v1 client for this service package's AWS API.
func (p *servicePackage) CustomizeConn(ctx context.Context, conn *lightsail_sdkv1.Lightsail) (*lightsail_sdkv1.Lightsail, error) {
	conn.Handlers.Retry.PushBack(func(r *request_sdkv1.Request) {
		switch r.Operation.Name {
		case "CreateContainerService", "UpdateContainerService", "CreateContainerServiceDeployment":
			if tfawserr.ErrMessageContains(r.Error, lightsail_sdkv1.ErrCodeInvalidInputException, "Please try again in a few minutes") {
				r.Retryable = aws_sdkv1.Bool(true)
			}
		case "DeleteContainerService":
			if tfawserr.ErrMessageContains(r.Error, lightsail_sdkv1.ErrCodeInvalidInputException, "Please try again in a few minutes") ||
				tfawserr.ErrMessageContains(r.Error, lightsail_sdkv1.ErrCodeInvalidInputException, "Please wait for it to complete before trying again") {
				r.Retryable = aws_sdkv1.Bool(true)
			}
		}
	})

	return conn, nil
}
