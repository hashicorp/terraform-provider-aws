package waiter

import (
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	backupAvailableTimeout = 10 * time.Minute
	backupDeletedTimeout   = 10 * time.Minute
)

func waitAdministrativeActionCompleted(conn *fsx.FSx, fsID, actionType string, timeout time.Duration) (*fsx.AdministrativeAction, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{fsx.StatusInProgress, fsx.StatusPending},
		Target:  []string{fsx.StatusCompleted, fsx.StatusUpdatedOptimizing},
		Refresh: statusAdministrativeAction(conn, fsID, actionType),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*fsx.AdministrativeAction); ok {
		if status, details := aws.StringValue(output.Status), output.FailureDetails; status == fsx.StatusFailed && details != nil {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.FailureDetails.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitBackupAvailable(conn *fsx.FSx, id string) (*fsx.Backup, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{fsx.BackupLifecycleCreating, fsx.BackupLifecyclePending, fsx.BackupLifecycleTransferring},
		Target:  []string{fsx.BackupLifecycleAvailable},
		Refresh: statusBackup(conn, id),
		Timeout: backupAvailableTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*fsx.Backup); ok {
		return output, err
	}

	return nil, err
}

func waitBackupDeleted(conn *fsx.FSx, id string) (*fsx.Backup, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{fsx.FileSystemLifecycleDeleting},
		Target:  []string{},
		Refresh: statusBackup(conn, id),
		Timeout: backupDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*fsx.Backup); ok {
		return output, err
	}

	return nil, err
}

func waitFileSystemCreated(conn *fsx.FSx, id string, timeout time.Duration) (*fsx.FileSystem, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{fsx.FileSystemLifecycleCreating},
		Target:  []string{fsx.FileSystemLifecycleAvailable},
		Refresh: statusFileSystem(conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*fsx.FileSystem); ok {
		if status, details := aws.StringValue(output.Lifecycle), output.FailureDetails; status == fsx.FileSystemLifecycleFailed && details != nil {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.FailureDetails.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitFileSystemUpdated(conn *fsx.FSx, id string, timeout time.Duration) (*fsx.FileSystem, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{fsx.FileSystemLifecycleUpdating},
		Target:  []string{fsx.FileSystemLifecycleAvailable},
		Refresh: statusFileSystem(conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*fsx.FileSystem); ok {
		if status, details := aws.StringValue(output.Lifecycle), output.FailureDetails; status == fsx.FileSystemLifecycleFailed && details != nil {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.FailureDetails.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitFileSystemDeleted(conn *fsx.FSx, id string, timeout time.Duration) (*fsx.FileSystem, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{fsx.FileSystemLifecycleAvailable, fsx.FileSystemLifecycleDeleting},
		Target:  []string{},
		Refresh: statusFileSystem(conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*fsx.FileSystem); ok {
		if status, details := aws.StringValue(output.Lifecycle), output.FailureDetails; status == fsx.FileSystemLifecycleFailed && details != nil {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.FailureDetails.Message)))
		}

		return output, err
	}

	return nil, err
}
