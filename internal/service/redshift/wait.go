package redshift

import (
	"time"

	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	clusterInvalidClusterStateFaultTimeout = 15 * time.Minute
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
