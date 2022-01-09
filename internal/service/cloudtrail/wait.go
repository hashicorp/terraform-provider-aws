package cloudtrail

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/cloudtrail"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	eventDataStoreAvailableTimeout = 5 * time.Minute
	eventDataStoreDeletedTimeout   = 5 * time.Minute
)

// waitEventDataStoreAvailable waits for Event Data Store to reach the available state.
func waitEventDataStoreAvailable(ctx context.Context, conn *cloudtrail.CloudTrail, snapshotId string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{eventDataStoreStatusCreating},
		Target:  []string{eventDataStoreStatusAvailable},
		Refresh: statusEventDataStore(ctx, conn, snapshotId),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

// waitEventDataStoreDeleted waits for Event Data Store to be deleted.
func waitEventDataStoreDeleted(ctx context.Context, conn *cloudtrail.CloudTrail, snapshotId string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{eventDataStoreStatusDeleting},
		Target:  []string{},
		Refresh: statusEventDataStore(ctx, conn, snapshotId),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}
