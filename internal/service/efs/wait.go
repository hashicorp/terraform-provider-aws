// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package efs

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

const (
	// Maximum amount of time to wait for an Operation to return Success
	accessPointCreatedTimeout = 10 * time.Minute
	accessPointDeletedTimeout = 10 * time.Minute

	backupPolicyDisabledTimeout = 10 * time.Minute
	backupPolicyEnabledTimeout  = 10 * time.Minute
)

// waitAccessPointCreated waits for an Operation to return Success
func waitAccessPointCreated(ctx context.Context, conn *efs.EFS, accessPointId string) (*efs.AccessPointDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{efs.LifeCycleStateCreating},
		Target:  []string{efs.LifeCycleStateAvailable},
		Refresh: statusAccessPointLifeCycleState(ctx, conn, accessPointId),
		Timeout: accessPointCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*efs.AccessPointDescription); ok {
		return output, err
	}

	return nil, err
}

// waitAccessPointDeleted waits for an Access Point to return Deleted
func waitAccessPointDeleted(ctx context.Context, conn *efs.EFS, accessPointId string) (*efs.AccessPointDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{efs.LifeCycleStateAvailable, efs.LifeCycleStateDeleting, efs.LifeCycleStateDeleted},
		Target:  []string{},
		Refresh: statusAccessPointLifeCycleState(ctx, conn, accessPointId),
		Timeout: accessPointDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*efs.AccessPointDescription); ok {
		return output, err
	}

	return nil, err
}

func waitBackupPolicyDisabled(ctx context.Context, conn *efs.EFS, id string) (*efs.BackupPolicy, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{efs.StatusDisabling},
		Target:  []string{efs.StatusDisabled},
		Refresh: statusBackupPolicy(ctx, conn, id),
		Timeout: backupPolicyDisabledTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*efs.BackupPolicy); ok {
		return output, err
	}

	return nil, err
}

func waitBackupPolicyEnabled(ctx context.Context, conn *efs.EFS, id string) (*efs.BackupPolicy, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{efs.StatusEnabling},
		Target:  []string{efs.StatusEnabled},
		Refresh: statusBackupPolicy(ctx, conn, id),
		Timeout: backupPolicyEnabledTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*efs.BackupPolicy); ok {
		return output, err
	}

	return nil, err
}

func waitReplicationConfigurationCreated(ctx context.Context, conn *efs.EFS, id string, timeout time.Duration) (*efs.ReplicationConfigurationDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{efs.ReplicationStatusEnabling},
		Target:  []string{efs.ReplicationStatusEnabled},
		Refresh: statusReplicationConfiguration(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*efs.ReplicationConfigurationDescription); ok {
		return output, err
	}

	return nil, err
}

func waitReplicationConfigurationDeleted(ctx context.Context, conn *efs.EFS, id string, timeout time.Duration) (*efs.ReplicationConfigurationDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   []string{efs.ReplicationStatusDeleting},
		Target:                    []string{},
		Refresh:                   statusReplicationConfiguration(ctx, conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*efs.ReplicationConfigurationDescription); ok {
		return output, err
	}

	return nil, err
}
