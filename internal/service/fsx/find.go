package fsx

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindAdministrativeActionByFileSystemIDAndActionType(ctx context.Context, conn *fsx.FSx, fsID, actionType string) (*fsx.AdministrativeAction, error) {
	fileSystem, err := FindFileSystemByID(ctx, conn, fsID)

	if err != nil {
		return nil, err
	}

	for _, administrativeAction := range fileSystem.AdministrativeActions {
		if administrativeAction == nil {
			continue
		}

		if aws.StringValue(administrativeAction.AdministrativeActionType) == actionType {
			return administrativeAction, nil
		}
	}

	// If the administrative action isn't found, assume it's complete.
	return &fsx.AdministrativeAction{Status: aws.String(fsx.StatusCompleted)}, nil
}

func FindBackupByID(ctx context.Context, conn *fsx.FSx, id string) (*fsx.Backup, error) {
	input := &fsx.DescribeBackupsInput{
		BackupIds: aws.StringSlice([]string{id}),
	}

	output, err := conn.DescribeBackupsWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, fsx.ErrCodeFileSystemNotFound) || tfawserr.ErrCodeEquals(err, fsx.ErrCodeBackupNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Backups) == 0 || output.Backups[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Backups[0], nil
}

func findFileCacheByID(ctx context.Context, conn *fsx.FSx, id string) (*fsx.FileCache, error) {
	input := &fsx.DescribeFileCachesInput{
		FileCacheIds: []*string{aws.String(id)},
	}
	var fileCaches []*fsx.FileCache

	err := conn.DescribeFileCachesPagesWithContext(ctx, input, func(page *fsx.DescribeFileCachesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}
		fileCaches = append(fileCaches, page.FileCaches...)

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, fsx.ErrCodeFileCacheNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}
	if err != nil {
		return nil, err
	}
	if len(fileCaches) == 0 || fileCaches[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}
	if count := len(fileCaches); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}
	return fileCaches[0], nil
}

func findDataRepositoryAssociationsByIDs(ctx context.Context, conn *fsx.FSx, ids []*string) ([]*fsx.DataRepositoryAssociation, error) {
	input := &fsx.DescribeDataRepositoryAssociationsInput{
		AssociationIds: ids,
	}
	var dataRepositoryAssociations []*fsx.DataRepositoryAssociation

	err := conn.DescribeDataRepositoryAssociationsPagesWithContext(ctx, input, func(page *fsx.DescribeDataRepositoryAssociationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}
		dataRepositoryAssociations = append(dataRepositoryAssociations, page.Associations...)
		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, fsx.ErrCodeDataRepositoryAssociationNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}
	if err != nil {
		return nil, err
	}
	if len(dataRepositoryAssociations) == 0 || dataRepositoryAssociations[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return dataRepositoryAssociations, nil
}

func FindFileSystemByID(ctx context.Context, conn *fsx.FSx, id string) (*fsx.FileSystem, error) {
	input := &fsx.DescribeFileSystemsInput{
		FileSystemIds: []*string{aws.String(id)},
	}

	var filesystems []*fsx.FileSystem

	err := conn.DescribeFileSystemsPagesWithContext(ctx, input, func(page *fsx.DescribeFileSystemsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		filesystems = append(filesystems, page.FileSystems...)

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, fsx.ErrCodeFileSystemNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if len(filesystems) == 0 || filesystems[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(filesystems); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return filesystems[0], nil
}

func FindDataRepositoryAssociationByID(ctx context.Context, conn *fsx.FSx, id string) (*fsx.DataRepositoryAssociation, error) {
	input := &fsx.DescribeDataRepositoryAssociationsInput{
		AssociationIds: []*string{aws.String(id)},
	}

	var associations []*fsx.DataRepositoryAssociation

	err := conn.DescribeDataRepositoryAssociationsPagesWithContext(ctx, input, func(page *fsx.DescribeDataRepositoryAssociationsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		associations = append(associations, page.Associations...)

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, fsx.ErrCodeDataRepositoryAssociationNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if len(associations) == 0 || associations[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(associations); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return associations[0], nil
}

func FindStorageVirtualMachineByID(ctx context.Context, conn *fsx.FSx, id string) (*fsx.StorageVirtualMachine, error) {
	input := &fsx.DescribeStorageVirtualMachinesInput{
		StorageVirtualMachineIds: []*string{aws.String(id)},
	}

	var storageVirtualMachines []*fsx.StorageVirtualMachine

	err := conn.DescribeStorageVirtualMachinesPagesWithContext(ctx, input, func(page *fsx.DescribeStorageVirtualMachinesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		storageVirtualMachines = append(storageVirtualMachines, page.StorageVirtualMachines...)

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, fsx.ErrCodeStorageVirtualMachineNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if len(storageVirtualMachines) == 0 || storageVirtualMachines[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(storageVirtualMachines); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return storageVirtualMachines[0], nil
}

func FindVolumeByID(ctx context.Context, conn *fsx.FSx, id string) (*fsx.Volume, error) {
	input := &fsx.DescribeVolumesInput{
		VolumeIds: []*string{aws.String(id)},
	}

	var volumes []*fsx.Volume

	err := conn.DescribeVolumesPagesWithContext(ctx, input, func(page *fsx.DescribeVolumesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		volumes = append(volumes, page.Volumes...)

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, fsx.ErrCodeVolumeNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if len(volumes) == 0 || volumes[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(volumes); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	return volumes[0], nil
}

func FindSnapshotByID(ctx context.Context, conn *fsx.FSx, id string) (*fsx.Snapshot, error) {
	input := &fsx.DescribeSnapshotsInput{
		SnapshotIds: aws.StringSlice([]string{id}),
	}

	output, err := conn.DescribeSnapshotsWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, fsx.ErrCodeVolumeNotFound) || tfawserr.ErrCodeEquals(err, fsx.ErrCodeSnapshotNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Snapshots) == 0 || output.Snapshots[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Snapshots[0], nil
}

func FindSnapshots(ctx context.Context, conn *fsx.FSx, input *fsx.DescribeSnapshotsInput) ([]*fsx.Snapshot, error) {
	var output []*fsx.Snapshot

	err := conn.DescribeSnapshotsPagesWithContext(ctx, input, func(page *fsx.DescribeSnapshotsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Snapshots {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, fsx.ErrCodeSnapshotNotFound) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}
