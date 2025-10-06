package lambda

import (
	"context"

	"github.com/aws/aws-sdk-go/aws/request"
	lambda_sdkv1 "github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
)

// Customize lambda retries.
//
// References:
//
// https://github.com/hashicorp/terraform-provider-aws/blob/main/docs/retries-and-waiters.md
func (p *servicePackage) CustomizeConn(ctx context.Context, conn *lambda_sdkv1.Lambda) (*lambda_sdkv1.Lambda, error) {
	conn.Handlers.Retry.PushBack(func(r *request.Request) {
		if tfawserr.ErrMessageContains(r.Error, lambda_sdkv1.ErrCodeKMSAccessDeniedException,
			"Lambda was unable to decrypt the environment variables because KMS access was denied.") {
			// Do not retry this condition at all.
			r.RetryCount = r.MaxRetries() + 1
		}
	})
	return conn, nil
}
