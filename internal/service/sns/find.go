package sns

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindPlatformApplicationAttributesByARN(ctx context.Context, conn *sns.SNS, arn string) (map[string]string, error) {
	input := &sns.GetPlatformApplicationAttributesInput{
		PlatformApplicationArn: aws.String(arn),
	}

	output, err := conn.GetPlatformApplicationAttributesWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, sns.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Attributes) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return aws.StringValueMap(output.Attributes), nil
}

func FindSubscriptionAttributesByARN(ctx context.Context, conn *sns.SNS, arn string) (map[string]string, error) {
	input := &sns.GetSubscriptionAttributesInput{
		SubscriptionArn: aws.String(arn),
	}

	output, err := conn.GetSubscriptionAttributesWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, sns.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Attributes) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return aws.StringValueMap(output.Attributes), nil
}

func FindTopicAttributesByARN(ctx context.Context, conn *sns.SNS, arn string) (map[string]string, error) {
	input := &sns.GetTopicAttributesInput{
		TopicArn: aws.String(arn),
	}

	output, err := conn.GetTopicAttributesWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, sns.ErrCodeNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Attributes) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return aws.StringValueMap(output.Attributes), nil
}
