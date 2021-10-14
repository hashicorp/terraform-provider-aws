package firehose

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/firehose"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func FindDeliveryStreamByName(conn *firehose.Firehose, name string) (*firehose.DeliveryStreamDescription, error) {
	input := &firehose.DescribeDeliveryStreamInput{
		DeliveryStreamName: aws.String(name),
	}

	output, err := conn.DescribeDeliveryStream(input)

	if tfawserr.ErrCodeEquals(err, firehose.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.DeliveryStreamDescription == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.DeliveryStreamDescription, nil
}

func FindDeliveryStreamEncryptionConfigurationByName(conn *firehose.Firehose, name string) (*firehose.DeliveryStreamEncryptionConfiguration, error) {
	output, err := FindDeliveryStreamByName(conn, name)

	if err != nil {
		return nil, err
	}

	if output.DeliveryStreamEncryptionConfiguration == nil {
		return nil, tfresource.NewEmptyResultError(nil)
	}

	return output.DeliveryStreamEncryptionConfiguration, nil
}
