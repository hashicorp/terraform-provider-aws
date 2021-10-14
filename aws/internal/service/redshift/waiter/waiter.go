package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	tfredshift "github.com/hashicorp/terraform-provider-aws/aws/internal/service/redshift"
)

const (
	ClusterInvalidClusterStateFaultTimeout = 15 * time.Minute
)

func ClusterDeleted(conn *redshift.Redshift, id string, timeout time.Duration) (*redshift.Cluster, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			tfredshift.ClusterStatusAvailable,
			tfredshift.ClusterStatusCreating,
			tfredshift.ClusterStatusDeleting,
			tfredshift.ClusterStatusFinalSnapshot,
			tfredshift.ClusterStatusRebooting,
			tfredshift.ClusterStatusRenaming,
			tfredshift.ClusterStatusResizing,
		},
		Target:  []string{},
		Refresh: ClusterStatus(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*redshift.Cluster); ok {
		return output, err
	}

	return nil, err
}
