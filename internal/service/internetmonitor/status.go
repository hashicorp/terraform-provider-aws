package internetmonitor

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/internetmonitor"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func statusMonitor(ctx context.Context, conn *internetmonitor.InternetMonitor, name string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		monitor, err := FindMonitor(ctx, conn, name)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return monitor, aws.StringValue(monitor.Status), nil
	}
}
