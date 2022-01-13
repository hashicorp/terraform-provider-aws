package s3

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func lifecycleConfigurationRulesStatus(ctx context.Context, conn *s3.S3, bucket string, rules []*s3.LifecycleRule) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &s3.GetBucketLifecycleConfigurationInput{
			Bucket: aws.String(bucket),
		}

		output, err := conn.GetBucketLifecycleConfigurationWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, ErrCodeNoSuchLifecycleConfiguration) {
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

		for _, expectedRule := range rules {
			found := false

			for _, actualRule := range output.Rules {
				if aws.StringValue(actualRule.ID) != aws.StringValue(expectedRule.ID) {
					continue
				}
				found = true
				if aws.StringValue(actualRule.Status) != aws.StringValue(expectedRule.Status) {
					return output, LifecycleConfigurationRulesStatusNotReady, nil
				}
			}

			if !found {
				return output, LifecycleConfigurationRulesStatusNotReady, nil
			}
		}

		return output, LifecycleConfigurationRulesStatusReady, nil
	}
}
