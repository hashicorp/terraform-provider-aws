package cloud9

import (
	"time"

	"github.com/aws/aws-sdk-go/service/cloud9"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	EnvironmentReadyTimeout   = 10 * time.Minute
	EnvironmentDeletedTimeout = 20 * time.Minute
)

func waitEnvironmentReady(conn *cloud9.Cloud9, id string) (*cloud9.DescribeEnvironmentStatusOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{cloud9.EnvironmentLifecycleStatusCreating},
		Target:  []string{cloud9.EnvironmentLifecycleStatusCreated},
		Refresh: statusEnvironmentStatus(conn, id),
		Timeout: EnvironmentReadyTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*cloud9.DescribeEnvironmentStatusOutput); ok {
		return output, err
	}

	return nil, err
}

func waitEnvironmentDeleted(conn *cloud9.Cloud9, id string) (*cloud9.DescribeEnvironmentStatusOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{cloud9.EnvironmentLifecycleStatusDeleting},
		Target:  []string{},
		Refresh: statusEnvironmentStatus(conn, id),
		Timeout: EnvironmentDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*cloud9.DescribeEnvironmentStatusOutput); ok {
		return output, err
	}

	return nil, err
}
