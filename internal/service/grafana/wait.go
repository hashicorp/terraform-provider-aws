package grafana

import (
	"fmt"
	"time"
)

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/managedgrafana"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func waitWorkspaceCreated(conn *managedgrafana.ManagedGrafana, id string, timeout time.Duration) (*managedgrafana.DescribeWorkspaceOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{managedgrafana.WorkspaceStatusCreating},
		Target:  []string{managedgrafana.WorkspaceStatusActive},
		Refresh: statusWorkspaceStatus(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*managedgrafana.DescribeWorkspaceOutput); ok {
		return output, err
	}

	return nil, err
}

func waitWorkspaceUpdated(conn *managedgrafana.ManagedGrafana, id string, timeout time.Duration) (interface{}, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{managedgrafana.WorkspaceStatusUpdating},
		Target:  []string{managedgrafana.WorkspaceStatusActive},
		Refresh: statusWorkspaceStatus(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*managedgrafana.DescribeWorkspaceOutput); ok {
		return output, err
	}

	return nil, err
}

func waitWorkspaceDeleted(conn *managedgrafana.ManagedGrafana, id string, timeout time.Duration) (interface{}, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{managedgrafana.WorkspaceStatusDeleting},
		Target:  []string{},
		Refresh: statusWorkspaceStatus(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*managedgrafana.WorkspaceDescription); ok {
		if status := aws.StringValue(output.Status); status == managedgrafana.WorkspaceStatusFailed {
			tfresource.SetLastError(err, fmt.Errorf("%s", status))
		}

		return output, err
	}

	return nil, err
}
