// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package neptune

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/neptune"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
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

	resource.AddTestSweepers("aws_neptune_cluster_parameter_group", &resource.Sweeper{
		Name: "aws_neptune_cluster_parameter_group",
		F:    sweepClusterParameterGroups,
		Dependencies: []string{
			"aws_neptune_cluster",
		},
	})

	resource.AddTestSweepers("aws_neptune_cluster_snapshot", &resource.Sweeper{
		Name: "aws_neptune_cluster_snapshot",
		F:    sweepClusterSnapshots,
		Dependencies: []string{
			"aws_neptune_cluster",
		},
	})

	resource.AddTestSweepers("aws_neptune_event_subscription", &resource.Sweeper{
		Name: "aws_neptune_event_subscription",
		F:    sweepEventSubscriptions,
	})

	resource.AddTestSweepers("aws_neptune_global_cluster", &resource.Sweeper{
		Name: "aws_neptune_global_cluster",
		F:    sweepGlobalClusters,
		Dependencies: []string{
			"aws_neptune_cluster",
		},
	})

	resource.AddTestSweepers("aws_neptune_parameter_group", &resource.Sweeper{
		Name: "aws_neptune_parameter_group",
		F:    sweepParameterGroups,
		Dependencies: []string{
			"aws_neptune_cluster_instance",
		},
	})

	resource.AddTestSweepers("aws_neptune_subnet_group", &resource.Sweeper{
		Name: "aws_neptune_subnet_group",
		F:    sweepSubnetGroups,
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
	conn := client.NeptuneClient(ctx)
	input := &neptune.DescribeEventSubscriptionsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := neptune.NewDescribeEventSubscriptionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Neptune Event Subscription sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("listing Neptune Event Subscriptions (%s): %w", region, err)
		}

		for _, v := range page.EventSubscriptionsList {
			r := resourceEventSubscription()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.CustSubscriptionId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
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
	conn := client.NeptuneClient(ctx)
	input := &neptune.DescribeDBClustersInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := neptune.NewDescribeDBClustersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Neptune Cluster sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("listing Neptune Clusters (%s): %w", region, err)
		}

		for _, v := range page.DBClusters {
			arn := aws.ToString(v.DBClusterArn)
			id := aws.ToString(v.DBClusterIdentifier)

			r := resourceCluster()
			d := r.Data(nil)
			d.SetId(id)
			d.Set(names.AttrApplyImmediately, true)
			d.Set(names.AttrARN, arn)
			d.Set(names.AttrDeletionProtection, false)
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
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("sweeping Neptune Clusters (%s): %w", region, err)
	}

	return nil
}

func sweepClusterSnapshots(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.NeptuneClient(ctx)
	input := &neptune.DescribeDBClusterSnapshotsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := neptune.NewDescribeDBClusterSnapshotsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Neptune Cluster Snapshot sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("listing Neptune Cluster Snapshots (%s): %w", region, err)
		}

		for _, v := range page.DBClusterSnapshots {
			r := resourceClusterSnapshot()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.DBClusterSnapshotIdentifier))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("sweeping Neptune Cluster Snapshots (%s): %w", region, err)
	}

	return nil
}

func sweepClusterParameterGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.NeptuneClient(ctx)
	input := &neptune.DescribeDBClusterParameterGroupsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := neptune.NewDescribeDBClusterParameterGroupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Neptune Cluster Parameter Group sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Neptune Cluster Parameter Groups (%s): %w", region, err)
		}

		for _, v := range page.DBClusterParameterGroups {
			name := aws.ToString(v.DBClusterParameterGroupName)

			if strings.HasPrefix(name, "default.") {
				log.Printf("[INFO] Skipping Neptune Cluster Parameter Group: %s", name)
				continue
			}

			r := resourceClusterParameterGroup()
			d := r.Data(nil)
			d.SetId(name)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Neptune Cluster Parameter Groups (%s): %w", region, err)
	}

	return nil
}

func sweepClusterInstances(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("getting client: %s", err)
	}
	conn := client.NeptuneClient(ctx)
	input := &neptune.DescribeDBInstancesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := neptune.NewDescribeDBInstancesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Neptune Cluster Instance sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("listing Neptune Cluster Instances (%s): %w", region, err)
		}

		for _, v := range page.DBInstances {
			if aws.ToString(v.Engine) != engineNeptune {
				continue
			}

			id := aws.ToString(v.DBInstanceIdentifier)

			if state := aws.ToString(v.DBInstanceStatus); state == dbInstanceStatusDeleting {
				log.Printf("[INFO] Skipping Neptune Cluster Instance %s: DBInstanceStatus=%s", id, state)
				continue
			}

			r := resourceClusterInstance()
			d := r.Data(nil)
			d.SetId(id)
			d.Set(names.AttrApplyImmediately, true)
			d.Set("skip_final_snapshot", true)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
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
	conn := client.NeptuneClient(ctx)
	input := &neptune.DescribeGlobalClustersInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := neptune.NewDescribeGlobalClustersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Neptune Global Cluster sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("listing Neptune Global Clusters (%s): %w", region, err)
		}

		for _, v := range page.GlobalClusters {
			r := resourceGlobalCluster()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.GlobalClusterIdentifier))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("sweeping Neptune Global Clusters (%s): %w", region, err)
	}

	return nil
}

func sweepParameterGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.NeptuneClient(ctx)
	input := &neptune.DescribeDBParameterGroupsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := neptune.NewDescribeDBParameterGroupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Neptune Parameter Group sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Neptune Parameter Groups (%s): %w", region, err)
		}

		for _, v := range page.DBParameterGroups {
			name := aws.ToString(v.DBParameterGroupName)

			if strings.HasPrefix(name, "default.") {
				log.Printf("[INFO] Skipping Neptune Parameter Group: %s", name)
				continue
			}

			r := resourceParameterGroup()
			d := r.Data(nil)
			d.SetId(name)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Neptune Parameter Groups (%s): %w", region, err)
	}

	return nil
}

func sweepSubnetGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.NeptuneClient(ctx)
	input := &neptune.DescribeDBSubnetGroupsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := neptune.NewDescribeDBSubnetGroupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping Neptune Subnet Group sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Neptune Subnet Groups (%s): %w", region, err)
		}

		for _, v := range page.DBSubnetGroups {
			r := resourceSubnetGroup()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.DBSubnetGroupName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping Neptune Subnet Groups (%s): %w", region, err)
	}

	return nil
}
