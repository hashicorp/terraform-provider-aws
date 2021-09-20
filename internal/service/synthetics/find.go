package synthetics

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/synthetics"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func FindCanaryByName(conn *synthetics.Synthetics, name string) (*synthetics.Canary, error) {
	input := &synthetics.GetCanaryInput{
		Name: aws.String(name),
	}

	output, err := conn.GetCanary(input)

	if tfawserr.ErrCodeEquals(err, synthetics.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Canary == nil || output.Canary.Status == nil {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return output.Canary, nil
}
