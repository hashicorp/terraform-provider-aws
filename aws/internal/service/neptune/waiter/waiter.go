package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
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
func WaitEventSubscriptionDeleted(conn *neptune.Neptune, subscriptionName string) (*neptune.EventSubscription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"deleting"},
		Target:  []string{EventSubscriptionStatusNotFound},
		Refresh: StatusEventSubscription(conn, subscriptionName),
		Timeout: EventSubscriptionDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*neptune.EventSubscription); ok {
		return v, err
	}

	return nil, err
}

// WaitDBClusterDeleted waits for a Cluster to return Deleted
func WaitDBClusterDeleted(conn *neptune.Neptune, id string, timeout time.Duration) (*neptune.DBCluster, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			"available",
			"deleting",
			"backing-up",
			"modifying",
		},
		Target:     []string{ClusterStatusNotFound},
		Refresh:    StatusCluster(conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*neptune.DBCluster); ok {
		return v, err
	}

	return nil, err
}

// WaitDBClusterAvailable waits for a Cluster to return Available
func WaitDBClusterAvailable(conn *neptune.Neptune, id string, timeout time.Duration) (*neptune.DBCluster, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			"creating",
			"backing-up",
			"modifying",
			"preparing-data-migration",
			"migrating",
			"configuring-iam-database-auth",
		},
		Target:     []string{"available"},
		Refresh:    StatusCluster(conn, id),
		Timeout:    timeout,
		MinTimeout: 10 * time.Second,
		Delay:      30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*neptune.DBCluster); ok {
		return v, err
	}

	return nil, err
}

// WaitDBClusterEndpointAvailable waits for a DBClusterEndpoint to return Available
func WaitDBClusterEndpointAvailable(conn *neptune.Neptune, id string) (*neptune.DBClusterEndpoint, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"creating", "modifying"},
		Target:  []string{"available"},
		Refresh: StatusDBClusterEndpoint(conn, id),
		Timeout: DBClusterEndpointAvailableTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*neptune.DBClusterEndpoint); ok {
		return v, err
	}

	return nil, err
}

// WaitDBClusterEndpointDeleted waits for a DBClusterEndpoint to return Deleted
func WaitDBClusterEndpointDeleted(conn *neptune.Neptune, id string) (*neptune.DBClusterEndpoint, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"deleting"},
		Target:  []string{},
		Refresh: StatusDBClusterEndpoint(conn, id),
		Timeout: DBClusterEndpointDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*neptune.DBClusterEndpoint); ok {
		return v, err
	}

	return nil, err
}
