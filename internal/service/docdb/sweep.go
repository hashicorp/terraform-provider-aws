// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package docdb

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/docdb"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv1"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_docdb_cluster", &resource.Sweeper{
		Name: "aws_docdb_cluster",
		F:    sweepClusters,
		Dependencies: []string{
			"aws_docdb_cluster_instance",
		},
	})

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

func sweepClusters(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %d", err)
	}
	conn := client.DocDBConn(ctx)
	input := &docdb.DescribeDBClustersInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.DescribeDBClustersPagesWithContext(ctx, input, func(page *docdb.DescribeDBClustersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.DBClusters {
			arn := aws.StringValue(v.DBClusterArn)
			id := aws.StringValue(v.DBClusterIdentifier)

			r := ResourceCluster()
			d := r.Data(nil)
			d.SetId(id)
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

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping DocumentDB Cluster sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing DocumentDB Clusters (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping DocumentDB Clusters (%s): %w", region, err)
	}

	return nil
}

func sweepClusterSnapshots(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.DocDBConn(ctx)
	input := &docdb.DescribeDBClusterSnapshotsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.DescribeDBClusterSnapshotsPagesWithContext(ctx, input, func(page *docdb.DescribeDBClusterSnapshotsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.DBClusterSnapshots {
			r := ResourceClusterSnapshot()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.DBClusterSnapshotIdentifier))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping DocumentDB Cluster Snapshot sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("listing DocumentDB Cluster Snapshots (%s): %w", region, err)
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
	conn := client.DocDBConn(ctx)
	input := &docdb.DescribeDBClusterParameterGroupsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.DescribeDBClusterParameterGroupsPagesWithContext(ctx, input, func(page *docdb.DescribeDBClusterParameterGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.DBClusterParameterGroups {
			name := aws.StringValue(v.DBClusterParameterGroupName)

			if strings.HasPrefix(name, "default.") {
				log.Printf("[INFO] Skipping DocumentDB Cluster Parameter Group: %s", name)
				continue
			}

			r := ResourceClusterParameterGroup()
			d := r.Data(nil)
			d.SetId(name)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping DocumentDB Cluster Parameter Group sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing DocumentDB Cluster Parameter Groups (%s): %w", region, err)
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
	conn := client.DocDBConn(ctx)
	input := &docdb.DescribeDBInstancesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.DescribeDBInstancesPagesWithContext(ctx, input, func(page *docdb.DescribeDBInstancesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.DBInstances {
			r := ResourceClusterInstance()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.DBInstanceIdentifier))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping DocumentDB Cluster Instance sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("listing DocumentDB Cluster Instances (%s): %w", region, err)
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
	conn := client.DocDBConn(ctx)
	input := &docdb.DescribeGlobalClustersInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.DescribeGlobalClustersPagesWithContext(ctx, input, func(page *docdb.DescribeGlobalClustersOutput, lastPage bool) bool {
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
		log.Printf("[WARN] Skipping DocumentDB Global Cluster sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("listing DocumentDB Global Clusters (%s): %w", region, err)
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
	conn := client.DocDBConn(ctx)
	input := &docdb.DescribeDBSubnetGroupsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.DescribeDBSubnetGroupsPagesWithContext(ctx, input, func(page *docdb.DescribeDBSubnetGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.DBSubnetGroups {
			r := ResourceSubnetGroup()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.DBSubnetGroupName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping DocumentDB Subnet Group sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing DocumentDB Subnet Groups (%s): %w", region, err)
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
	conn := client.DocDBConn(ctx)
	input := &docdb.DescribeEventSubscriptionsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.DescribeEventSubscriptionsPagesWithContext(ctx, input, func(page *docdb.DescribeEventSubscriptionsOutput, lastPage bool) bool {
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
		log.Printf("[WARN] Skipping DocumentDB Event Subscription sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("listing DocumentDB Event Subscriptions (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("sweeping DocumentDB Event Subscriptions (%s): %w", region, err)
	}

	return nil
}
