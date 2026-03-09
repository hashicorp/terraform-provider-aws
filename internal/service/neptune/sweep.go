// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package neptune

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/neptune"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_neptune_cluster", sweepClusters, "aws_neptune_cluster_instance")
	awsv2.Register("aws_neptune_cluster_instance", sweepClusterInstances)
	awsv2.Register("aws_neptune_cluster_snapshot", sweepClusterSnapshots, "aws_neptune_cluster")
	awsv2.Register("aws_neptune_global_cluster", sweepGlobalClusters, "aws_neptune_cluster")

	// No sweepers for
	// * aws_neptune_cluster_parameter_group
	// * aws_neptune_event_subscription
	// * aws_neptune_parameter_group
	// * aws_neptune_subnet_group
	// as they are the same as the RDS resources, and will be swept by RDS.
}

func sweepClusters(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.NeptuneClient(ctx)
	var input neptune.DescribeDBClustersInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := neptune.NewDescribeDBClustersPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.DBClusters {
			if engine := aws.ToString(v.Engine); engine != defaultEngine {
				continue
			}

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

			if err != nil && !retry.NotFound(err) {
				log.Printf("[WARN] Reading Neptune Cluster %s Global Cluster information: %s", id, err)
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

func sweepClusterSnapshots(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.NeptuneClient(ctx)
	var input neptune.DescribeDBClusterSnapshotsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := neptune.NewDescribeDBClusterSnapshotsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.DBClusterSnapshots {
			if engine := aws.ToString(v.Engine); engine != defaultEngine {
				continue
			}

			id := aws.ToString(v.DBClusterSnapshotIdentifier)

			if typ := aws.ToString(v.SnapshotType); typ != "manual" {
				log.Printf("[INFO] Skipping Neptune Cluster Snapshot %s: SnapshotType=%s", id, typ)
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
	conn := client.NeptuneClient(ctx)
	var input neptune.DescribeDBInstancesInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := neptune.NewDescribeDBInstancesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.DBInstances {
			if aws.ToString(v.Engine) != defaultEngine {
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

	return sweepResources, nil
}

func sweepGlobalClusters(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.NeptuneClient(ctx)
	var input neptune.DescribeGlobalClustersInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := neptune.NewDescribeGlobalClustersPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.GlobalClusters {
			if engine := aws.ToString(v.Engine); engine != defaultEngine {
				continue
			}

			r := resourceGlobalCluster()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.GlobalClusterIdentifier))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}
