package events

import (
	"time"

	"github.com/aws/aws-sdk-go/service/eventbridge"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	connectionCreatedTimeout = 2 * time.Minute
	connectionDeletedTimeout = 2 * time.Minute
	connectionUpdatedTimeout = 2 * time.Minute
)

func waitConnectionCreated(conn *eventbridge.EventBridge, id string) (*eventbridge.DescribeConnectionOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{eventbridge.ConnectionStateCreating, eventbridge.ConnectionStateAuthorizing},
		Target:  []string{eventbridge.ConnectionStateAuthorized, eventbridge.ConnectionStateDeauthorized},
		Refresh: statusConnectionState(conn, id),
		Timeout: connectionCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*eventbridge.DescribeConnectionOutput); ok {
		return v, err
	}

	return nil, err
}

func waitConnectionDeleted(conn *eventbridge.EventBridge, id string) (*eventbridge.DescribeConnectionOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{eventbridge.ConnectionStateDeleting},
		Target:  []string{},
		Refresh: statusConnectionState(conn, id),
		Timeout: connectionDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*eventbridge.DescribeConnectionOutput); ok {
		return v, err
	}

	return nil, err
}

func waitConnectionUpdated(conn *eventbridge.EventBridge, id string) (*eventbridge.DescribeConnectionOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{eventbridge.ConnectionStateUpdating, eventbridge.ConnectionStateAuthorizing, eventbridge.ConnectionStateDeauthorizing},
		Target:  []string{eventbridge.ConnectionStateAuthorized, eventbridge.ConnectionStateDeauthorized},
		Refresh: statusConnectionState(conn, id),
		Timeout: connectionUpdatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*eventbridge.DescribeConnectionOutput); ok {
		return v, err
	}

	return nil, err
}
