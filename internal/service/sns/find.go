package sns

import (
	"context"
	"errors"

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

// FindTopicAttributesByARN returns topic attributes, ensuring that any Policy field is populated with
// valid principals, i.e. the principal is either an AWS Account ID or an ARN
func FindTopicAttributesByARN(ctx context.Context, conn *sns.SNS, arn string) (map[string]string, error) {
	var attributes map[string]string
	err := tfresource.Retry(ctx, propagationTimeout, func() *resource.RetryError {
		var err error
		attributes, err = GetTopicAttributesByARN(ctx, conn, arn)
		if err != nil {
			return resource.NonRetryableError(err)
		}

		valid, err := policyHasValidAWSPrincipals(attributes[TopicAttributeNamePolicy])
		if err != nil {
			return resource.NonRetryableError(err)
		}
		if !valid {
			return resource.RetryableError(errors.New("contains invalid principals"))
		}

		return nil
	})

	return attributes, err
}

// GetTopicAttributesByARN returns topic attributes without any validation. Any principals in a Policy field
// may contain Unique IDs instead of valid values. To ensure policies are valid, use FindTopicAttributesByARN
func GetTopicAttributesByARN(ctx context.Context, conn *sns.SNS, arn string) (map[string]string, error) {
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
