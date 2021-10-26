package emrcontainers

import (
	"time"

	"github.com/aws/aws-sdk-go/service/emrcontainers"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// Maximum amount of time to wait for a virtual cluster creation
	VirtualClusterCreatedTimeout = 90 * time.Minute
	// Amount of delay to check a virtual cluster
	VirtualClusterCreatedDelay = 1 * time.Minute

	// Maximum amount of time to wait for a virtual cluster deletion
	VirtualClusterDeletedTimeout = 90 * time.Minute
	// Amount of delay to check a virtual cluster status
	VirtualClusterDeletedDelay = 1 * time.Minute
)

// waitVirtualClusterCreated waits for a virtual cluster to return running
func waitVirtualClusterCreated(conn *emrcontainers.EMRContainers, id string) (*emrcontainers.VirtualCluster, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{},
		Target:  []string{emrcontainers.VirtualClusterStateRunning},
		Refresh: statusVirtualCluster(conn, id),
		Timeout: VirtualClusterCreatedTimeout,
		Delay:   VirtualClusterCreatedDelay,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*emrcontainers.VirtualCluster); ok {
		return v, err
	}

	return nil, err
}

// waitVirtualClusterDeleted waits for a virtual cluster to be deleted
func waitVirtualClusterDeleted(conn *emrcontainers.EMRContainers, id string) (*emrcontainers.VirtualCluster, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{emrcontainers.VirtualClusterStateTerminating},
		Target:  []string{emrcontainers.VirtualClusterStateTerminated},
		Refresh: statusVirtualCluster(conn, id),
		Timeout: VirtualClusterDeletedTimeout,
		Delay:   VirtualClusterDeletedDelay,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*emrcontainers.VirtualCluster); ok {
		return v, err
	}

	return nil, err
}
