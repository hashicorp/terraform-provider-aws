package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	BackupAvailableTimeout = 10 * time.Minute
	BackupDeletedTimeout   = 10 * time.Minute
)

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
