package firehose

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/firehose"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func statusDeliveryStream(conn *firehose.Firehose, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindDeliveryStreamByName(conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.DeliveryStreamStatus), nil
	}
}

func statusDeliveryStreamEncryptionConfiguration(conn *firehose.Firehose, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindDeliveryStreamEncryptionConfigurationByName(conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}
