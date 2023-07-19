// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package workspaces

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/workspaces"
	"github.com/aws/aws-sdk-go-v2/service/workspaces/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
)

const (
	DirectoryDeregisterInvalidResourceStateTimeout = 2 * time.Minute
	DirectoryRegisterInvalidResourceStateTimeout   = 2 * time.Minute

	// Maximum amount of time to wait for a Directory to return Registered
	DirectoryRegisteredTimeout = 10 * time.Minute

	// Maximum amount of time to wait for a Directory to return Deregistered
	DirectoryDeregisteredTimeout = 10 * time.Minute

	// Maximum amount of time to wait for a WorkSpace to return Available
	WorkspaceAvailableTimeout = 30 * time.Minute

	// Maximum amount of time to wait for a WorkSpace while returning Updating
	WorkspaceUpdatingTimeout = 10 * time.Minute

	// Amount of time to delay before checking WorkSpace when updating
	WorkspaceUpdatingDelay = 1 * time.Minute

	// Maximum amount of time to wait for a WorkSpace to return Terminated
	WorkspaceTerminatedTimeout = 10 * time.Minute
)

func WaitDirectoryRegistered(ctx context.Context, conn *workspaces.Client, directoryID string) (*types.WorkspaceDirectory, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.WorkspaceDirectoryStateRegistering),
		Target:  enum.Slice(types.WorkspaceDirectoryStateRegistered),
		Refresh: StatusDirectoryState(ctx, conn, directoryID),
		Timeout: DirectoryRegisteredTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*types.WorkspaceDirectory); ok {
		return v, err
	}

	return nil, err
}

func WaitDirectoryDeregistered(ctx context.Context, conn *workspaces.Client, directoryID string) (*types.WorkspaceDirectory, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			types.WorkspaceDirectoryStateRegistering,
			types.WorkspaceDirectoryStateRegistered,
			types.WorkspaceDirectoryStateDeregistering,
		),
		Target:  []string{},
		Refresh: StatusDirectoryState(ctx, conn, directoryID),
		Timeout: DirectoryDeregisteredTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*types.WorkspaceDirectory); ok {
		return v, err
	}

	return nil, err
}

func WaitWorkspaceAvailable(ctx context.Context, conn *workspaces.Client, workspaceID string, timeout time.Duration) (*types.Workspace, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			types.WorkspaceStatePending,
			types.WorkspaceStateStarting,
		),
		Target:  enum.Slice(types.WorkspaceStateAvailable),
		Refresh: StatusWorkspaceState(ctx, conn, workspaceID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*types.Workspace); ok {
		return v, err
	}

	return nil, err
}

func WaitWorkspaceTerminated(ctx context.Context, conn *workspaces.Client, workspaceID string, timeout time.Duration) (*types.Workspace, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			types.WorkspaceStatePending,
			types.WorkspaceStateAvailable,
			types.WorkspaceStateImpaired,
			types.WorkspaceStateUnhealthy,
			types.WorkspaceStateRebooting,
			types.WorkspaceStateStarting,
			types.WorkspaceStateRebuilding,
			types.WorkspaceStateRestoring,
			types.WorkspaceStateMaintenance,
			types.WorkspaceStateAdminMaintenance,
			types.WorkspaceStateSuspended,
			types.WorkspaceStateUpdating,
			types.WorkspaceStateStopping,
			types.WorkspaceStateStopped,
			types.WorkspaceStateTerminating,
			types.WorkspaceStateError,
		),
		Target:  enum.Slice(types.WorkspaceStateTerminated),
		Refresh: StatusWorkspaceState(ctx, conn, workspaceID),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*types.Workspace); ok {
		return v, err
	}

	return nil, err
}

func WaitWorkspaceUpdated(ctx context.Context, conn *workspaces.Client, workspaceID string, timeout time.Duration) (*types.Workspace, error) {
	// OperationInProgressException: The properties of this WorkSpace are currently under modification. Please try again in a moment.
	// AWS Workspaces service doesn't change instance status to "Updating" during property modification. Respective AWS Support feature request has been created. Meanwhile, artificial delay is placed here as a workaround.
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			types.WorkspaceStateUpdating,
		),
		Target: enum.Slice(
			types.WorkspaceStateAvailable,
			types.WorkspaceStateStopped,
		),
		Refresh: StatusWorkspaceState(ctx, conn, workspaceID),
		Delay:   WorkspaceUpdatingDelay,
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*types.Workspace); ok {
		return v, err
	}

	return nil, err
}
