package waiter

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/appstream"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// StackOperationTimeout Maximum amount of time to wait for Stack operation eventual consistency
	StackOperationTimeout = 4 * time.Minute

	// FleetStateTimeout Maximum amount of time to wait for the FleetState to be RUNNING or STOPPED
	FleetStateTimeout = 30 * time.Minute
	// FleetOperationTimeout Maximum amount of time to wait for Fleet operation eventual consistency
	FleetOperationTimeout = 4 * time.Minute
)

// StackStateDeleted waits for a deleted stack
func StackStateDeleted(ctx context.Context, conn *appstream.AppStream, name string) (*appstream.Stack, error) {
	stateConf := &resource.StateChangeConf{
		Target:  []string{"NotFound", "Unknown"},
		Refresh: StackState(ctx, conn, name),
		Timeout: StackOperationTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*appstream.Stack); ok {
		return output, err
	}

	return nil, err
}

// FleetStateRunning waits for a fleet running
func FleetStateRunning(ctx context.Context, conn *appstream.AppStream, name string) (*appstream.Fleet, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{appstream.FleetStateStarting},
		Target:  []string{appstream.FleetStateRunning},
		Refresh: FleetState(conn, name),
		Timeout: FleetStateTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*appstream.Fleet); ok {
		return output, err
	}

	return nil, err
}

// FleetStateStopped waits for a fleet stopped
func FleetStateStopped(ctx context.Context, conn *appstream.AppStream, name string) (*appstream.Fleet, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{appstream.FleetStateStopping},
		Target:  []string{appstream.FleetStateStopped},
		Refresh: FleetState(conn, name),
		Timeout: FleetStateTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*appstream.Fleet); ok {
		return output, err
	}

	return nil, err
}
