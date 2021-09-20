package kafka

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	configurationStateDeleted = "Deleted"
	configurationStateUnknown = "Unknown"
)

// statusConfigurationState fetches the Operation and its Status
func statusConfigurationState(conn *kafka.Kafka, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &kafka.DescribeConfigurationInput{
			Arn: aws.String(arn),
		}

		output, err := conn.DescribeConfiguration(input)

		if tfawserr.ErrMessageContains(err, kafka.ErrCodeBadRequestException, "Configuration ARN does not exist") {
			return output, configurationStateDeleted, nil
		}

		if err != nil {
			return output, configurationStateUnknown, err
		}

		if output == nil {
			return output, configurationStateUnknown, nil
		}

		return output, aws.StringValue(output.State), nil
	}
}
