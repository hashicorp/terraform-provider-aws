package autoscaling

import (
	"time"

	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	// Maximum amount of time to wait for an InstanceRefresh to be started
	// Must be at least as long as instanceRefreshCancelledTimeout, since we try to cancel any
	// existing Instance Refreshes when starting.
	instanceRefreshStartedTimeout = instanceRefreshCancelledTimeout

	// Maximum amount of time to wait for an Instance Refresh to be Cancelled
	instanceRefreshCancelledTimeout = 15 * time.Minute
)

func waitInstanceRefreshCancelled(conn *autoscaling.AutoScaling, asgName, instanceRefreshId string) (*autoscaling.InstanceRefresh, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			autoscaling.InstanceRefreshStatusPending,
			autoscaling.InstanceRefreshStatusInProgress,
			autoscaling.InstanceRefreshStatusCancelling,
		},
		Target: []string{
			autoscaling.InstanceRefreshStatusCancelled,
			// Failed and Successful are also acceptable end-states
			autoscaling.InstanceRefreshStatusFailed,
			autoscaling.InstanceRefreshStatusSuccessful,
		},
		Refresh: statusInstanceRefresh(conn, asgName, instanceRefreshId),
		Timeout: instanceRefreshCancelledTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*autoscaling.InstanceRefresh); ok {
		return v, err
	}

	return nil, err
}
