// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appstream

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/appstream"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	// fleetStateTimeout Maximum amount of time to wait for the statusFleetState to be RUNNING or STOPPED
	fleetStateTimeout = 180 * time.Minute
	// fleetOperationTimeout Maximum amount of time to wait for Fleet operation eventual consistency
	fleetOperationTimeout = 15 * time.Minute
	// imageBuilderStateTimeout Maximum amount of time to wait for the statusImageBuilderState to be RUNNING
	// or for the ImageBuilder to be deleted
	imageBuilderStateTimeout = 60 * time.Minute
	// userOperationTimeout Maximum amount of time to wait for User operation eventual consistency
	userOperationTimeout = 4 * time.Minute
	// iamPropagationTimeout Maximum amount of time to wait for an iam resource eventual consistency
	iamPropagationTimeout = 2 * time.Minute
	userAvailable         = "AVAILABLE"
)

// waitFleetStateRunning waits for a fleet running
func waitFleetStateRunning(ctx context.Context, conn *appstream.AppStream, name string) (*appstream.Fleet, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: []string{appstream.FleetStateStarting},
		Target:  []string{appstream.FleetStateRunning},
		Refresh: statusFleetState(ctx, conn, name),
		Timeout: fleetStateTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*appstream.Fleet); ok {
		if errors := output.FleetErrors; len(errors) > 0 {
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

// waitFleetStateStopped waits for a fleet stopped
func waitFleetStateStopped(ctx context.Context, conn *appstream.AppStream, name string) (*appstream.Fleet, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: []string{appstream.FleetStateStopping},
		Target:  []string{appstream.FleetStateStopped},
		Refresh: statusFleetState(ctx, conn, name),
		Timeout: fleetStateTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*appstream.Fleet); ok {
		if errors := output.FleetErrors; len(errors) > 0 {
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

func waitImageBuilderStateRunning(ctx context.Context, conn *appstream.AppStream, name string) (*appstream.ImageBuilder, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{appstream.ImageBuilderStatePending},
		Target:  []string{appstream.ImageBuilderStateRunning},
		Refresh: statusImageBuilderState(ctx, conn, name),
		Timeout: imageBuilderStateTimeout,
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

func waitImageBuilderStateDeleted(ctx context.Context, conn *appstream.AppStream, name string) (*appstream.ImageBuilder, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{appstream.ImageBuilderStatePending, appstream.ImageBuilderStateDeleting},
		Target:  []string{},
		Refresh: statusImageBuilderState(ctx, conn, name),
		Timeout: imageBuilderStateTimeout,
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

// waitUserAvailable waits for a user be available
func waitUserAvailable(ctx context.Context, conn *appstream.AppStream, username, authType string) (*appstream.User, error) {
	stateConf := &retry.StateChangeConf{
		Target:  []string{userAvailable},
		Refresh: statusUserAvailable(ctx, conn, username, authType),
		Timeout: userOperationTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*appstream.User); ok {
		return output, err
	}

	return nil, err
}
