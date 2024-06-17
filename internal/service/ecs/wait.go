// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ecs/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
)

const (
	capacityProviderDeleteTimeout = 20 * time.Minute
	capacityProviderUpdateTimeout = 10 * time.Minute

	serviceCreateTimeout      = 2 * time.Minute
	serviceInactiveMinTimeout = 1 * time.Second
	serviceDescribeTimeout    = 2 * time.Minute
	serviceUpdateTimeout      = 2 * time.Minute

	clusterAvailableDelay   = 10 * time.Second
	clusterAvailableTimeout = 10 * time.Minute
	clusterDeleteTimeout    = 10 * time.Minute
	clusterReadTimeout      = 2 * time.Second
	clusterUpdateTimeout    = 10 * time.Minute

	taskSetCreateTimeout = 10 * time.Minute
	taskSetDeleteTimeout = 10 * time.Minute
)

func waitCapacityProviderDeleted(ctx context.Context, conn *ecs.Client, partition, arn string) (*awstypes.CapacityProvider, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.CapacityProviderStatusActive),
		Target:  []string{},
		Refresh: statusCapacityProvider(ctx, conn, partition, arn),
		Timeout: capacityProviderDeleteTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*awstypes.CapacityProvider); ok {
		return v, err
	}

	return nil, err
}

func waitCapacityProviderUpdated(ctx context.Context, conn *ecs.Client, partition, arn string) (*awstypes.CapacityProvider, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.CapacityProviderUpdateStatusUpdateInProgress),
		Target:  enum.Slice(awstypes.CapacityProviderUpdateStatusUpdateComplete),
		Refresh: statusCapacityProviderUpdate(ctx, conn, partition, arn),
		Timeout: capacityProviderUpdateTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*awstypes.CapacityProvider); ok {
		return v, err
	}

	return nil, err
}

// waitServiceStable waits for an ECS Service to reach the status "ACTIVE" and have all desired tasks running. Does not return tags.
func waitServiceStable(ctx context.Context, conn *ecs.Client, partition, id, cluster string, timeout time.Duration) (*awstypes.Service, error) {
	input := &ecs.DescribeServicesInput{
		Services: []string{id},
	}

	if cluster != "" {
		input.Cluster = aws.String(cluster)
	}

	stateConf := &retry.StateChangeConf{
		Pending: []string{serviceStatusInactive, serviceStatusDraining, serviceStatusPending},
		Target:  []string{serviceStatusStable},
		Refresh: statusServiceWaitForStable(ctx, conn, partition, id, cluster),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*awstypes.Service); ok {
		return v, err
	}

	return nil, err
}

// waitServiceInactive waits for an ECS Service to reach the status "INACTIVE".
func waitServiceInactive(ctx context.Context, conn *ecs.Client, partition, id, cluster string, timeout time.Duration) error {
	input := &ecs.DescribeServicesInput{
		Services: []string{id},
	}

	if cluster != "" {
		input.Cluster = aws.String(cluster)
	}

	stateConf := &retry.StateChangeConf{
		Pending:    []string{serviceStatusActive, serviceStatusDraining},
		Target:     []string{serviceStatusInactive},
		Refresh:    statusServiceNoTags(ctx, conn, partition, id, cluster),
		Timeout:    timeout,
		MinTimeout: serviceInactiveMinTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

// waitServiceActive waits for an ECS Service to reach the status "ACTIVE". Does not return tags.
func waitServiceActive(ctx context.Context, conn *ecs.Client, partition, id, cluster string, timeout time.Duration) (*awstypes.Service, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{serviceStatusInactive, serviceStatusDraining},
		Target:  []string{serviceStatusActive},
		Refresh: statusServiceNoTags(ctx, conn, partition, id, cluster),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*awstypes.Service); ok {
		return v, err
	}

	return nil, err
}

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
