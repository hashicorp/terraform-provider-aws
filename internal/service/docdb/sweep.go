// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package docdb

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/docdb"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/sdk"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_docdb_cluster", sweepClusters, "aws_docdb_cluster_instance")
	awsv2.Register("aws_docdb_cluster_instance", sweepClusterInstances)
	awsv2.Register("aws_docdb_cluster_snapshot", sweepClusterSnapshots, "aws_docdb_cluster")
	awsv2.Register("aws_docdb_global_cluster", sweepGlobalClusters, "aws_docdb_cluster")

	// No sweepers for
	// * aws_docdb_cluster_parameter_group
	// * aws_docdb_event_subscription
	// * aws_docdb_subnet_group
	// as they are the same as the RDS resources, and will be swept by RDS.
}

func sweepClusters(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.DocDBClient(ctx)
	var input docdb.DescribeDBClustersInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := docdb.NewDescribeDBClustersPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.DBClusters {
			if engine := aws.ToString(v.Engine); engine != engineDocDB {
				continue
			}

			id := aws.ToString(v.DBClusterIdentifier)
			r := resourceCluster()
			d := r.Data(nil)
			d.SetId(id)

			// Refresh.
			if err := sdk.ReadResource(ctx, r, d, client); err != nil {
				log.Printf("[WARN] Skipping DocumentDB Cluster %s: %s", id, err)
				continue
			}

			d.Set(names.AttrApplyImmediately, true)
			d.Set(names.AttrDeletionProtection, false)
			d.Set("skip_final_snapshot", true)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepClusterSnapshots(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.DocDBClient(ctx)
	var input docdb.DescribeDBClusterSnapshotsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := docdb.NewDescribeDBClusterSnapshotsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.DBClusterSnapshots {
			if engine := aws.ToString(v.Engine); engine != engineDocDB {
				continue
			}

			id := aws.ToString(v.DBClusterSnapshotIdentifier)

			if typ := aws.ToString(v.SnapshotType); typ != "manual" {
				log.Printf("[INFO] Skipping DocDB Cluster Snapshot %s: SnapshotType=%s", id, typ)
				continue
			}

			r := resourceClusterSnapshot()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepClusterInstances(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.DocDBClient(ctx)
	var input docdb.DescribeDBInstancesInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := docdb.NewDescribeDBInstancesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
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

	return sweepResources, nil
}

func sweepGlobalClusters(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.DocDBClient(ctx)
	var input docdb.DescribeGlobalClustersInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := docdb.NewDescribeGlobalClustersPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.GlobalClusters {
			if engine := aws.ToString(v.Engine); engine != engineDocDB {
				continue
			}

			id := aws.ToString(v.GlobalClusterIdentifier)
			r := resourceGlobalCluster()
			d := r.Data(nil)
			d.SetId(id)

			// Refresh.
			if err := sdk.ReadResource(ctx, r, d, client); err != nil {
				log.Printf("[WARN] Skipping DocumentDB Global Cluster %s: %s", id, err)
				continue
			}

			d.Set(names.AttrDeletionProtection, false)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}
