// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package docdb

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/docdb"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_docdb_cluster", sweepClusters,
		"aws_docdb_cluster_instance",
	)

	resource.AddTestSweepers("aws_docdb_cluster_instance", &resource.Sweeper{
		Name: "aws_docdb_cluster_instance",
		F:    sweepClusterInstances,
	})

	resource.AddTestSweepers("aws_docdb_cluster_parameter_group", &resource.Sweeper{
		Name: "aws_docdb_cluster_parameter_group",
		F:    sweepClusterParameterGroups,
		Dependencies: []string{
			"aws_docdb_cluster",
		},
	})

	resource.AddTestSweepers("aws_docdb_cluster_snapshot", &resource.Sweeper{
		Name: "aws_docdb_cluster_snapshot",
		F:    sweepClusterSnapshots,
		Dependencies: []string{
			"aws_docdb_cluster",
		},
	})

	resource.AddTestSweepers("aws_docdb_event_subscription", &resource.Sweeper{
		Name: "aws_docdb_event_subscription",
		F:    sweepEventSubscriptions,
	})

	resource.AddTestSweepers("aws_docdb_global_cluster", &resource.Sweeper{
		Name: "aws_docdb_global_cluster",
		F:    sweepGlobalClusters,
		Dependencies: []string{
			"aws_docdb_cluster",
		},
	})

	resource.AddTestSweepers("aws_docdb_subnet_group", &resource.Sweeper{
		Name: "aws_docdb_subnet_group",
		F:    sweepSubnetGroups,
		Dependencies: []string{
			"aws_docdb_cluster",
		},
	})
}

func sweepClusters(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.DocDBClient(ctx)

	var sweepResources []sweep.Sweepable
	r := resourceCluster()

	input := docdb.DescribeDBClustersInput{}
	pages := docdb.NewDescribeDBClustersPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, v := range page.DBClusters {
			arn := aws.ToString(v.DBClusterArn)
			id := aws.ToString(v.DBClusterIdentifier)

			d := r.Data(nil)
			d.SetId(id)
			d.Set(names.AttrARN, arn)
			d.Set("skip_final_snapshot", true)

			globalCluster, err := findGlobalClusterByClusterARN(ctx, conn, arn)

			if err != nil && !tfresource.NotFound(err) {
				log.Printf("[WARN] Reading DocumentDB Cluster %s Global Cluster information: %s", id, err)
				continue
			}

			if globalCluster != nil && globalCluster.GlobalClusterIdentifier != nil {
				d.Set("global_cluster_identifier", globalCluster.GlobalClusterIdentifier)
			}

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepClusterSnapshots(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.DocDBClient(ctx)
	input := &docdb.DescribeDBClusterSnapshotsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := docdb.NewDescribeDBClusterSnapshotsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping DocumentDB Cluster Snapshot sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("listing DocumentDB Cluster Snapshots (%s): %w", region, err)
		}

		for _, v := range page.DBClusterSnapshots {
			r := ResourceClusterSnapshot()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.DBClusterSnapshotIdentifier))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("sweeping DocumentDB Cluster Snapshots (%s): %w", region, err)
	}

	return nil
}

func sweepClusterParameterGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.DocDBClient(ctx)
	input := &docdb.DescribeDBClusterParameterGroupsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := docdb.NewDescribeDBClusterParameterGroupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping DocumentDB Cluster Parameter Group sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing DocumentDB Cluster Parameter Groups (%s): %s", region, err)
		}

		for _, v := range page.DBClusterParameterGroups {
			name := aws.ToString(v.DBClusterParameterGroupName)

			if strings.HasPrefix(name, "default.") {
				log.Printf("[INFO] Skipping DocumentDB Cluster Parameter Group: %s", name)
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
		return fmt.Errorf("error sweeping DocumentDB Cluster Parameter Groups (%s): %w", region, err)
	}

	return nil
}

func sweepClusterInstances(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.DocDBClient(ctx)
	input := &docdb.DescribeDBInstancesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := docdb.NewDescribeDBInstancesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping DocumentDB Cluster Instance sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("listing DocumentDB Cluster Instances (%s): %w", region, err)
		}

		for _, v := range page.DBInstances {
			if engine := aws.ToString(v.Engine); engine != engineDocDB {
				continue
			}

			r := resourceClusterInstance()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.DBInstanceIdentifier))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("sweeping DocumentDB Cluster Instances (%s): %w", region, err)
	}

	return nil
}

func sweepGlobalClusters(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.DocDBClient(ctx)
	input := &docdb.DescribeGlobalClustersInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := docdb.NewDescribeGlobalClustersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping DocumentDB Global Cluster sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("listing DocumentDB Global Clusters (%s): %w", region, err)
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
		return fmt.Errorf("sweeping DocumentDB Global Clusters (%s): %w", region, err)
	}

	return nil
}

func sweepSubnetGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.DocDBClient(ctx)
	input := &docdb.DescribeDBSubnetGroupsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := docdb.NewDescribeDBSubnetGroupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping DocumentDB Subnet Group sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("listing DocumentDB Subnet Groups (%s): %w", region, err)
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
		return fmt.Errorf("error sweeping DocumentDB Subnet Groups (%s): %w", region, err)
	}

	return nil
}

func sweepEventSubscriptions(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.DocDBClient(ctx)
	input := &docdb.DescribeEventSubscriptionsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := docdb.NewDescribeEventSubscriptionsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping DocumentDB Event Subscription sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("listing DocumentDB Event Subscriptions (%s): %w", region, err)
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
		return fmt.Errorf("sweeping DocumentDB Event Subscriptions (%s): %w", region, err)
	}

	return nil
}
