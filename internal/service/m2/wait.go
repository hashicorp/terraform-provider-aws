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
