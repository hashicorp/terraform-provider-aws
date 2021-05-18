package waiter

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apprunner"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func AutoScalingConfigurationStatus(ctx context.Context, conn *apprunner.AppRunner, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &apprunner.DescribeAutoScalingConfigurationInput{
			AutoScalingConfigurationArn: aws.String(arn),
		}

		output, err := conn.DescribeAutoScalingConfigurationWithContext(ctx, input)

		if err != nil {
			return nil, "", err
		}

		if output == nil || output.AutoScalingConfiguration == nil {
			return nil, "", nil
		}

		return output.AutoScalingConfiguration, aws.StringValue(output.AutoScalingConfiguration.Status), nil
	}
}

func ConnectionStatus(ctx context.Context, conn *apprunner.AppRunner, name string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		c, err := finder.ConnectionSummaryByName(ctx, conn, name)

		if err != nil {
			return nil, "", err
		}

		if c == nil {
			return nil, "", nil
		}

		return c, aws.StringValue(c.Status), nil
	}
}
