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
	eventDataStoreStatusCreating   = "CREATED"
	eventDataStoreStatusAvailable  = "ENABLED"
	eventDataStoreStatusDeleting   = "PENDING_DELETION"
)

// waitEventDataStoreAvailable waits for CloudTrail Event Data Store to reach the available state.
func waitEventDataStoreAvailable(ctx context.Context, conn *cloudtrail.CloudTrail, eventDataStoreArn string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{eventDataStoreStatusCreating},
		Target:  []string{eventDataStoreStatusAvailable},
		Refresh: statusEventDataStore(ctx, conn, eventDataStoreArn),
		Timeout: eventDataStoreAvailableTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

// waitEventDataStoreDeleted waits for CloudTrail Event Data Store to be deleted.
func waitEventDataStoreDeleted(ctx context.Context, conn *cloudtrail.CloudTrail, eventDataStoreArn string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{},
		Target:  []string{eventDataStoreStatusDeleting},
		Refresh: statusEventDataStore(ctx, conn, eventDataStoreArn),
		Timeout: eventDataStoreDeletedTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}
