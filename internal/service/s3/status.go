package s3

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func bucketVersioningStatus(ctx context.Context, conn *s3.S3, bucket, expectedBucketOwner string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &s3.GetBucketVersioningInput{
			Bucket: aws.String(bucket),
		}

		if expectedBucketOwner != "" {
			input.ExpectedBucketOwner = aws.String(expectedBucketOwner)
		}

		output, err := conn.GetBucketVersioningWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, s3.ErrCodeNoSuchBucket) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if output == nil {
			return nil, "", &resource.NotFoundError{
				Message:     "Empty result",
				LastRequest: input,
			}
		}

		return output, aws.StringValue(output.Status), nil
	}
}
