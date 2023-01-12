package efs

import (
	"time"

	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// Maximum amount of time to wait for an Operation to return Success
	accessPointCreatedTimeout = 10 * time.Minute
	accessPointDeletedTimeout = 10 * time.Minute

	backupPolicyDisabledTimeout = 10 * time.Minute
	backupPolicyEnabledTimeout  = 10 * time.Minute
)

// waitAccessPointCreated waits for an Operation to return Success
func waitAccessPointCreated(conn *efs.EFS, accessPointId string) (*efs.AccessPointDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{efs.LifeCycleStateCreating},
		Target:  []string{efs.LifeCycleStateAvailable},
		Refresh: statusAccessPointLifeCycleState(conn, accessPointId),
		Timeout: accessPointCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*efs.AccessPointDescription); ok {
		return output, err
	}

	return nil, err
}

// waitAccessPointDeleted waits for an Access Point to return Deleted
func waitAccessPointDeleted(conn *efs.EFS, accessPointId string) (*efs.AccessPointDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{efs.LifeCycleStateAvailable, efs.LifeCycleStateDeleting, efs.LifeCycleStateDeleted},
		Target:  []string{},
		Refresh: statusAccessPointLifeCycleState(conn, accessPointId),
		Timeout: accessPointDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*efs.AccessPointDescription); ok {
		return output, err
	}

	return nil, err
}

func waitBackupPolicyDisabled(conn *efs.EFS, id string) (*efs.BackupPolicy, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{efs.StatusDisabling},
		Target:  []string{efs.StatusDisabled},
		Refresh: statusBackupPolicy(conn, id),
		Timeout: backupPolicyDisabledTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*efs.BackupPolicy); ok {
		return output, err
	}

	return nil, err
}

func waitBackupPolicyEnabled(conn *efs.EFS, id string) (*efs.BackupPolicy, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{efs.StatusEnabling},
		Target:  []string{efs.StatusEnabled},
		Refresh: statusBackupPolicy(conn, id),
		Timeout: backupPolicyEnabledTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*efs.BackupPolicy); ok {
		return output, err
	}

	return nil, err
}

func waitReplicationConfigurationCreated(conn *efs.EFS, id string, timeout time.Duration) (*efs.ReplicationConfigurationDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{efs.ReplicationStatusEnabling},
		Target:  []string{efs.ReplicationStatusEnabled},
		Refresh: statusReplicationConfiguration(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*efs.ReplicationConfigurationDescription); ok {
		return output, err
	}

	return nil, err
}

func waitReplicationConfigurationDeleted(conn *efs.EFS, id string, timeout time.Duration) (*efs.ReplicationConfigurationDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending:                   []string{efs.ReplicationStatusDeleting},
		Target:                    []string{},
		Refresh:                   statusReplicationConfiguration(conn, id),
		Timeout:                   timeout,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*efs.ReplicationConfigurationDescription); ok {
		return output, err
	}

	return nil, err
}
