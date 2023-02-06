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
