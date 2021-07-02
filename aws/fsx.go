package aws

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	fsxWindowsFileSystemAliasAvailable = 10 * time.Minute
	fsxWindowsFileSystemAliasDeleted   = 10 * time.Minute
)

func describeFsxFileSystem(conn *fsx.FSx, id string) (*fsx.FileSystem, error) {
	input := &fsx.DescribeFileSystemsInput{
		FileSystemIds: []*string{aws.String(id)},
	}
	var filesystem *fsx.FileSystem

	err := conn.DescribeFileSystemsPages(input, func(page *fsx.DescribeFileSystemsOutput, lastPage bool) bool {
		for _, fs := range page.FileSystems {
			if aws.StringValue(fs.FileSystemId) == id {
				filesystem = fs
				return false
			}
		}

		return !lastPage
	})

	return filesystem, err
}

func refreshFsxFileSystemLifecycle(conn *fsx.FSx, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		filesystem, err := describeFsxFileSystem(conn, id)

		if isAWSErr(err, fsx.ErrCodeFileSystemNotFound, "") {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if filesystem == nil {
			return nil, "", nil
		}

		return filesystem, aws.StringValue(filesystem.Lifecycle), nil
	}
}

func refreshFsxWindowsFileSystemAliasLifecycle(conn *fsx.FSx, id, aliasName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		filesystem, err := describeFsxFileSystem(conn, id)

		if isAWSErr(err, fsx.ErrCodeFileSystemNotFound, "") {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if filesystem == nil {
			return nil, "", nil
		}

		if filesystem.WindowsConfiguration == nil {
			return nil, "", nil
		}

		aliases := filesystem.WindowsConfiguration.Aliases
		if aliases == nil {
			return nil, "", nil
		}

		for _, alias := range aliases {
			if alias == nil {
				continue
			}

			if aws.StringValue(alias.Name) == aliasName {
				return filesystem, aws.StringValue(alias.Lifecycle), nil
			}
		}

		return filesystem, "", nil
	}
}

func refreshFsxFileSystemAdministrativeActionsStatusFileSystemUpdate(conn *fsx.FSx, id string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		filesystem, err := describeFsxFileSystem(conn, id)

		if isAWSErr(err, fsx.ErrCodeFileSystemNotFound, "") {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if filesystem == nil {
			return nil, "", nil
		}

		for _, administrativeAction := range filesystem.AdministrativeActions {
			if administrativeAction == nil {
				continue
			}

			if aws.StringValue(administrativeAction.AdministrativeActionType) == fsx.AdministrativeActionTypeFileSystemUpdate {
				return filesystem, aws.StringValue(administrativeAction.Status), nil
			}
		}

		return filesystem, fsx.StatusCompleted, nil
	}
}

func waitForFsxWindowsFileSystemAliasAvailable(conn *fsx.FSx, id, alias string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{fsx.AliasLifecycleCreating},
		Target:  []string{fsx.AliasLifecycleAvailable},
		Refresh: refreshFsxWindowsFileSystemAliasLifecycle(conn, id, alias),
		Timeout: fsxWindowsFileSystemAliasAvailable,
		Delay:   30 * time.Second,
	}

	_, err := stateConf.WaitForState()

	return err
}

func waitForFsxWindowsFileSystemAliasDeleted(conn *fsx.FSx, id, alias string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{fsx.AliasLifecycleAvailable, fsx.AliasLifecycleDeleting},
		Target:  []string{""},
		Refresh: refreshFsxWindowsFileSystemAliasLifecycle(conn, id, alias),
		Timeout: fsxWindowsFileSystemAliasDeleted,
		Delay:   30 * time.Second,
	}

	_, err := stateConf.WaitForState()

	return err
}

func waitForFsxFileSystemCreation(conn *fsx.FSx, id string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{fsx.FileSystemLifecycleCreating},
		Target:  []string{fsx.FileSystemLifecycleAvailable},
		Refresh: refreshFsxFileSystemLifecycle(conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	_, err := stateConf.WaitForState()

	return err
}

func waitForFsxFileSystemDeletion(conn *fsx.FSx, id string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{fsx.FileSystemLifecycleAvailable, fsx.FileSystemLifecycleDeleting},
		Target:  []string{},
		Refresh: refreshFsxFileSystemLifecycle(conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	_, err := stateConf.WaitForState()

	return err
}

func waitForFsxFileSystemUpdate(conn *fsx.FSx, id string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{fsx.FileSystemLifecycleUpdating},
		Target:  []string{fsx.FileSystemLifecycleAvailable},
		Refresh: refreshFsxFileSystemLifecycle(conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	_, err := stateConf.WaitForState()

	return err
}

func waitForFsxFileSystemUpdateAdministrativeActionsStatusFileSystemUpdate(conn *fsx.FSx, id string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			fsx.StatusInProgress,
			fsx.StatusPending,
		},
		Target: []string{
			fsx.StatusCompleted,
			fsx.StatusUpdatedOptimizing,
		},
		Refresh: refreshFsxFileSystemAdministrativeActionsStatusFileSystemUpdate(conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	_, err := stateConf.WaitForState()

	return err
}
