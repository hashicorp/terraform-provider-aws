package kinesis

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	streamConsumerStatusNotFound = "NotFound"
	streamConsumerStatusUnknown  = "Unknown"
)

// statusStreamConsumer fetches the StreamConsumer and its Status
func statusStreamConsumer(conn *kinesis.Kinesis, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		consumer, err := FindStreamConsumerByARN(conn, arn)

		if err != nil {
			return nil, streamConsumerStatusUnknown, err
		}

		if consumer == nil {
			return nil, streamConsumerStatusNotFound, nil
		}

		return consumer, aws.StringValue(consumer.ConsumerStatus), nil
	}
}
