package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/mq"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	BrokerNotFoundStatus = "NotFound"
)

func BrokerStatus(conn *mq.MQ, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &mq.DescribeBrokerInput{
			BrokerId: aws.String(id),
		}

		output, err := conn.DescribeBroker(input)

		if err != nil && tfawserr.ErrCodeEquals(err, "NotFoundException") {
			return nil, BrokerNotFoundStatus, nil
		}

		if err != nil {
			return nil, aws.StringValue(output.BrokerState), err
		}

		return output, aws.StringValue(output.BrokerState), nil
	}
}
