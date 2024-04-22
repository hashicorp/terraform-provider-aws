// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dms

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	dms "github.com/aws/aws-sdk-go/service/databasemigrationservice"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv1"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_dms_endpoint", &resource.Sweeper{
		Name: "aws_dms_endpoint",
		F:    sweepEndpoints,
		Dependencies: []string{
			"aws_dms_replication_config",
		},
	})

	resource.AddTestSweepers("aws_dms_replication_config", &resource.Sweeper{
		Name: "aws_dms_replication_config",
		F:    sweepReplicationConfigs,
	})

	resource.AddTestSweepers("aws_dms_replication_instance", &resource.Sweeper{
		Name: "aws_dms_replication_instance",
		F:    sweepReplicationInstances,
		Dependencies: []string{
			"aws_dms_replication_task",
		},
	})

	resource.AddTestSweepers("aws_dms_replication_subnet_group", &resource.Sweeper{
		Name: "aws_dms_replication_subnet_group",
		F:    sweepReplicationSubnetGroups,
		Dependencies: []string{
			"aws_dms_replication_instance",
		},
	})

	resource.AddTestSweepers("aws_dms_replication_task", &resource.Sweeper{
		Name: "aws_dms_replication_task",
		F:    sweepReplicationTasks,
	})
}

func sweepEndpoints(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.DMSConn(ctx)
	input := &dms.DescribeEndpointsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.DescribeEndpointsPagesWithContext(ctx, input, func(page *dms.DescribeEndpointsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Endpoints {
			r := ResourceEndpoint()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.EndpointIdentifier))
			d.Set("endpoint_arn", v.EndpointArn)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping DMS Endpoint sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing DMS Endpoints (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping DMS Endpoints (%s): %w", region, err)
	}

	return nil
}

func sweepReplicationConfigs(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.DMSConn(ctx)
	input := &dms.DescribeReplicationConfigsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.DescribeReplicationConfigsPagesWithContext(ctx, input, func(page *dms.DescribeReplicationConfigsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ReplicationConfigs {
			r := ResourceReplicationConfig()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.ReplicationConfigArn))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping DMS Replication Config sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing DMS Replication Configs (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping DMS Replication Configs (%s): %w", region, err)
	}

	return nil
}

func sweepReplicationInstances(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.DMSConn(ctx)
	input := &dms.DescribeReplicationInstancesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.DescribeReplicationInstancesPagesWithContext(ctx, input, func(page *dms.DescribeReplicationInstancesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ReplicationInstances {
			r := ResourceReplicationInstance()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.ReplicationInstanceIdentifier))
			d.Set("replication_instance_arn", v.ReplicationInstanceArn)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping DMS Replication Instance sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing DMS Replication Instances (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping DMS Replication Instances (%s): %w", region, err)
	}

	return nil
}

func sweepReplicationSubnetGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.DMSConn(ctx)
	input := &dms.DescribeReplicationSubnetGroupsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.DescribeReplicationSubnetGroupsPagesWithContext(ctx, input, func(page *dms.DescribeReplicationSubnetGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ReplicationSubnetGroups {
			r := ResourceReplicationSubnetGroup()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.ReplicationSubnetGroupIdentifier))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping DMS Replication Subnet Group sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing DMS Replication Subnet Groups (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping DMS Replication Subnet Groups (%s): %w", region, err)
	}

	return nil
}

func sweepReplicationTasks(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.DMSConn(ctx)
	input := &dms.DescribeReplicationTasksInput{
		WithoutSettings: aws.Bool(true),
	}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.DescribeReplicationTasksPagesWithContext(ctx, input, func(page *dms.DescribeReplicationTasksOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.ReplicationTasks {
			r := ResourceReplicationTask()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.ReplicationTaskIdentifier))
			d.Set("replication_task_arn", v.ReplicationTaskArn)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping DMS Replication Task sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing DMS Replication Tasks (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping DMS Replication Tasks (%s): %w", region, err)
	}

	return nil
}
