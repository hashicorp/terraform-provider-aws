package evidently

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/cloudwatchevidently"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func waitFeatureCreated(ctx context.Context, conn *cloudwatchevidently.CloudWatchEvidently, id string, timeout time.Duration) (*cloudwatchevidently.Feature, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{},
		Target:  []string{cloudwatchevidently.FeatureStatusAvailable},
		Refresh: statusFeature(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*cloudwatchevidently.Feature); ok {
		return output, err
	}

	return nil, err
}

func waitFeatureUpdated(ctx context.Context, conn *cloudwatchevidently.CloudWatchEvidently, id string, timeout time.Duration) (*cloudwatchevidently.Feature, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{cloudwatchevidently.FeatureStatusUpdating},
		Target:  []string{cloudwatchevidently.FeatureStatusAvailable},
		Refresh: statusFeature(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*cloudwatchevidently.Feature); ok {
		return output, err
	}

	return nil, err
}

func waitFeatureDeleted(ctx context.Context, conn *cloudwatchevidently.CloudWatchEvidently, id string, timeout time.Duration) (*cloudwatchevidently.Feature, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{cloudwatchevidently.FeatureStatusAvailable},
		Target:  []string{},
		Refresh: statusFeature(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*cloudwatchevidently.Feature); ok {
		return output, err
	}

	return nil, err
}

func waitLaunchCreated(ctx context.Context, conn *cloudwatchevidently.CloudWatchEvidently, id string, timeout time.Duration) (*cloudwatchevidently.Launch, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{},
		Target:  []string{cloudwatchevidently.LaunchStatusCreated},
		Refresh: statusLaunch(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(context.Background())

	if output, ok := outputRaw.(*cloudwatchevidently.Launch); ok {
		return output, err
	}

	return nil, err
}

func waitLaunchUpdated(ctx context.Context, conn *cloudwatchevidently.CloudWatchEvidently, id string, timeout time.Duration) (*cloudwatchevidently.Launch, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{cloudwatchevidently.LaunchStatusUpdating},
		Target:  []string{cloudwatchevidently.LaunchStatusCreated, cloudwatchevidently.LaunchStatusRunning},
		Refresh: statusLaunch(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(context.Background())

	if output, ok := outputRaw.(*cloudwatchevidently.Launch); ok {
		return output, err
	}

	return nil, err
}

func waitLaunchDeleted(ctx context.Context, conn *cloudwatchevidently.CloudWatchEvidently, id string, timeout time.Duration) (*cloudwatchevidently.Launch, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{cloudwatchevidently.LaunchStatusCreated, cloudwatchevidently.LaunchStatusCompleted, cloudwatchevidently.LaunchStatusRunning, cloudwatchevidently.LaunchStatusCancelled, cloudwatchevidently.LaunchStatusUpdating},
		Target:  []string{},
		Refresh: statusLaunch(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(context.Background())

	if output, ok := outputRaw.(*cloudwatchevidently.Launch); ok {
		return output, err
	}

	return nil, err
}

func waitProjectCreated(ctx context.Context, conn *cloudwatchevidently.CloudWatchEvidently, nameOrARN string, timeout time.Duration) (*cloudwatchevidently.Project, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{},
		Target:  []string{cloudwatchevidently.ProjectStatusAvailable},
		Refresh: statusProject(ctx, conn, nameOrARN),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*cloudwatchevidently.Project); ok {
		return output, err
	}

	return nil, err
}

func waitProjectUpdated(ctx context.Context, conn *cloudwatchevidently.CloudWatchEvidently, nameOrARN string, timeout time.Duration) (*cloudwatchevidently.Project, error) { //nolint:unparam
	stateConf := &resource.StateChangeConf{
		Pending: []string{cloudwatchevidently.ProjectStatusUpdating},
		Target:  []string{cloudwatchevidently.ProjectStatusAvailable},
		Refresh: statusProject(ctx, conn, nameOrARN),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*cloudwatchevidently.Project); ok {
		return output, err
	}

	return nil, err
}

func waitProjectDeleted(ctx context.Context, conn *cloudwatchevidently.CloudWatchEvidently, nameOrARN string, timeout time.Duration) (*cloudwatchevidently.Project, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{cloudwatchevidently.ProjectStatusAvailable},
		Target:  []string{},
		Refresh: statusProject(ctx, conn, nameOrARN),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*cloudwatchevidently.Project); ok {
		return output, err
	}

	return nil, err
}
