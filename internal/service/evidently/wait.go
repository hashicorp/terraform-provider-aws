package evidently

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/cloudwatchevidently"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func waitFeatureCreated(conn *cloudwatchevidently.CloudWatchEvidently, id string, timeout time.Duration) (*cloudwatchevidently.Feature, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{},
		Target:  []string{cloudwatchevidently.FeatureStatusAvailable},
		Refresh: statusFeature(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(context.Background())

	if output, ok := outputRaw.(*cloudwatchevidently.Feature); ok {
		return output, err
	}

	return nil, err
}

func waitFeatureUpdated(conn *cloudwatchevidently.CloudWatchEvidently, id string, timeout time.Duration) (*cloudwatchevidently.Feature, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{cloudwatchevidently.FeatureStatusUpdating},
		Target:  []string{cloudwatchevidently.FeatureStatusAvailable},
		Refresh: statusFeature(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(context.Background())

	if output, ok := outputRaw.(*cloudwatchevidently.Feature); ok {
		return output, err
	}

	return nil, err
}

func waitFeatureDeleted(conn *cloudwatchevidently.CloudWatchEvidently, id string, timeout time.Duration) (*cloudwatchevidently.Feature, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{cloudwatchevidently.FeatureStatusAvailable},
		Target:  []string{},
		Refresh: statusFeature(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(context.Background())

	if output, ok := outputRaw.(*cloudwatchevidently.Feature); ok {
		return output, err
	}

	return nil, err
}

func waitProjectCreated(conn *cloudwatchevidently.CloudWatchEvidently, nameOrARN string, timeout time.Duration) (*cloudwatchevidently.Project, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{},
		Target:  []string{cloudwatchevidently.ProjectStatusAvailable},
		Refresh: statusProject(conn, nameOrARN),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(context.Background())

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

	outputRaw, err := stateConf.WaitForStateContext(context.Background())

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

	outputRaw, err := stateConf.WaitForStateContext(context.Background())

	if output, ok := outputRaw.(*cloudwatchevidently.Project); ok {
		return output, err
	}

	return nil, err
}
