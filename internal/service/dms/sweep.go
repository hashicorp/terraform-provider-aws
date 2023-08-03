// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package dms

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	dms "github.com/aws/aws-sdk-go/service/databasemigrationservice"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_dms_replication_instance", &resource.Sweeper{
		Name: "aws_dms_replication_instance",
		F:    sweepReplicationInstances,
		Dependencies: []string{
			"aws_dms_replication_task",
		},
	})

	resource.AddTestSweepers("aws_dms_replication_task", &resource.Sweeper{
		Name: "aws_dms_replication_task",
		F:    sweepReplicationTasks,
	})

	resource.AddTestSweepers("aws_dms_endpoint", &resource.Sweeper{
		Name: "aws_dms_endpoint",
		F:    sweepEndpoints,
	})
}

func sweepReplicationInstances(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.DMSConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	err = conn.DescribeReplicationInstancesPagesWithContext(ctx, &dms.DescribeReplicationInstancesInput{}, func(page *dms.DescribeReplicationInstancesOutput, lastPage bool) bool {
		for _, instance := range page.ReplicationInstances {
			r := ResourceReplicationInstance()
			d := r.Data(nil)
			d.Set("replication_instance_arn", instance.ReplicationInstanceArn)
			d.SetId(aws.StringValue(instance.ReplicationInstanceIdentifier))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error describing DMS Replication Instances: %w", err))
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping DMS Replication Instances for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping DMS Replication Instance sweep for %s: %s", region, err)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepReplicationTasks(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.DMSConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	input := &dms.DescribeReplicationTasksInput{
		WithoutSettings: aws.Bool(true),
	}
	err = conn.DescribeReplicationTasksPagesWithContext(ctx, input, func(page *dms.DescribeReplicationTasksOutput, lastPage bool) bool {
		for _, instance := range page.ReplicationTasks {
			r := ResourceReplicationTask()
			d := r.Data(nil)
			d.SetId(aws.StringValue(instance.ReplicationTaskIdentifier))
			d.Set("replication_task_arn", instance.ReplicationTaskArn)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error describing DMS Replication Tasks: %w", err))
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping DMS Replication Tasks for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping DMS Replication Instance sweep for %s: %s", region, err)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepEndpoints(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.DMSConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error

	err = conn.DescribeEndpointsPagesWithContext(ctx, &dms.DescribeEndpointsInput{}, func(page *dms.DescribeEndpointsOutput, lastPage bool) bool {
		for _, ep := range page.Endpoints {
			r := ResourceEndpoint()
			d := r.Data(nil)
			d.Set("endpoint_arn", ep.EndpointArn)
			d.SetId(aws.StringValue(ep.EndpointIdentifier))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error describing DMS Endpoints: %w", err))
	}

	if err = sweep.SweepOrchestrator(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping DMS Endpoints for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping DMS Endpoint sweep for %s: %s", region, err)
		return nil
	}

	return errs.ErrorOrNil()
}
