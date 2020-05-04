package waiter

import (
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

		return output, aws.StringValue(output.Canary.Status.State), nil
	}
}
