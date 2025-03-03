// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	awsv2.Register("aws_rds_cluster_parameter_group", sweepClusterParameterGroups, "aws_rds_cluster")
	awsv2.Register("aws_db_cluster_snapshot", sweepClusterSnapshots, "aws_rds_cluster")
	awsv2.Register("aws_rds_cluster", sweepClusters, "aws_db_instance", "aws_rds_shard_group")
	awsv2.Register("aws_db_event_subscription", sweepEventSubscriptions)
	awsv2.Register("aws_rds_global_cluster", sweepGlobalClusters)
	awsv2.Register("aws_db_instance", sweepInstances, "aws_rds_global_cluster")

	resource.AddTestSweepers("aws_db_option_group", &resource.Sweeper{
		Name: "aws_db_option_group",
		F:    sweepOptionGroups,
		Dependencies: []string{
			"aws_rds_cluster",
			"aws_db_snapshot",
		},
	})

	resource.AddTestSweepers("aws_db_parameter_group", &resource.Sweeper{
		Name: "aws_db_parameter_group",
		F:    sweepParameterGroups,
		Dependencies: []string{
			"aws_db_instance",
		},
	})

	resource.AddTestSweepers("aws_db_proxy", &resource.Sweeper{
		Name: "aws_db_proxy",
		F:    sweepProxies,
	})

	resource.AddTestSweepers("aws_db_snapshot", &resource.Sweeper{
		Name: "aws_db_snapshot",
		F:    sweepSnapshots,
		Dependencies: []string{
			"aws_db_instance",
		},
	})

	resource.AddTestSweepers("aws_db_subnet_group", &resource.Sweeper{
		Name: "aws_db_subnet_group",
		F:    sweepSubnetGroups,
		Dependencies: []string{
			"aws_rds_cluster",
		},
	})

	resource.AddTestSweepers("aws_db_instance_automated_backups_replication", &resource.Sweeper{
		Name: "aws_db_instance_automated_backups_replication",
		F:    sweepInstanceAutomatedBackups,
		Dependencies: []string{
			"aws_db_instance",
		},
	})

	awsv2.Register("aws_rds_shard_group", sweepShardGroups)
}

func sweepClusterParameterGroups(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.RDSClient(ctx)
	var input rds.DescribeDBClusterParameterGroupsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := rds.NewDescribeDBClusterParameterGroupsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.DBClusterParameterGroups {
			name := aws.ToString(v.DBClusterParameterGroupName)

			if strings.HasPrefix(name, "default.") {
				log.Printf("[INFO] Skipping RDS Cluster Parameter Group %s", name)
				continue
			}

			r := resourceClusterParameterGroup()
			d := r.Data(nil)
			d.SetId(name)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepClusterSnapshots(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.RDSClient(ctx)
	input := rds.DescribeDBClusterSnapshotsInput{
		// "InvalidDBClusterSnapshotStateFault: Only manual snapshots may be deleted."
		Filters: []types.Filter{{
			Name:   aws.String("snapshot-type"),
			Values: []string{"manual"},
		}},
	}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := rds.NewDescribeDBClusterSnapshotsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.DBClusterSnapshots {
			r := resourceClusterSnapshot()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.DBClusterSnapshotIdentifier))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepClusters(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.RDSClient(ctx)
	var input rds.DescribeDBClustersInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := rds.NewDescribeDBClustersPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.DBClusters {
			arn := aws.ToString(v.DBClusterArn)
			id := aws.ToString(v.DBClusterIdentifier)
			r := resourceCluster()
			d := r.Data(nil)
			d.SetId(id)
			d.Set(names.AttrApplyImmediately, true)
			d.Set(names.AttrARN, arn)
			d.Set("delete_automated_backups", true)
			d.Set(names.AttrDeletionProtection, false)
			d.Set("skip_final_snapshot", true)

			if engineMode := aws.ToString(v.EngineMode); engineMode == engineModeGlobal || engineMode == engineModeProvisioned {
				globalCluster, err := findGlobalClusterByDBClusterARN(ctx, conn, arn)

				if err != nil && !tfresource.NotFound(err) {
					log.Printf("[WARN] Reading RDS Global Cluster information for DB Cluster (%s): %s", id, err)
					continue
				}

				if globalCluster != nil && globalCluster.GlobalClusterIdentifier != nil {
					d.Set("global_cluster_identifier", globalCluster.GlobalClusterIdentifier)
				}
			}

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepEventSubscriptions(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.RDSClient(ctx)
	var input rds.DescribeEventSubscriptionsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := rds.NewDescribeEventSubscriptionsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.EventSubscriptionsList {
			r := resourceEventSubscription()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.CustSubscriptionId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepGlobalClusters(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.RDSClient(ctx)
	var input rds.DescribeGlobalClustersInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := rds.NewDescribeGlobalClustersPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.GlobalClusters {
			r := resourceGlobalCluster()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.GlobalClusterIdentifier))
			d.Set(names.AttrForceDestroy, true)
			d.Set("global_cluster_members", flattenGlobalClusterMembers(v.GlobalClusterMembers))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepInstances(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.RDSClient(ctx)
	var input rds.DescribeDBInstancesInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := rds.NewDescribeDBInstancesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.DBInstances {
			id := aws.ToString(v.DbiResourceId)

			switch engine := aws.ToString(v.Engine); engine {
			case "docdb", "neptune":
				// These engines are handled by their respective services' sweepers.
				continue
			case InstanceEngineMySQL:
				// "InvalidParameterValue: Deleting cluster instances isn't supported for DB engine mysql".
				if clusterID := aws.ToString(v.DBClusterIdentifier); clusterID != "" {
					log.Printf("[INFO] Skipping RDS DB Instance %s: DBClusterIdentifier=%s", id, clusterID)
					continue
				}
			}

			r := resourceInstance()
			d := r.Data(nil)
			d.SetId(id)
			d.Set(names.AttrApplyImmediately, true)
			d.Set("delete_automated_backups", true)
			d.Set(names.AttrDeletionProtection, false)
			d.Set(names.AttrIdentifier, v.DBInstanceIdentifier)
			d.Set("skip_final_snapshot", true)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepOptionGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	input := &rds.DescribeOptionGroupsInput{}
	conn := client.RDSClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := rds.NewDescribeOptionGroupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping RDS Option Group sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing RDS Option Groups (%s): %w", region, err)
		}

		for _, v := range page.OptionGroupsList {
			name := aws.ToString(v.OptionGroupName)

			if strings.HasPrefix(name, "default:") {
				log.Printf("[INFO] Skipping RDS Option Group %s", name)
				continue
			}

			r := resourceOptionGroup()
			d := r.Data(nil)
			d.SetId(name)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping RDS Option Groups (%s): %w", region, err)
	}

	return nil
}

func sweepParameterGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	input := &rds.DescribeDBParameterGroupsInput{}
	conn := client.RDSClient(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	pages := rds.NewDescribeDBParameterGroupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping RDS DB Parameter Group sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing RDS DB Parameter Groups (%s): %w", region, err)
		}

		for _, v := range page.DBParameterGroups {
			name := aws.ToString(v.DBParameterGroupName)

			if strings.HasPrefix(name, "default.") {
				log.Printf("[INFO] Skipping RDS DB Parameter Group %s", name)
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
		return fmt.Errorf("error sweeping RDS DB Parameter Groups (%s): %w", region, err)
	}

	return nil
}

func sweepProxies(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("Error getting client: %s", err)
	}
	conn := client.RDSClient(ctx)
	input := &rds.DescribeDBProxiesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := rds.NewDescribeDBProxiesPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping RDS DB Proxy sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing RDS DB Proxies (%s): %w", region, err)
		}

		for _, v := range page.DBProxies {
			r := resourceProxy()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.DBProxyName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping RDS DB Proxies (%s): %w", region, err)
	}

	return nil
}

func sweepSnapshots(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.RDSClient(ctx)
	input := &rds.DescribeDBSnapshotsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := rds.NewDescribeDBSnapshotsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping RDS DB Snapshot sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing RDS DB Snapshots (%s): %w", region, err)
		}

		for _, v := range page.DBSnapshots {
			id := aws.ToString(v.DBSnapshotIdentifier)

			if strings.HasPrefix(id, "rds:") {
				log.Printf("[INFO] Skipping RDS DB Snapshot %s", id)
				continue
			}

			r := resourceSnapshot()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping RDS DB Snapshots (%s): %w", region, err)
	}

	return nil
}

func sweepSubnetGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.RDSClient(ctx)
	input := &rds.DescribeDBSubnetGroupsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	pages := rds.NewDescribeDBSubnetGroupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping RDS DB Subnet Group sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing RDS DB Subnet Groups (%s): %w", region, err)
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
		return fmt.Errorf("error sweeping RDS DB Subnet Groups (%s): %w", region, err)
	}

	return nil
}

func sweepInstanceAutomatedBackups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.RDSClient(ctx)
	input := &rds.DescribeDBInstanceAutomatedBackupsInput{}
	sweepResources := make([]sweep.Sweepable, 0)
	var backupARNs []string

	pages := rds.NewDescribeDBInstanceAutomatedBackupsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if awsv2.SkipSweepError(err) {
			log.Printf("[WARN] Skipping RDS Instance Automated Backup sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing RDS Instance Automated Backups (%s): %w", region, err)
		}

		for _, v := range page.DBInstanceAutomatedBackups {
			arn := aws.ToString(v.DBInstanceAutomatedBackupsArn)
			r := resourceInstanceAutomatedBackupsReplication()
			d := r.Data(nil)
			d.SetId(arn)
			d.Set("source_db_instance_arn", v.DBInstanceArn)
			backupARNs = append(backupARNs, arn)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping RDS Instance Automated Backups (%s): %w", region, err)
	}

	// Since there is no resource for automated backups themselves, they are swept here.
	for _, v := range backupARNs {
		log.Printf("[DEBUG] Deleting RDS Instance Automated Backup: %s", v)
		_, err = conn.DeleteDBInstanceAutomatedBackup(ctx, &rds.DeleteDBInstanceAutomatedBackupInput{
			DBInstanceAutomatedBackupsArn: aws.String(v),
		})

		if errs.IsA[*types.DBInstanceAutomatedBackupNotFoundFault](err) {
			continue
		}

		if err != nil {
			log.Printf("[WARN] Deleting RDS Instance Automated Backup (%s): %s", v, err)
		}
	}

	return nil
}

func sweepShardGroups(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.RDSClient(ctx)
	var input rds.DescribeDBShardGroupsInput
	sweepResources := make([]sweep.Sweepable, 0)

	err := describeDBShardGroupsPages(ctx, conn, &input, func(page *rds.DescribeDBShardGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.DBShardGroups {
			sweepResources = append(sweepResources, framework.NewSweepResource(newShardGroupResource, client,
				framework.NewAttribute("db_shard_group_identifier", aws.ToString(v.DBShardGroupIdentifier))))
		}

		return !lastPage
	})

	if err != nil {
		return nil, err
	}

	return sweepResources, nil
}
