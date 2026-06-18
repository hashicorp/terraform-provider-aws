// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package evidently

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/evidently"
	awstypes "github.com/aws/aws-sdk-go-v2/service/evidently/types"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
)

func waitFeatureCreated(ctx context.Context, conn *evidently.Client, id string, timeout time.Duration) (*awstypes.Feature, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{},
		Target:  enum.Slice(awstypes.FeatureStatusAvailable),
		Refresh: statusFeature(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Feature); ok {
		return output, err
	}

	return nil, err
}

func waitFeatureUpdated(ctx context.Context, conn *evidently.Client, id string, timeout time.Duration) (*awstypes.Feature, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.FeatureStatusUpdating),
		Target:  enum.Slice(awstypes.FeatureStatusAvailable),
		Refresh: statusFeature(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Feature); ok {
		return output, err
	}

	return nil, err
}

func waitFeatureDeleted(ctx context.Context, conn *evidently.Client, id string, timeout time.Duration) (*awstypes.Feature, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.FeatureStatusAvailable),
		Target:  []string{},
		Refresh: statusFeature(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Feature); ok {
		return output, err
	}

	return nil, err
}

func waitLaunchCreated(ctx context.Context, conn *evidently.Client, id string, timeout time.Duration) (*awstypes.Launch, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{},
		Target:  enum.Slice(awstypes.LaunchStatusCreated),
		Refresh: statusLaunch(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Launch); ok {
		if v := aws.ToString(output.StatusReason); v != "" {
			retry.SetLastError(err, errors.New(v))
		}

		return output, err
	}

	return nil, err
}

func waitLaunchUpdated(ctx context.Context, conn *evidently.Client, id string, timeout time.Duration) (*awstypes.Launch, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.LaunchStatusUpdating),
		Target:  enum.Slice(awstypes.LaunchStatusCreated, awstypes.LaunchStatusRunning),
		Refresh: statusLaunch(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Launch); ok {
		if v := aws.ToString(output.StatusReason); v != "" {
			retry.SetLastError(err, errors.New(v))
		}

		return output, err
	}

	return nil, err
}

func waitLaunchDeleted(ctx context.Context, conn *evidently.Client, id string, timeout time.Duration) (*awstypes.Launch, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.LaunchStatusCreated, awstypes.LaunchStatusCompleted, awstypes.LaunchStatusRunning, awstypes.LaunchStatusCancelled, awstypes.LaunchStatusUpdating),
		Target:  []string{},
		Refresh: statusLaunch(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Launch); ok {
		if v := aws.ToString(output.StatusReason); v != "" {
			retry.SetLastError(err, errors.New(v))
		}

		return output, err
	}

	return nil, err
}

func waitProjectCreated(ctx context.Context, conn *evidently.Client, nameOrARN string, timeout time.Duration) (*awstypes.Project, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{},
		Target:  enum.Slice(awstypes.ProjectStatusAvailable),
		Refresh: statusProject(conn, nameOrARN),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Project); ok {
		return output, err
	}

	return nil, err
}

func waitProjectUpdated(ctx context.Context, conn *evidently.Client, nameOrARN string, timeout time.Duration) (*awstypes.Project, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ProjectStatusUpdating),
		Target:  enum.Slice(awstypes.ProjectStatusAvailable),
		Refresh: statusProject(conn, nameOrARN),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Project); ok {
		return output, err
	}

	return nil, err
}

func waitProjectDeleted(ctx context.Context, conn *evidently.Client, nameOrARN string, timeout time.Duration) (*awstypes.Project, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ProjectStatusAvailable),
		Target:  []string{},
		Refresh: statusProject(conn, nameOrARN),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Project); ok {
		return output, err
	}

	return nil, err
}
