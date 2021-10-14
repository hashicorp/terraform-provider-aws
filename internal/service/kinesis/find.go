package kinesis

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesis"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

// FindStreamConsumerByARN returns the stream consumer corresponding to the specified ARN.
// Returns nil if no stream consumer is found.
func FindStreamConsumerByARN(conn *kinesis.Kinesis, arn string) (*kinesis.ConsumerDescription, error) {
	input := &kinesis.DescribeStreamConsumerInput{
		ConsumerARN: aws.String(arn),
	}

	output, err := conn.DescribeStreamConsumer(input)
	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	return output.ConsumerDescription, nil
}
