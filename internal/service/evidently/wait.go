package evidently

import (
	"time"

	"github.com/aws/aws-sdk-go/service/cloudwatchevidently"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func waitProjectCreated(conn *cloudwatchevidently.CloudWatchEvidently, nameOrARN string, timeout time.Duration) (*cloudwatchevidently.Project, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{},
		Target:  []string{cloudwatchevidently.ProjectStatusAvailable},
		Refresh: statusProject(conn, nameOrARN),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*cloudwatchevidently.Project); ok {
		return output, err
	}

	return nil, err
}

func waitProjectUpdated(conn *cloudwatchevidently.CloudWatchEvidently, nameOrARN string, timeout time.Duration) (*cloudwatchevidently.Project, error) { //nolint:unparam
	stateConf := &resource.StateChangeConf{
		Pending: []string{cloudwatchevidently.ProjectStatusUpdating},
		Target:  []string{cloudwatchevidently.ProjectStatusAvailable},
		Refresh: statusProject(conn, nameOrARN),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*cloudwatchevidently.Project); ok {
		return output, err
	}

	return nil, err
}

func waitProjectDeleted(conn *cloudwatchevidently.CloudWatchEvidently, nameOrARN string, timeout time.Duration) (*cloudwatchevidently.Project, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{cloudwatchevidently.ProjectStatusAvailable},
		Target:  []string{},
		Refresh: statusProject(conn, nameOrARN),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*cloudwatchevidently.Project); ok {
		return output, err
	}

	return nil, err
}
