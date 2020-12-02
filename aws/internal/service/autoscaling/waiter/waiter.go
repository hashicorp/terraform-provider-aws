package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/autoscaling"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// Maximum amount of time to wait for an InstanceRefresh to be Successful
	InstanceRefreshSuccessfulTimeout = 5 * time.Minute

	// Maximum amount of time to wait for an InstanceRefresh to be Cancelled
	InstanceRefreshCancelledTimeout = 10 * time.Minute
)

func InstanceRefreshSuccessful(conn *autoscaling.AutoScaling, asgName, instanceRefreshId string) (*autoscaling.InstanceRefresh, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{autoscaling.InstanceRefreshStatusPending, autoscaling.InstanceRefreshStatusInProgress},
		Target:  []string{autoscaling.InstanceRefreshStatusSuccessful},
		Refresh: InstanceRefreshStatus(conn, asgName, instanceRefreshId),
		Timeout: InstanceRefreshSuccessfulTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*autoscaling.InstanceRefresh); ok {
		return v, err
	}

	return nil, err
}

func InstanceRefreshCancelled(conn *autoscaling.AutoScaling, asgName, instanceRefreshId string) (*autoscaling.InstanceRefresh, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{autoscaling.InstanceRefreshStatusPending, autoscaling.InstanceRefreshStatusInProgress, autoscaling.InstanceRefreshStatusCancelling},
		Target:  []string{autoscaling.InstanceRefreshStatusCancelled},
		Refresh: InstanceRefreshStatus(conn, asgName, instanceRefreshId),
		Timeout: InstanceRefreshCancelledTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*autoscaling.InstanceRefresh); ok {
		return v, err
	}

	return nil, err
}
