package ecs

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	serviceStatusInactive = "INACTIVE"
	serviceStatusActive   = "ACTIVE"
	serviceStatusDraining = "DRAINING"
	// Non-standard statuses for statusServiceWaitForStable()
	serviceStatusPending = "tfPENDING"
	serviceStatusStable  = "tfSTABLE"

	clusterStatusError = "ERROR"
	clusterStatusNone  = "NONE"

	taskSetStatusActive   = "ACTIVE"
	taskSetStatusDraining = "DRAINING"
	taskSetStatusPrimary  = "PRIMARY"
)

func statusCapacityProvider(ctx context.Context, conn *ecs.ECS, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindCapacityProviderByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}

func statusCapacityProviderUpdate(ctx context.Context, conn *ecs.ECS, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindCapacityProviderByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.UpdateStatus), nil
	}
}

func statusServiceNoTags(ctx context.Context, conn *ecs.ECS, id, cluster string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		service, err := FindServiceNoTagsByID(ctx, conn, id, cluster)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}
		if err != nil {
			return nil, "", err
		}

		return service, aws.StringValue(service.Status), err
	}
}

func statusServiceWaitForStable(ctx context.Context, conn *ecs.ECS, id, cluster string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		serviceRaw, status, err := statusServiceNoTags(ctx, conn, id, cluster)()
		if err != nil {
			return nil, "", err
		}

		if status != serviceStatusActive {
			return serviceRaw, status, nil
		}

		service := serviceRaw.(*ecs.Service)

		if d, dc, rc := len(service.Deployments),
			aws.Int64Value(service.DesiredCount),
			aws.Int64Value(service.RunningCount); d == 1 && dc == rc {
			status = serviceStatusStable
		} else {
			status = serviceStatusPending
		}

		return service, status, nil
	}
}

func statusCluster(ctx context.Context, conn *ecs.ECS, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		cluster, err := FindClusterByNameOrARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, clusterStatusNone, nil
		}

		if err != nil {
			return nil, clusterStatusError, err
		}

		return cluster, aws.StringValue(cluster.Status), err
	}
}

func stabilityStatusTaskSet(ctx context.Context, conn *ecs.ECS, taskSetID, service, cluster string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &ecs.DescribeTaskSetsInput{
			Cluster:  aws.String(cluster),
			Service:  aws.String(service),
			TaskSets: aws.StringSlice([]string{taskSetID}),
		}

		output, err := conn.DescribeTaskSetsWithContext(ctx, input)

		if err != nil {
			return nil, "", err
		}

		if output == nil || len(output.TaskSets) == 0 {
			return nil, "", nil
		}

		return output.TaskSets[0], aws.StringValue(output.TaskSets[0].StabilityStatus), nil
	}
}

func statusTaskSet(ctx context.Context, conn *ecs.ECS, taskSetID, service, cluster string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &ecs.DescribeTaskSetsInput{
			Cluster:  aws.String(cluster),
			Service:  aws.String(service),
			TaskSets: aws.StringSlice([]string{taskSetID}),
		}

		output, err := conn.DescribeTaskSetsWithContext(ctx, input)

		if err != nil {
			return nil, "", err
		}

		if output == nil || len(output.TaskSets) == 0 {
			return nil, "", nil
		}

		return output.TaskSets[0], aws.StringValue(output.TaskSets[0].Status), nil
	}
}
