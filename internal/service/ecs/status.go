// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ecs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

const (
	taskSetStatusActive   = "ACTIVE"
	taskSetStatusDraining = "DRAINING"
	taskSetStatusPrimary  = "PRIMARY"
)

func stabilityStatusTaskSet(ctx context.Context, conn *ecs.Client, taskSetID, service, cluster string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &ecs.DescribeTaskSetsInput{
			Cluster:  aws.String(cluster),
			Service:  aws.String(service),
			TaskSets: []string{taskSetID},
		}

		output, err := conn.DescribeTaskSets(ctx, input)

		if err != nil {
			return nil, "", err
		}

		if output == nil || len(output.TaskSets) == 0 {
			return nil, "", nil
		}

		return output.TaskSets[0], string(output.TaskSets[0].StabilityStatus), nil
	}
}

func statusTaskSet(ctx context.Context, conn *ecs.Client, taskSetID, service, cluster string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &ecs.DescribeTaskSetsInput{
			Cluster:  aws.String(cluster),
			Service:  aws.String(service),
			TaskSets: []string{taskSetID},
		}

		output, err := conn.DescribeTaskSets(ctx, input)

		if err != nil {
			return nil, "", err
		}

		if output == nil || len(output.TaskSets) == 0 {
			return nil, "", nil
		}

		return output.TaskSets[0], aws.ToString(output.TaskSets[0].Status), nil
	}
}
