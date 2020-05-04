package waiter

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/synthetics"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func CanaryStatus(conn *synthetics.Synthetics, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &synthetics.GetCanaryInput{
			Name: aws.String(name),
		}

		output, err := conn.GetCanary(input)

		if err != nil {
			return nil, synthetics.CanaryStateError, err
		}

		if aws.StringValue(output.Canary.Status.State) == synthetics.CanaryStateError {
			return output, synthetics.CanaryStateError, fmt.Errorf("%s: %s", aws.StringValue(output.Canary.Status.StateReasonCode), aws.StringValue(output.Canary.Status.StateReason))
		}

		return output, aws.StringValue(output.Canary.Status.State), nil
	}
}
