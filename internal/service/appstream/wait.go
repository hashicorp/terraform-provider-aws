// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appstream

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/appstream"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appstream/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
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
func waitFleetStateRunning(ctx context.Context, conn *appstream.Client, name string) (*awstypes.Fleet, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.FleetStateStarting),
		Target:  enum.Slice(awstypes.FleetStateRunning),
		Refresh: statusFleetState(ctx, conn, name),
		Timeout: fleetStateTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Fleet); ok {
		if v := output.FleetErrors; len(v) > 0 {
			var errs []error

			for _, err := range v {
				errs = append(errs, fmt.Errorf("%s: %s", string(err.ErrorCode), aws.ToString(err.ErrorMessage)))
			}

			tfresource.SetLastError(err, errors.Join(errs...))
		}

		return output, err
	}

	return nil, err
}

// waitFleetStateStopped waits for a fleet stopped
func waitFleetStateStopped(ctx context.Context, conn *appstream.Client, name string) (*awstypes.Fleet, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.FleetStateStopping),
		Target:  enum.Slice(awstypes.FleetStateStopped),
		Refresh: statusFleetState(ctx, conn, name),
		Timeout: fleetStateTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Fleet); ok {
		if v := output.FleetErrors; len(v) > 0 {
			var errs []error

			for _, err := range v {
				errs = append(errs, fmt.Errorf("%s: %s", string(err.ErrorCode), aws.ToString(err.ErrorMessage)))
			}

			tfresource.SetLastError(err, errors.Join(errs...))
		}

		return output, err
	}

	return nil, err
}

func waitImageBuilderStateRunning(ctx context.Context, conn *appstream.Client, name string) (*awstypes.ImageBuilder, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ImageBuilderStatePending),
		Target:  enum.Slice(awstypes.ImageBuilderStateRunning),
		Refresh: statusImageBuilderState(ctx, conn, name),
		Timeout: imageBuilderStateTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ImageBuilder); ok {
		if state, v := output.State, output.ImageBuilderErrors; state == awstypes.ImageBuilderStateFailed && len(v) > 0 {
			var errs []error

			for _, err := range v {
				errs = append(errs, fmt.Errorf("%s: %s", string(err.ErrorCode), aws.ToString(err.ErrorMessage)))
			}

			tfresource.SetLastError(err, errors.Join(errs...))
		}

		return output, err
	}

	return nil, err
}

func waitImageBuilderStateDeleted(ctx context.Context, conn *appstream.Client, name string) (*awstypes.ImageBuilder, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ImageBuilderStatePending, awstypes.ImageBuilderStateDeleting),
		Target:  []string{},
		Refresh: statusImageBuilderState(ctx, conn, name),
		Timeout: imageBuilderStateTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ImageBuilder); ok {
		if state, v := output.State, output.ImageBuilderErrors; state == awstypes.ImageBuilderStateFailed && len(v) > 0 {
			var errs []error

			for _, err := range v {
				errs = append(errs, fmt.Errorf("%s: %s", string(err.ErrorCode), aws.ToString(err.ErrorMessage)))
			}

			tfresource.SetLastError(err, errors.Join(errs...))
		}

		return output, err
	}

	return nil, err
}

// waitUserAvailable waits for a user be available
func waitUserAvailable(ctx context.Context, conn *appstream.Client, username, authType string) (*awstypes.User, error) {
	stateConf := &retry.StateChangeConf{
		Target:  []string{userAvailable},
		Refresh: statusUserAvailable(ctx, conn, username, authType),
		Timeout: userOperationTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.User); ok {
		return output, err
	}

	return nil, err
}
