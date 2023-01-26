package neptune

import (
	"context"
	"log"
	"time"

	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	// Maximum amount of time to wait for an EventSubscription to return Deleted
	EventSubscriptionDeletedTimeout = 10 * time.Minute

	// Maximum amount of time to wait for an DBClusterEndpoint to return Available
	DBClusterEndpointAvailableTimeout = 10 * time.Minute

	// Maximum amount of time to wait for an DBClusterEndpoint to return Deleted
	DBClusterEndpointDeletedTimeout = 10 * time.Minute
)

// WaitEventSubscriptionDeleted waits for a EventSubscription to return Deleted
func WaitEventSubscriptionDeleted(ctx context.Context, conn *neptune.Neptune, subscriptionName string) (*neptune.EventSubscription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"deleting"},
		Target:  []string{EventSubscriptionStatusNotFound},
		Refresh: StatusEventSubscription(ctx, conn, subscriptionName),
		Timeout: EventSubscriptionDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*neptune.EventSubscription); ok {
		return v, err
	}

	return nil, err
}

// WaitDBClusterDeleted waits for a Cluster to return Deleted
func WaitDBClusterDeleted(ctx context.Context, conn *neptune.Neptune, id string, timeout time.Duration) (*neptune.DBCluster, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			"available",
			"deleting",
			"backing-up",
			"modifying",
		},
		Target:     []string{ClusterStatusNotFound},
		Refresh:    StatusCluster(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*neptune.DBCluster); ok {
		return v, err
	}

	return nil, err
}

// WaitDBClusterAvailable waits for a Cluster to return Available
func WaitDBClusterAvailable(ctx context.Context, conn *neptune.Neptune, id string, timeout time.Duration) (*neptune.DBCluster, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			"creating",
			"backing-up",
			"modifying",
			"preparing-data-migration",
			"migrating",
			"configuring-iam-database-auth",
			"upgrading",
		},
		Target:     []string{"available"},
		Refresh:    StatusCluster(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*neptune.DBCluster); ok {
		return v, err
	}

	return nil, err
}

// WaitDBClusterEndpointAvailable waits for a DBClusterEndpoint to return Available
func WaitDBClusterEndpointAvailable(ctx context.Context, conn *neptune.Neptune, id string) (*neptune.DBClusterEndpoint, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"creating", "modifying"},
		Target:  []string{"available"},
		Refresh: StatusDBClusterEndpoint(ctx, conn, id),
		Timeout: DBClusterEndpointAvailableTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*neptune.DBClusterEndpoint); ok {
		return v, err
	}

	return nil, err
}

// WaitDBClusterEndpointDeleted waits for a DBClusterEndpoint to return Deleted
func WaitDBClusterEndpointDeleted(ctx context.Context, conn *neptune.Neptune, id string) (*neptune.DBClusterEndpoint, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"deleting"},
		Target:  []string{},
		Refresh: StatusDBClusterEndpoint(ctx, conn, id),
		Timeout: DBClusterEndpointDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*neptune.DBClusterEndpoint); ok {
		return v, err
	}

	return nil, err
}

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

func WaitForGlobalClusterDeletion(ctx context.Context, conn *neptune.Neptune, globalClusterID string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending:        []string{GlobalClusterStatusAvailable, GlobalClusterStatusDeleting},
		Target:         []string{GlobalClusterStatusDeleted},
		Refresh:        statusGlobalClusterRefreshFunc(ctx, conn, globalClusterID),
		Timeout:        timeout,
		NotFoundChecks: 1,
	}

	log.Printf("[DEBUG] Waiting for Neptune Global Cluster (%s) deletion", globalClusterID)
	_, err := stateConf.WaitForStateContext(ctx)

	if tfresource.NotFound(err) {
		return nil
	}

	return err
}
