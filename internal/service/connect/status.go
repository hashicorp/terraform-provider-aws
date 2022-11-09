package connect

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func statusInstance(ctx context.Context, conn *connect.Connect, instanceId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &connect.DescribeInstanceInput{
			InstanceId: aws.String(instanceId),
		}

		output, err := conn.DescribeInstanceWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, connect.ErrCodeResourceNotFoundException) {
			return output, connect.ErrCodeResourceNotFoundException, nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Instance.InstanceStatus), nil
	}
}

func statusPhoneNumber(ctx context.Context, conn *connect.Connect, phoneNumberId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &connect.DescribePhoneNumberInput{
			PhoneNumberId: aws.String(phoneNumberId),
		}

		output, err := conn.DescribePhoneNumberWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, connect.ErrCodeResourceNotFoundException) {
			return output, connect.ErrCodeResourceNotFoundException, nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.ClaimedPhoneNumberSummary.PhoneNumberStatus.Status), nil
	}
}

func statusVocabulary(ctx context.Context, conn *connect.Connect, instanceId, vocabularyId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &connect.DescribeVocabularyInput{
			InstanceId:   aws.String(instanceId),
			VocabularyId: aws.String(vocabularyId),
		}

		output, err := conn.DescribeVocabularyWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, connect.ErrCodeResourceNotFoundException) {
			return output, connect.ErrCodeResourceNotFoundException, nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Vocabulary.State), nil
	}
}
