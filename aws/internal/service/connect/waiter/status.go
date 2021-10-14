package waiter

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	tfconnect "github.com/hashicorp/terraform-provider-aws/aws/internal/service/connect"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func InstanceStatus(ctx context.Context, conn *connect.Connect, instanceId string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &connect.DescribeInstanceInput{
			InstanceId: aws.String(instanceId),
		}

		output, err := conn.DescribeInstanceWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, tfconnect.InstanceStatusStatusNotFound) {
			return output, tfconnect.InstanceStatusStatusNotFound, nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Instance.InstanceStatus), nil
	}
}
