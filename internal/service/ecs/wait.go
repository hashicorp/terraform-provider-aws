package ecs

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	capacityProviderDeleteTimeout = 20 * time.Minute
	capacityProviderUpdateTimeout = 10 * time.Minute

	serviceCreateTimeout      = 2 * time.Minute
	serviceInactiveTimeoutMin = 1 * time.Second
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

func waitCapacityProviderDeleted(ctx context.Context, conn *ecs.ECS, arn string) (*ecs.CapacityProvider, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ecs.CapacityProviderStatusActive},
		Target:  []string{},
		Refresh: statusCapacityProvider(ctx, conn, arn),
		Timeout: capacityProviderDeleteTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*ecs.CapacityProvider); ok {
		return v, err
	}

	return nil, err
}

func waitCapacityProviderUpdated(ctx context.Context, conn *ecs.ECS, arn string) (*ecs.CapacityProvider, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ecs.CapacityProviderUpdateStatusUpdateInProgress},
		Target:  []string{ecs.CapacityProviderUpdateStatusUpdateComplete},
		Refresh: statusCapacityProviderUpdate(ctx, conn, arn),
		Timeout: capacityProviderUpdateTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*ecs.CapacityProvider); ok {
		return v, err
	}

	return nil, err
}

// waitServiceStable waits for an ECS Service to reach the status "ACTIVE" and have all desired tasks running. Does not return tags.
func waitServiceStable(ctx context.Context, conn *ecs.ECS, id, cluster string, timeout time.Duration) (*ecs.Service, error) { //nolint:unparam
	input := &ecs.DescribeServicesInput{
		Services: aws.StringSlice([]string{id}),
	}

	if cluster != "" {
		input.Cluster = aws.String(cluster)
	}

	stateConf := &resource.StateChangeConf{
		Pending: []string{serviceStatusInactive, serviceStatusDraining, serviceStatusPending},
		Target:  []string{serviceStatusStable},
		Refresh: statusServiceWaitForStable(ctx, conn, id, cluster),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*ecs.Service); ok {
		return v, err
	}

	return nil, err
}

// waitServiceInactive waits for an ECS Service to reach the status "INACTIVE".
func waitServiceInactive(ctx context.Context, conn *ecs.ECS, id, cluster string, timeout time.Duration) error {
	input := &ecs.DescribeServicesInput{
		Services: aws.StringSlice([]string{id}),
	}

	if cluster != "" {
		input.Cluster = aws.String(cluster)
	}

	stateConf := &resource.StateChangeConf{
		Pending:    []string{serviceStatusActive, serviceStatusDraining},
		Target:     []string{serviceStatusInactive},
		Refresh:    statusServiceNoTags(ctx, conn, id, cluster),
		Timeout:    timeout,
		MinTimeout: serviceInactiveTimeoutMin,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

// waitServiceActive waits for an ECS Service to reach the status "ACTIVE". Does not return tags.
func waitServiceActive(ctx context.Context, conn *ecs.ECS, id, cluster string, timeout time.Duration) (*ecs.Service, error) { //nolint:unparam
	stateConf := &resource.StateChangeConf{
		Pending: []string{serviceStatusInactive, serviceStatusDraining},
		Target:  []string{serviceStatusActive},
		Refresh: statusServiceNoTags(ctx, conn, id, cluster),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*ecs.Service); ok {
		return v, err
	}

	return nil, err
}

func waitClusterAvailable(ctx context.Context, conn *ecs.ECS, arn string) (*ecs.Cluster, error) { //nolint:unparam
	stateConf := &resource.StateChangeConf{
		Pending: []string{"PROVISIONING"},
		Target:  []string{"ACTIVE"},
		Refresh: statusCluster(ctx, conn, arn),
		Timeout: clusterAvailableTimeout,
		Delay:   clusterAvailableDelay,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*ecs.Cluster); ok {
		return v, err
	}

	return nil, err
}

func waitClusterDeleted(ctx context.Context, conn *ecs.ECS, arn string) (*ecs.Cluster, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{"ACTIVE", "DEPROVISIONING"},
		Target:  []string{"INACTIVE"},
		Refresh: statusCluster(ctx, conn, arn),
		Timeout: clusterDeleteTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*ecs.Cluster); ok {
		return v, err
	}

	return nil, err
}

func waitTaskSetStable(ctx context.Context, conn *ecs.ECS, timeout time.Duration, taskSetID, service, cluster string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{ecs.StabilityStatusStabilizing},
		Target:  []string{ecs.StabilityStatusSteadyState},
		Refresh: stabilityStatusTaskSet(ctx, conn, taskSetID, service, cluster),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitTaskSetDeleted(ctx context.Context, conn *ecs.ECS, taskSetID, service, cluster string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{taskSetStatusActive, taskSetStatusPrimary, taskSetStatusDraining},
		Target:  []string{},
		Refresh: statusTaskSet(ctx, conn, taskSetID, service, cluster),
		Timeout: taskSetDeleteTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}
