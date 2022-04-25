package docdb

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/service/docdb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	GlobalClusterCreateTimeout = 5 * time.Minute
	GlobalClusterDeleteTimeout = 5 * time.Minute
	GlobalClusterUpdateTimeout = 5 * time.Minute
)

const (
	GlobalClusterStatusAvailable = "available"
	GlobalClusterStatusCreating  = "creating"
	GlobalClusterStatusDeleted   = "deleted"
	GlobalClusterStatusDeleting  = "deleting"
	GlobalClusterStatusModifying = "modifying"
	GlobalClusterStatusUpgrading = "upgrading"
)

func WaitForGlobalClusterDeletion(ctx context.Context, conn *docdb.DocDB, globalClusterID string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending:        []string{GlobalClusterStatusAvailable, GlobalClusterStatusDeleting},
		Target:         []string{GlobalClusterStatusDeleted},
		Refresh:        statusGlobalClusterRefreshFunc(ctx, conn, globalClusterID),
		Timeout:        timeout,
		NotFoundChecks: 1,
	}

	log.Printf("[DEBUG] Waiting for DocDB Global Cluster (%s) deletion", globalClusterID)
	_, err := stateConf.WaitForStateContext(ctx)

	if tfresource.NotFound(err) {
		return nil
	}

	return err
}

func waitEventSubscriptionActive(ctx context.Context, conn *docdb.DocDB, id string, timeout time.Duration) (*docdb.EventSubscription, error) { //nolint:unparam
	stateConf := &resource.StateChangeConf{
		Pending: []string{"creating", "modifying"},
		Target:  []string{"active"},
		Refresh: statusEventSubscription(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*docdb.EventSubscription); ok {
		return output, err
	}

	return nil, err
}

func waitEventSubscriptionDeleted(ctx context.Context, conn *docdb.DocDB, id string, timeout time.Duration) (*docdb.EventSubscription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"deleting"},
		Target:  []string{},
		Refresh: statusEventSubscription(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*docdb.EventSubscription); ok {
		return output, err
	}

	return nil, err
}
