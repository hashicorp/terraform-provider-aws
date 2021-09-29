package waiter

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appstream"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

const (
	// StackOperationTimeout Maximum amount of time to wait for Stack operation eventual consistency
	StackOperationTimeout = 4 * time.Minute

	// FleetStateTimeout Maximum amount of time to wait for the FleetState to be RUNNING or STOPPED
	FleetStateTimeout = 180 * time.Minute
	// FleetOperationTimeout Maximum amount of time to wait for Fleet operation eventual consistency
	FleetOperationTimeout = 15 * time.Minute
	// ImageBuilderStateTimeout Maximum amount of time to wait for the ImageBuilderState to be RUNNING
	// or for the ImageBuilder to be deleted
	ImageBuilderStateTimeout = 60 * time.Minute
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
		Refresh: ImageBuilderState(ctx, conn, name),
		Timeout: ImageBuilderStateTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*appstream.ImageBuilder); ok {
		if state, errors := aws.StringValue(output.State), output.ImageBuilderErrors; state == appstream.ImageBuilderStateFailed && len(errors) > 0 {
			var errs *multierror.Error

			for _, err := range errors {
				errs = multierror.Append(errs, fmt.Errorf("%s: %s", aws.StringValue(err.ErrorCode), aws.StringValue(err.ErrorMessage)))
			}

			tfresource.SetLastError(err, errs.ErrorOrNil())
		}

		return output, err
	}

	return nil, err
}

// ImageBuilderStateDeleted waits for a ImageBuilder deleted
func ImageBuilderStateDeleted(ctx context.Context, conn *appstream.AppStream, name string) (*appstream.ImageBuilder, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{appstream.ImageBuilderStatePending, appstream.ImageBuilderStateDeleting},
		Target:  []string{},
		Refresh: ImageBuilderState(ctx, conn, name),
		Timeout: ImageBuilderStateTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*appstream.ImageBuilder); ok {
		if state, errors := aws.StringValue(output.State), output.ImageBuilderErrors; state == appstream.ImageBuilderStateFailed && len(errors) > 0 {
			var errs *multierror.Error

			for _, err := range errors {
				errs = multierror.Append(errs, fmt.Errorf("%s: %s", aws.StringValue(err.ErrorCode), aws.StringValue(err.ErrorMessage)))
			}

			tfresource.SetLastError(err, errs.ErrorOrNil())
		}

		return output, err
	}

	return nil, err
}
