package waiter

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func MetricStreamDeleted(ctx context.Context, conn *cloudwatch.CloudWatch, name string) (*cloudwatch.GetMetricStreamOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			"running",
			"stopped",
		},
		Target:  []string{},
		Refresh: MetricStreamState(ctx, conn, name),
		Timeout: 10 * time.Minute,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*cloudwatch.GetMetricStreamOutput); ok {
		return v, err
	}

	return nil, err
}
