package waiter

import (
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
)

const (
	BackupAvailableTimeout = 10 * time.Minute
	BackupDeletedTimeout   = 10 * time.Minute
)

func AdministrativeActionCompleted(conn *fsx.FSx, fsID, actionType string, timeout time.Duration) (*fsx.AdministrativeAction, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{fsx.StatusInProgress, fsx.StatusPending},
		Target:  []string{fsx.StatusCompleted, fsx.StatusUpdatedOptimizing},
		Refresh: AdministrativeActionStatus(conn, fsID, actionType),
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

func BackupAvailable(conn *fsx.FSx, id string) (*fsx.Backup, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{fsx.BackupLifecycleCreating, fsx.BackupLifecyclePending, fsx.BackupLifecycleTransferring},
		Target:  []string{fsx.BackupLifecycleAvailable},
		Refresh: BackupStatus(conn, id),
		Timeout: BackupAvailableTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*fsx.Backup); ok {
		return output, err
	}

	return nil, err
}

func BackupDeleted(conn *fsx.FSx, id string) (*fsx.Backup, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{fsx.FileSystemLifecycleDeleting},
		Target:  []string{},
		Refresh: BackupStatus(conn, id),
		Timeout: BackupDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*fsx.Backup); ok {
		return output, err
	}

	return nil, err
}

func FileSystemCreated(conn *fsx.FSx, id string, timeout time.Duration) (*fsx.FileSystem, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{fsx.FileSystemLifecycleCreating},
		Target:  []string{fsx.FileSystemLifecycleAvailable},
		Refresh: FileSystemStatus(conn, id),
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

func FileSystemUpdated(conn *fsx.FSx, id string, timeout time.Duration) (*fsx.FileSystem, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{fsx.FileSystemLifecycleUpdating},
		Target:  []string{fsx.FileSystemLifecycleAvailable},
		Refresh: FileSystemStatus(conn, id),
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

func FileSystemDeleted(conn *fsx.FSx, id string, timeout time.Duration) (*fsx.FileSystem, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{fsx.FileSystemLifecycleAvailable, fsx.FileSystemLifecycleDeleting},
		Target:  []string{},
		Refresh: FileSystemStatus(conn, id),
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
