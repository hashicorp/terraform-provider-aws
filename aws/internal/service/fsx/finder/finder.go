package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
)

func AdministrativeActionByFileSystemIDAndActionType(conn *fsx.FSx, fsID, actionType string) (*fsx.AdministrativeAction, error) {
	fileSystem, err := FileSystemByID(conn, fsID)

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

func BackupByID(conn *fsx.FSx, id string) (*fsx.Backup, error) {
	input := &fsx.DescribeBackupsInput{
		BackupIds: aws.StringSlice([]string{id}),
	}

	output, err := conn.DescribeBackups(input)

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

func FileSystemByID(conn *fsx.FSx, id string) (*fsx.FileSystem, error) {
	input := &fsx.DescribeFileSystemsInput{
		FileSystemIds: []*string{aws.String(id)},
	}

	var filesystems []*fsx.FileSystem

	err := conn.DescribeFileSystemsPages(input, func(page *fsx.DescribeFileSystemsOutput, lastPage bool) bool {
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
