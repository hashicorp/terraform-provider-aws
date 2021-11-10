package emr

import (
	"time"

	emr "github.com/aws/aws-sdk-go/service/emr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	ClusterCreatedTimeout    = 75 * time.Minute
	ClusterCreatedMinTimeout = 10 * time.Second
	ClusterCreatedDelay      = 30 * time.Second
)

func waitClusterCreated(conn *emr.EMR, id string) (*emr.Cluster, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			emr.ClusterStateBootstrapping,
			emr.ClusterStateStarting,
		},
		Target: []string{
			emr.ClusterStateRunning,
			emr.ClusterStateWaiting,
		},
		Refresh:    statusCluster(conn, id),
		Timeout:    ClusterCreatedTimeout,
		MinTimeout: ClusterCreatedMinTimeout,
		Delay:      ClusterCreatedDelay,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*emr.Cluster); ok {
		return output, err
	}

	return nil, err
}
