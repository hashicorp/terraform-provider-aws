package sns

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sns"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindSubscriptionByARN(conn *sns.SNS, id string) (*sns.GetSubscriptionAttributesOutput, error) {
	output, err := conn.GetSubscriptionAttributes(&sns.GetSubscriptionAttributesInput{
		SubscriptionArn: aws.String(id),
	})
	if tfawserr.ErrCodeEquals(err, sns.ErrCodeNotFoundException) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Attributes == nil || len(output.Attributes) == 0 {
		return nil, nil
	}

	return output, nil
}

func FindTopicAttributesByARN(conn *sns.SNS, arn string) (map[string]string, error) {
	input := &sns.GetTopicAttributesInput{
		TopicArn: aws.String(arn),
	}

	output, err := conn.GetTopicAttributes(input)

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
