package emr

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/emr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	ClusterCreatedTimeout    = 75 * time.Minute
	ClusterCreatedMinTimeout = 10 * time.Second
	ClusterCreatedDelay      = 30 * time.Second

	ClusterDeletedTimeout    = 20 * time.Minute
	ClusterDeletedMinTimeout = 10 * time.Second
	ClusterDeletedDelay      = 30 * time.Second
)

func waitClusterCreated(conn *emr.EMR, id string) (*emr.Cluster, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{emr.ClusterStateBootstrapping, emr.ClusterStateStarting},
		Target:     []string{emr.ClusterStateRunning, emr.ClusterStateWaiting},
		Refresh:    statusCluster(conn, id),
		Timeout:    ClusterCreatedTimeout,
		MinTimeout: ClusterCreatedMinTimeout,
		Delay:      ClusterCreatedDelay,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*emr.Cluster); ok {
		if stateChangeReason := output.Status.StateChangeReason; stateChangeReason != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(stateChangeReason.Code), aws.StringValue(stateChangeReason.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitClusterDeleted(conn *emr.EMR, id string) (*emr.Cluster, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{emr.ClusterStateTerminating},
		Target:     []string{emr.ClusterStateTerminated, emr.ClusterStateTerminatedWithErrors},
		Refresh:    statusCluster(conn, id),
		Timeout:    ClusterDeletedTimeout,
		MinTimeout: ClusterDeletedMinTimeout,
		Delay:      ClusterDeletedDelay,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*emr.Cluster); ok {
		if stateChangeReason := output.Status.StateChangeReason; stateChangeReason != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(stateChangeReason.Code), aws.StringValue(stateChangeReason.Message)))
		}

		return output, err
	}

	return nil, err
}
