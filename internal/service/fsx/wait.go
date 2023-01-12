package fsx

import (
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	backupAvailableTimeout = 10 * time.Minute
	backupDeletedTimeout   = 10 * time.Minute
)

func waitAdministrativeActionCompleted(conn *fsx.FSx, fsID, actionType string, timeout time.Duration) (*fsx.AdministrativeAction, error) { //nolint:unparam
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

func waitFileCacheCreated(conn *fsx.FSx, id string, timeout time.Duration) (*fsx.FileCache, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{fsx.FileCacheLifecycleCreating},
		Target:  []string{fsx.FileCacheLifecycleAvailable},
		Refresh: statusFileCache(conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*fsx.FileCache); ok {
		if status, details := aws.StringValue(output.Lifecycle), output.FailureDetails; status == fsx.FileCacheLifecycleFailed && details != nil {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.FailureDetails.Message)))
		}
		return output, err
	}
	return nil, err
}

func waitFileCacheUpdated(conn *fsx.FSx, id string, timeout time.Duration) (*fsx.FileCache, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{fsx.FileCacheLifecycleUpdating},
		Target:  []string{fsx.FileCacheLifecycleAvailable},
		Refresh: statusFileCache(conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*fsx.FileCache); ok {
		if status, details := aws.StringValue(output.Lifecycle), output.FailureDetails; status == fsx.FileCacheLifecycleFailed && details != nil {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.FailureDetails.Message)))
		}
		return output, err
	}

	return nil, err
}

func waitFileCacheDeleted(conn *fsx.FSx, id string, timeout time.Duration) (*fsx.FileCache, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{fsx.FileCacheLifecycleAvailable, fsx.FileCacheLifecycleDeleting},
		Target:  []string{},
		Refresh: statusFileCache(conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*fsx.FileCache); ok {
		if status, details := aws.StringValue(output.Lifecycle), output.FailureDetails; status == fsx.FileCacheLifecycleFailed && details != nil {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.FailureDetails.Message)))
		}
		return output, err
	}

	return nil, err
}

func waitFileSystemCreated(conn *fsx.FSx, id string, timeout time.Duration) (*fsx.FileSystem, error) { //nolint:unparam
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

func waitFileSystemUpdated(conn *fsx.FSx, id string, timeout time.Duration) (*fsx.FileSystem, error) { //nolint:unparam
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

func waitFileSystemDeleted(conn *fsx.FSx, id string, timeout time.Duration) (*fsx.FileSystem, error) { //nolint:unparam
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

func waitDataRepositoryAssociationCreated(conn *fsx.FSx, id string, timeout time.Duration) (*fsx.DataRepositoryAssociation, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{fsx.DataRepositoryLifecycleCreating},
		Target:  []string{fsx.DataRepositoryLifecycleAvailable},
		Refresh: statusDataRepositoryAssociation(conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*fsx.DataRepositoryAssociation); ok {
		if status, details := aws.StringValue(output.Lifecycle), output.FailureDetails; status == fsx.DataRepositoryLifecycleFailed && details != nil {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.FailureDetails.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitDataRepositoryAssociationUpdated(conn *fsx.FSx, id string, timeout time.Duration) (*fsx.DataRepositoryAssociation, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{fsx.DataRepositoryLifecycleUpdating},
		Target:  []string{fsx.DataRepositoryLifecycleAvailable},
		Refresh: statusDataRepositoryAssociation(conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*fsx.DataRepositoryAssociation); ok {
		if status, details := aws.StringValue(output.Lifecycle), output.FailureDetails; status == fsx.DataRepositoryLifecycleFailed && details != nil {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.FailureDetails.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitDataRepositoryAssociationDeleted(conn *fsx.FSx, id string, timeout time.Duration) (*fsx.DataRepositoryAssociation, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{fsx.DataRepositoryLifecycleAvailable, fsx.DataRepositoryLifecycleDeleting},
		Target:  []string{},
		Refresh: statusDataRepositoryAssociation(conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*fsx.DataRepositoryAssociation); ok {
		if status, details := aws.StringValue(output.Lifecycle), output.FailureDetails; status == fsx.DataRepositoryLifecycleFailed && details != nil {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.FailureDetails.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitStorageVirtualMachineCreated(conn *fsx.FSx, id string, timeout time.Duration) (*fsx.StorageVirtualMachine, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{fsx.StorageVirtualMachineLifecycleCreating, fsx.StorageVirtualMachineLifecyclePending},
		Target:  []string{fsx.StorageVirtualMachineLifecycleCreated, fsx.StorageVirtualMachineLifecycleMisconfigured},
		Refresh: statusStorageVirtualMachine(conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*fsx.StorageVirtualMachine); ok {
		if status, details := aws.StringValue(output.Lifecycle), output.LifecycleTransitionReason; status == fsx.FileSystemLifecycleFailed && details != nil {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.LifecycleTransitionReason.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitStorageVirtualMachineUpdated(conn *fsx.FSx, id string, timeout time.Duration) (*fsx.StorageVirtualMachine, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{fsx.StorageVirtualMachineLifecyclePending},
		Target:  []string{fsx.StorageVirtualMachineLifecycleCreated, fsx.StorageVirtualMachineLifecycleMisconfigured},
		Refresh: statusStorageVirtualMachine(conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*fsx.StorageVirtualMachine); ok {
		if status, details := aws.StringValue(output.Lifecycle), output.LifecycleTransitionReason; status == fsx.FileSystemLifecycleFailed && details != nil {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.LifecycleTransitionReason.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitStorageVirtualMachineDeleted(conn *fsx.FSx, id string, timeout time.Duration) (*fsx.StorageVirtualMachine, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{fsx.StorageVirtualMachineLifecycleCreated, fsx.StorageVirtualMachineLifecycleDeleting},
		Target:  []string{},
		Refresh: statusStorageVirtualMachine(conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*fsx.StorageVirtualMachine); ok {
		if status, details := aws.StringValue(output.Lifecycle), output.LifecycleTransitionReason; status == fsx.StorageVirtualMachineLifecycleFailed && details != nil {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.LifecycleTransitionReason.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitVolumeCreated(conn *fsx.FSx, id string, timeout time.Duration) (*fsx.Volume, error) { //nolint:unparam
	stateConf := &resource.StateChangeConf{
		Pending: []string{fsx.VolumeLifecycleCreating, fsx.VolumeLifecyclePending},
		Target:  []string{fsx.VolumeLifecycleCreated, fsx.VolumeLifecycleMisconfigured, fsx.VolumeLifecycleAvailable},
		Refresh: statusVolume(conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*fsx.Volume); ok {
		if status, details := aws.StringValue(output.Lifecycle), output.LifecycleTransitionReason; status == fsx.VolumeLifecycleFailed && details != nil {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.LifecycleTransitionReason.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitVolumeUpdated(conn *fsx.FSx, id string, timeout time.Duration) (*fsx.Volume, error) { //nolint:unparam
	stateConf := &resource.StateChangeConf{
		Pending: []string{fsx.VolumeLifecyclePending},
		Target:  []string{fsx.VolumeLifecycleCreated, fsx.VolumeLifecycleMisconfigured, fsx.VolumeLifecycleAvailable},
		Refresh: statusVolume(conn, id),
		Timeout: timeout,
		Delay:   150 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*fsx.Volume); ok {
		if status, details := aws.StringValue(output.Lifecycle), output.LifecycleTransitionReason; status == fsx.VolumeLifecycleFailed && details != nil {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.LifecycleTransitionReason.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitVolumeDeleted(conn *fsx.FSx, id string, timeout time.Duration) (*fsx.Volume, error) { //nolint:unparam
	stateConf := &resource.StateChangeConf{
		Pending: []string{fsx.VolumeLifecycleCreated, fsx.VolumeLifecycleMisconfigured, fsx.VolumeLifecycleAvailable, fsx.VolumeLifecycleDeleting},
		Target:  []string{},
		Refresh: statusVolume(conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*fsx.Volume); ok {
		if status, details := aws.StringValue(output.Lifecycle), output.LifecycleTransitionReason; status == fsx.VolumeLifecycleFailed && details != nil {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.LifecycleTransitionReason.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitSnapshotCreated(conn *fsx.FSx, id string, timeout time.Duration) (*fsx.Snapshot, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{fsx.SnapshotLifecycleCreating, fsx.SnapshotLifecyclePending},
		Target:  []string{fsx.SnapshotLifecycleAvailable},
		Refresh: statusSnapshot(conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*fsx.Snapshot); ok {
		return output, err
	}

	return nil, err
}

func waitSnapshotUpdated(conn *fsx.FSx, id string, timeout time.Duration) (*fsx.Snapshot, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{fsx.SnapshotLifecyclePending},
		Target:  []string{fsx.SnapshotLifecycleAvailable},
		Refresh: statusSnapshot(conn, id),
		Timeout: timeout,
		Delay:   150 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*fsx.Snapshot); ok {
		return output, err
	}

	return nil, err
}

func waitSnapshotDeleted(conn *fsx.FSx, id string, timeout time.Duration) (*fsx.Snapshot, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{fsx.SnapshotLifecyclePending, fsx.SnapshotLifecycleDeleting},
		Target:  []string{},
		Refresh: statusSnapshot(conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*fsx.Snapshot); ok {
		return output, err
	}

	return nil, err
}
