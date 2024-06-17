// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package efs

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/efs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
)

const (
	// Maximum amount of time to wait for an Operation to return Success
	accessPointCreatedTimeout = 10 * time.Minute
	accessPointDeletedTimeout = 10 * time.Minute

	backupPolicyDisabledTimeout = 10 * time.Minute
	backupPolicyEnabledTimeout  = 10 * time.Minute
)

// waitAccessPointCreated waits for an Operation to return Success
func waitAccessPointCreated(ctx context.Context, conn *efs.Client, accessPointId string) (*awstypes.AccessPointDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.LifeCycleStateCreating),
		Target:  enum.Slice(awstypes.LifeCycleStateAvailable),
		Refresh: statusAccessPointLifeCycleState(ctx, conn, accessPointId),
		Timeout: accessPointCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.AccessPointDescription); ok {
		return output, err
	}

	return nil, err
}

// waitAccessPointDeleted waits for an Access Point to return Deleted
func waitAccessPointDeleted(ctx context.Context, conn *efs.Client, accessPointId string) (*awstypes.AccessPointDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.LifeCycleStateAvailable, awstypes.LifeCycleStateDeleting, awstypes.LifeCycleStateDeleted),
		Target:  []string{},
		Refresh: statusAccessPointLifeCycleState(ctx, conn, accessPointId),
		Timeout: accessPointDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.AccessPointDescription); ok {
		return output, err
	}

	return nil, err
}

func waitBackupPolicyDisabled(ctx context.Context, conn *efs.Client, id string) (*awstypes.BackupPolicy, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.StatusDisabling),
		Target:  enum.Slice(awstypes.StatusDisabled),
		Refresh: statusBackupPolicy(ctx, conn, id),
		Timeout: backupPolicyDisabledTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.BackupPolicy); ok {
		return output, err
	}

	return nil, err
}

func waitBackupPolicyEnabled(ctx context.Context, conn *efs.Client, id string) (*awstypes.BackupPolicy, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.StatusEnabling),
		Target:  enum.Slice(awstypes.StatusEnabled),
		Refresh: statusBackupPolicy(ctx, conn, id),
		Timeout: backupPolicyEnabledTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.BackupPolicy); ok {
		return output, err
	}

	return nil, err
}
