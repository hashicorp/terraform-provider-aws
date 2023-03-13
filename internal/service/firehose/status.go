package firehose

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/firehose"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func statusDeliveryStream(ctx context.Context, conn *firehose.Firehose, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindDeliveryStreamByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.DeliveryStreamStatus), nil
	}
}

func statusDeliveryStreamEncryptionConfiguration(ctx context.Context, conn *firehose.Firehose, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindDeliveryStreamEncryptionConfigurationByName(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}
