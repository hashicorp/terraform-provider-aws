package internetmonitor

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/internetmonitor"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	monitorCreatedTimeout = 5 * time.Minute
)

func waitMonitor(ctx context.Context, conn *internetmonitor.InternetMonitor, name, target string) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{internetmonitor.MonitorConfigStatePending},
		Target:  []string{target},
		Refresh: statusMonitor(ctx, conn, name),
		Timeout: monitorCreatedTimeout,
		Delay:   10 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*internetmonitor.GetMonitorOutput); ok {
		if statusCode := aws.StringValue(output.Status); statusCode == internetmonitor.MonitorConfigStateError {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.ProcessingStatusInfo)))
		}

		return err
	}

	return err
}
