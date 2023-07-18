// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

//go:build sweep
// +build sweep

package neptune

import (
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func init() {
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
		Dependencies: []string{
			"aws_rds_global_cluster",
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
	var sweeperErrs *multierror.Error

	err = conn.DescribeEventSubscriptionsPagesWithContext(ctx, &neptune.DescribeEventSubscriptionsInput{}, func(page *neptune.DescribeEventSubscriptionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, eventSubscription := range page.EventSubscriptionsList {
			name := aws.StringValue(eventSubscription.CustSubscriptionId)

			log.Printf("[INFO] Deleting Neptune Event Subscription: %s", name)
			_, err = conn.DeleteEventSubscriptionWithContext(ctx, &neptune.DeleteEventSubscriptionInput{
				SubscriptionName: aws.String(name),
			})
			if tfawserr.ErrCodeEquals(err, neptune.ErrCodeSubscriptionNotFoundFault) {
				continue
			}
			if err != nil {
				sweeperErr := fmt.Errorf("deleting Neptune Event Subscription (%s): %w", name, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}

			_, err = WaitEventSubscriptionDeleted(ctx, conn, name)
			if tfawserr.ErrCodeEquals(err, neptune.ErrCodeSubscriptionNotFoundFault) {
				continue
			}
			if err != nil {
				sweeperErr := fmt.Errorf("waiting for Neptune Event Subscription (%s) deletion: %w", name, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})
	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Neptune Event Subscriptions sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("retrieving Neptune Event Subscriptions: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepClusters(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	conn := client.NeptuneConn(ctx)

	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	input := &neptune.DescribeDBClustersInput{}
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

			if err != nil {
				if !tfresource.NotFound(err) {
					sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("reading Neptune Global Cluster information for Neptune Cluster (%s): %s", id, err))
					continue
				}
			}

			if globalCluster != nil && globalCluster.GlobalClusterIdentifier != nil {
				d.Set("global_cluster_identifier", globalCluster.GlobalClusterIdentifier)
			}

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})
	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping Neptune Cluster sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("listing Neptune Clusters (%s): %w", region, err))
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("sweeping Neptune Clusters (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepClusterInstances(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	conn := client.NeptuneConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	input := &neptune.DescribeDBInstancesInput{}
	err = conn.DescribeDBInstancesPagesWithContext(ctx, input, func(page *neptune.DescribeDBInstancesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.DBInstances {
			log.Printf("Neptune Instance ID: %q", aws.StringValue(v.DBInstanceIdentifier))
			r := ResourceClusterInstance()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.DBInstanceIdentifier))
			d.Set("apply_immediately", true)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
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
