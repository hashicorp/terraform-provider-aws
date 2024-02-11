// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package m2

import (
	"context"
	"time"

	m2 "github.com/aws/aws-sdk-go-v2/service/m2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/m2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

// waitM2 Environment available
func waitEnvironmentAvailable(ctx context.Context, conn *m2.Client, id string, timeout time.Duration) (*m2.GetEnvironmentOutput, error) {
	stateConf := &retry.StateChangeConf{
		//Pending:                   enum.Slice(awstypes.EnvironmentLifecycleCreating),
		Pending:                   enum.Slice(awstypes.EnvironmentLifecycleCreating, awstypes.EnvironmentLifecycleUpdating),
		Target:                    enum.Slice(awstypes.EnvironmentLifecycleAvailable),
		Refresh:                   statusEnvironment(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*m2.GetEnvironmentOutput); ok {
		return out, err
	}

	return nil, err
}

// M2 Environment Delete
const (
	EnvironmentDeletedMinTimeout = 10 * time.Second
	EnvironmentDeletedDelay      = 30 * time.Second
)

func waitEnvironmentDeleted(ctx context.Context, conn *m2.Client, id string, timeout time.Duration) (*m2.GetEnvironmentOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    enum.Slice(awstypes.EnvironmentLifecycleCreating, awstypes.EnvironmentLifecycleDeleting, awstypes.EnvironmentLifecycleUpdating),
		Target:     []string{},
		Refresh:    statusEnvironment(ctx, conn, id),
		Timeout:    timeout,
		MinTimeout: EnvironmentDeletedMinTimeout,
		Delay:      EnvironmentDeletedDelay,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*m2.GetEnvironmentOutput); ok {
		return out, err
	}

	return nil, err
}

func waitApplicationCreated(ctx context.Context, conn *m2.Client, id string, timeout time.Duration) (*m2.GetApplicationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.ApplicationLifecycleCreating),
		Target:                    enum.Slice(awstypes.ApplicationLifecycleCreated, awstypes.ApplicationLifecycleAvailable),
		Refresh:                   statusApplication(ctx, conn, id),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*m2.GetApplicationOutput); ok {
		return out, err
	}

	return nil, err
}

func waitApplicationUpdated(ctx context.Context, conn *m2.Client, id string, version int32, timeout time.Duration) (*m2.GetApplicationVersionOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending:                   enum.Slice(awstypes.ApplicationVersionLifecycleCreating),
		Target:                    enum.Slice(awstypes.ApplicationVersionLifecycleAvailable),
		Refresh:                   statusApplicationVersion(ctx, conn, id, version),
		Timeout:                   timeout,
		NotFoundChecks:            20,
		ContinuousTargetOccurence: 2,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*m2.GetApplicationVersionOutput); ok {
		return out, err
	}

	return nil, err
}

func waitApplicationDeleted(ctx context.Context, conn *m2.Client, id string, timeout time.Duration) (*m2.GetApplicationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ApplicationLifecycleDeleting, awstypes.ApplicationLifecycleDeletingFromEnvironment),
		Target:  []string{},
		Refresh: statusApplication(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if out, ok := outputRaw.(*m2.GetApplicationOutput); ok {
		return out, err
	}

	return nil, err
}

func statusApplication(ctx context.Context, conn *m2.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindAppByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}

func statusApplicationVersion(ctx context.Context, conn *m2.Client, id string, version int32) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := findApplicationVersion(ctx, conn, id, version)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return out, string(out.Status), nil
	}
}
