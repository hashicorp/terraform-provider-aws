package events

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/eventbridge"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	connectionCreatedTimeout = 2 * time.Minute
	connectionDeletedTimeout = 2 * time.Minute
	connectionUpdatedTimeout = 2 * time.Minute
)

func waitConnectionCreated(ctx context.Context, conn *eventbridge.EventBridge, id string) (*eventbridge.DescribeConnectionOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{eventbridge.ConnectionStateCreating, eventbridge.ConnectionStateAuthorizing},
		Target:  []string{eventbridge.ConnectionStateAuthorized, eventbridge.ConnectionStateDeauthorized},
		Refresh: statusConnectionState(ctx, conn, id),
		Timeout: connectionCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*eventbridge.DescribeConnectionOutput); ok {
		return v, err
	}

	return nil, err
}

func waitConnectionDeleted(ctx context.Context, conn *eventbridge.EventBridge, id string) (*eventbridge.DescribeConnectionOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{eventbridge.ConnectionStateDeleting},
		Target:  []string{},
		Refresh: statusConnectionState(ctx, conn, id),
		Timeout: connectionDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*eventbridge.DescribeConnectionOutput); ok {
		return v, err
	}

	return nil, err
}

func waitConnectionUpdated(ctx context.Context, conn *eventbridge.EventBridge, id string) (*eventbridge.DescribeConnectionOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{eventbridge.ConnectionStateUpdating, eventbridge.ConnectionStateAuthorizing, eventbridge.ConnectionStateDeauthorizing},
		Target:  []string{eventbridge.ConnectionStateAuthorized, eventbridge.ConnectionStateDeauthorized},
		Refresh: statusConnectionState(ctx, conn, id),
		Timeout: connectionUpdatedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*eventbridge.DescribeConnectionOutput); ok {
		return v, err
	}

	return nil, err
}
