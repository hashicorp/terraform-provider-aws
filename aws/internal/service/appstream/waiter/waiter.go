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
	FleetStateTimeout = 180 * time.Minute
	// FleetOperationTimeout Maximum amount of time to wait for Fleet operation eventual consistency
	FleetOperationTimeout = 15 * time.Minute
	// ImageBuilderStateTimeout Maximum amount of time to wait for the ImageBuilderState to be RUNNING or STOPPED
	ImageBuilderStateTimeout = 60 * time.Minute
	// ImageBuilderOperationTimeout Maximum amount of time to wait for ImageBuilder operation eventual consistency
	ImageBuilderOperationTimeout = 4 * time.Minute
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
		Refresh: FleetState(ctx, conn, name),
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
		Refresh: FleetState(ctx, conn, name),
		Timeout: FleetStateTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*appstream.Fleet); ok {
		return output, err
	}

	return nil, err
}

// ImageBuilderStateRunning waits for a ImageBuilder running
func ImageBuilderStateRunning(ctx context.Context, conn *appstream.AppStream, name string) (*appstream.ImageBuilder, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{appstream.ImageBuilderStatePending},
		Target:  []string{appstream.ImageBuilderStateRunning},
		Refresh: ImageBuilderState(conn, name),
		Timeout: ImageBuilderStateTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*appstream.ImageBuilder); ok {
		return output, err
	}

	return nil, err
}

// ImageBuilderStateStopped waits for a ImageBuilder stopped
func ImageBuilderStateStopped(ctx context.Context, conn *appstream.AppStream, name string) (*appstream.ImageBuilder, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{appstream.ImageBuilderStatePending, appstream.ImageBuilderStateStopping},
		Target:  []string{appstream.ImageBuilderStateStopped},
		Refresh: ImageBuilderState(conn, name),
		Timeout: ImageBuilderStateTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*appstream.ImageBuilder); ok {
		return output, err
	}

	return nil, err
}

// ImageBuilderStateDeleted waits for a ImageBuilder deleted
func ImageBuilderStateDeleted(ctx context.Context, conn *appstream.AppStream, name string) (*appstream.ImageBuilder, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{appstream.ImageBuilderStatePending, appstream.ImageBuilderStateDeleting},
		Target:  []string{"NotFound"},
		Refresh: ImageBuilderState(conn, name),
		Timeout: ImageBuilderStateTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*appstream.ImageBuilder); ok {
		return output, err
	}

	return nil, err
}
