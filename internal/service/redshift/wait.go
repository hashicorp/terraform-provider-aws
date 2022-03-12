package redshift

import (
	"time"

	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	clusterInvalidClusterStateFaultTimeout = 15 * time.Minute

	clusterRelocationStatusResolvedTimeout = 1 * time.Minute
)

func waitClusterDeleted(conn *redshift.Redshift, id string, timeout time.Duration) (*redshift.Cluster, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			clusterStatusAvailable,
			clusterStatusCreating,
			clusterStatusDeleting,
			clusterStatusFinalSnapshot,
			clusterStatusRebooting,
			clusterStatusRenaming,
			clusterStatusResizing,
		},
		Target:  []string{},
		Refresh: statusCluster(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*redshift.Cluster); ok {
		return output, err
	}

	return nil, err
}

func waitClusterRelocationStatusResolved(conn *redshift.Redshift, id string) (*redshift.Cluster, error) { //nolint:unparam
	stateConf := &resource.StateChangeConf{
		Pending: clusterAvailabilityZoneRelocationStatus_PendingValues(),
		Target:  clusterAvailabilityZoneRelocationStatus_TerminalValues(),
		Refresh: statusClusterAvailabilityZoneRelocation(conn, id),
		Timeout: clusterRelocationStatusResolvedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*redshift.Cluster); ok {
		return output, err
	}

	return nil, err
}
