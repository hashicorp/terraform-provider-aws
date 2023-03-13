package kinesisanalytics

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesisanalytics"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

// FindApplicationDetailByName returns the application corresponding to the specified name.
// Returns NotFoundError if no application is found.
func FindApplicationDetailByName(ctx context.Context, conn *kinesisanalytics.KinesisAnalytics, name string) (*kinesisanalytics.ApplicationDetail, error) {
	input := &kinesisanalytics.DescribeApplicationInput{
		ApplicationName: aws.String(name),
	}

	return FindApplicationDetail(ctx, conn, input)
}

// FindApplicationDetail returns the application details corresponding to the specified name.
// Returns NotFoundError if no application is found.
func FindApplicationDetail(ctx context.Context, conn *kinesisanalytics.KinesisAnalytics, input *kinesisanalytics.DescribeApplicationInput) (*kinesisanalytics.ApplicationDetail, error) {
	output, err := conn.DescribeApplicationWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, kinesisanalytics.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ApplicationDetail == nil {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return output.ApplicationDetail, nil
}
