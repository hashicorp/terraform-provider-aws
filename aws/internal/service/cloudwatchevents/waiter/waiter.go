package waiter

import (
	"time"

	events "github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	connectionCreatedTimeout = 2 * time.Minute
	connectionDeletedTimeout = 2 * time.Minute
	connectionUpdatedTimeout = 2 * time.Minute
)

func waitConnectionCreated(conn *events.CloudWatchEvents, id string) (*events.DescribeConnectionOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{events.ConnectionStateCreating, events.ConnectionStateAuthorizing},
		Target:  []string{events.ConnectionStateAuthorized, events.ConnectionStateDeauthorized},
		Refresh: statusConnectionState(conn, id),
		Timeout: connectionCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*events.DescribeConnectionOutput); ok {
		return v, err
	}

	return nil, err
}

func waitConnectionDeleted(conn *events.CloudWatchEvents, id string) (*events.DescribeConnectionOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{events.ConnectionStateDeleting},
		Target:  []string{},
		Refresh: statusConnectionState(conn, id),
		Timeout: connectionDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*events.DescribeConnectionOutput); ok {
		return v, err
	}

	return nil, err
}

func waitConnectionUpdated(conn *events.CloudWatchEvents, id string) (*events.DescribeConnectionOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{events.ConnectionStateUpdating, events.ConnectionStateAuthorizing, events.ConnectionStateDeauthorizing},
		Target:  []string{events.ConnectionStateAuthorized, events.ConnectionStateDeauthorized},
		Refresh: statusConnectionState(conn, id),
		Timeout: connectionUpdatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*events.DescribeConnectionOutput); ok {
		return v, err
	}

	return nil, err
}
