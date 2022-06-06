package resiliencehub

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/resiliencehub"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func FindResiliencyPolicyByArn(conn *resiliencehub.ResilienceHub, arn string) (*resiliencehub.ResiliencyPolicy, error) {
	input := &resiliencehub.DescribeResiliencyPolicyInput{
		PolicyArn: aws.String(arn),
	}

	output, err := conn.DescribeResiliencyPolicy(input)

	if tfawserr.ErrCodeContains(err, resiliencehub.ErrCodeResourceNotFoundException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Policy == nil {
		return nil, &resource.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return output.Policy, nil
}
