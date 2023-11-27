// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package neptune

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv1"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_neptune_event_subscription", &resource.Sweeper{
		Name: "aws_neptune_event_subscription",
		F:    sweepEventSubscriptions,
	})

	resource.AddTestSweepers("aws_neptune_cluster", &resource.Sweeper{
		Name: "aws_neptune_cluster",
		F:    sweepClusters,
		Dependencies: []string{
			"aws_neptune_cluster_instance",
		},
	})

	resource.AddTestSweepers("aws_neptune_cluster_instance", &resource.Sweeper{
		Name: "aws_neptune_cluster_instance",
		F:    sweepClusterInstances,
	})

	resource.AddTestSweepers("aws_neptune_global_cluster", &resource.Sweeper{
		Name: "aws_neptune_global_cluster",
		F:    sweepGlobalClusters,
		Dependencies: []string{
			"aws_neptune_cluster",
		},
	})
}

func sweepEventSubscriptions(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.NeptuneConn(ctx)
	input := &neptune.DescribeEventSubscriptionsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.DescribeEventSubscriptionsPagesWithContext(ctx, input, func(page *neptune.DescribeEventSubscriptionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.EventSubscriptionsList {
			r := ResourceEventSubscription()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.CustSubscriptionId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Neptune Event Subscription sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("listing Neptune Event Subscriptions (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("sweeping Neptune Event Subscriptions (%s): %w", region, err)
	}

	return nil
}

func sweepClusters(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	conn := client.NeptuneConn(ctx)
	input := &neptune.DescribeDBClustersInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.DescribeDBClustersPagesWithContext(ctx, input, func(page *neptune.DescribeDBClustersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.DBClusters {
			arn := aws.StringValue(v.DBClusterArn)
			id := aws.StringValue(v.DBClusterIdentifier)

			r := ResourceCluster()
			d := r.Data(nil)
			d.SetId(id)
			d.Set("apply_immediately", true)
			d.Set("arn", arn)
			d.Set("deletion_protection", false)
			d.Set("skip_final_snapshot", true)

			globalCluster, err := findGlobalClusterByClusterARN(ctx, conn, arn)

			if err != nil && !tfresource.NotFound(err) {
				log.Printf("[WARN] Reading Neptune Cluster %s Global Cluster information: %s", id, err)
				continue
			}

			if globalCluster != nil && globalCluster.GlobalClusterIdentifier != nil {
				d.Set("global_cluster_identifier", globalCluster.GlobalClusterIdentifier)
			}

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Neptune Cluster sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("listing Neptune Clusters (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("sweeping Neptune Clusters (%s): %w", region, err)
	}

	return nil
}

func sweepClusterInstances(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	conn := client.NeptuneConn(ctx)
	input := &neptune.DescribeDBInstancesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.DescribeDBInstancesPagesWithContext(ctx, input, func(page *neptune.DescribeDBInstancesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.DBInstances {
			id := aws.StringValue(v.DBInstanceIdentifier)

			if state := aws.StringValue(v.DBInstanceStatus); state == dbInstanceStatusDeleting {
				log.Printf("[INFO] Skipping Neptune Cluster Instance %s: DBInstanceStatus=%s", id, state)
				continue
			}

			r := ResourceClusterInstance()
			d := r.Data(nil)
			d.SetId(id)
			d.Set("apply_immediately", true)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Neptune Cluster Instance sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("listing Neptune Cluster Instances (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("sweeping Neptune Cluster Instances (%s): %w", region, err)
	}

	return nil
}

func sweepGlobalClusters(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %w", err)
	}
	conn := client.NeptuneConn(ctx)
	input := &neptune.DescribeGlobalClustersInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.DescribeGlobalClustersPagesWithContext(ctx, input, func(page *neptune.DescribeGlobalClustersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.GlobalClusters {
			r := ResourceGlobalCluster()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.GlobalClusterIdentifier))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Neptune Global Cluster sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("listing Neptune Global Clusters (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("sweeping Neptune Global Clusters (%s): %w", region, err)
	}

	return nil
}
