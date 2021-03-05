package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mq"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func BrokerStatus(conn *mq.MQ, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &mq.DescribeBrokerInput{
			BrokerId: aws.String(id),
		}

		output, err := conn.DescribeBroker(input)

		if err != nil {
			return nil, aws.StringValue(output.BrokerState), err
		}

		if output == nil {
			return output, aws.StringValue(output.BrokerState), nil
		}

		return output, aws.StringValue(output.BrokerState), nil
	}
}
