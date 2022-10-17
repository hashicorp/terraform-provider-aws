package cloudtrail

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/cloudtrail"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func waitEventDataStoreAvailable(ctx context.Context, conn *cloudtrail.CloudTrail, arn string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{cloudtrail.EventDataStoreStatusCreated},
		Target:  []string{cloudtrail.EventDataStoreStatusEnabled},
		Refresh: statusEventDataStore(ctx, conn, arn),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitEventDataStoreDeleted(ctx context.Context, conn *cloudtrail.CloudTrail, arn string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{cloudtrail.EventDataStoreStatusCreated, cloudtrail.EventDataStoreStatusEnabled},
		Target:  []string{},
		Refresh: statusEventDataStore(ctx, conn, arn),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}
