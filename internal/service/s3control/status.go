package s3control

import (
	"fmt"
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3control"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// RequestStatus SUCCEEDED
	RequestStatusSucceeded = "SUCCEEDED"

	// RequestStatus FAILED
	RequestStatusFailed = "FAILED"
)

// statusPublicAccessBlockConfigurationBlockPublicACLs fetches the PublicAccessBlockConfiguration and its BlockPublicAcls
func statusPublicAccessBlockConfigurationBlockPublicACLs(conn *s3control.S3Control, accountID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		publicAccessBlockConfiguration, err := findPublicAccessBlockConfiguration(conn, accountID)

		if err != nil {
			return nil, "false", err
		}

		if publicAccessBlockConfiguration == nil {
			return nil, "false", nil
		}

		return publicAccessBlockConfiguration, strconv.FormatBool(aws.BoolValue(publicAccessBlockConfiguration.BlockPublicAcls)), nil
	}
}

// statusPublicAccessBlockConfigurationBlockPublicPolicy fetches the PublicAccessBlockConfiguration and its BlockPublicPolicy
func statusPublicAccessBlockConfigurationBlockPublicPolicy(conn *s3control.S3Control, accountID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		publicAccessBlockConfiguration, err := findPublicAccessBlockConfiguration(conn, accountID)

		if err != nil {
			return nil, "false", err
		}

		if publicAccessBlockConfiguration == nil {
			return nil, "false", nil
		}

		return publicAccessBlockConfiguration, strconv.FormatBool(aws.BoolValue(publicAccessBlockConfiguration.BlockPublicPolicy)), nil
	}
}

// statusPublicAccessBlockConfigurationIgnorePublicACLs fetches the PublicAccessBlockConfiguration and its IgnorePublicAcls
func statusPublicAccessBlockConfigurationIgnorePublicACLs(conn *s3control.S3Control, accountID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		publicAccessBlockConfiguration, err := findPublicAccessBlockConfiguration(conn, accountID)

		if err != nil {
			return nil, "false", err
		}

		if publicAccessBlockConfiguration == nil {
			return nil, "false", nil
		}

		return publicAccessBlockConfiguration, strconv.FormatBool(aws.BoolValue(publicAccessBlockConfiguration.IgnorePublicAcls)), nil
	}
}

// statusPublicAccessBlockConfigurationRestrictPublicBuckets fetches the PublicAccessBlockConfiguration and its RestrictPublicBuckets
func statusPublicAccessBlockConfigurationRestrictPublicBuckets(conn *s3control.S3Control, accountID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		publicAccessBlockConfiguration, err := findPublicAccessBlockConfiguration(conn, accountID)

		if err != nil {
			return nil, "false", err
		}

		if publicAccessBlockConfiguration == nil {
			return nil, "false", nil
		}

		return publicAccessBlockConfiguration, strconv.FormatBool(aws.BoolValue(publicAccessBlockConfiguration.RestrictPublicBuckets)), nil
	}
}

func statusMultiRegionAccessPointRequest(conn *s3control.S3Control, accountId string, requestTokenArn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &s3control.DescribeMultiRegionAccessPointOperationInput{
			AccountId:       aws.String(accountId),
			RequestTokenARN: aws.String(requestTokenArn),
		}

		log.Printf("[DEBUG] Describing S3 Multi-Region Access Point Operation (%s): %s", requestTokenArn, input)

		output, err := conn.DescribeMultiRegionAccessPointOperation(input)

		if err != nil {
			log.Printf("error Describing S3 Multi-Region Access Point Operation (%s): %s", requestTokenArn, err)
			return nil, "", err
		}

		asyncOperation := output.AsyncOperation

		if aws.StringValue(asyncOperation.RequestStatus) == RequestStatusFailed {
			errorDetails := asyncOperation.ResponseDetails.ErrorDetails
			return nil, RequestStatusFailed, fmt.Errorf("S3 Multi-Region Access Point asynchronous operation failed (%s): %s: %s", requestTokenArn, aws.StringValue(errorDetails.Code), aws.StringValue(errorDetails.Message))
		}

		return asyncOperation, aws.StringValue(asyncOperation.RequestStatus), nil
	}
}
