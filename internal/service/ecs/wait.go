// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
)

const (
	taskSetCreateTimeout = 10 * time.Minute
	taskSetDeleteTimeout = 10 * time.Minute
)

func waitTaskSetStable(ctx context.Context, conn *ecs.Client, timeout time.Duration, taskSetID, service, cluster string) error {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.StabilityStatusStabilizing),
		Target:  enum.Slice(awstypes.StabilityStatusSteadyState),
		Refresh: stabilityStatusTaskSet(ctx, conn, taskSetID, service, cluster),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitTaskSetDeleted(ctx context.Context, conn *ecs.Client, taskSetID, service, cluster string) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{taskSetStatusActive, taskSetStatusPrimary, taskSetStatusDraining},
		Target:  []string{},
		Refresh: statusTaskSet(ctx, conn, taskSetID, service, cluster),
		Timeout: taskSetDeleteTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}
