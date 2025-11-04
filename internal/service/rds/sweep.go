// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv2"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/framework"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/sdk"
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
	awsv2.Register("aws_db_option_group", sweepOptionGroups, "aws_rds_cluster", "aws_db_snapshot")
	awsv2.Register("aws_db_parameter_group", sweepParameterGroups, "aws_db_instance")
	awsv2.Register("aws_db_proxy", sweepProxies)
	awsv2.Register("aws_db_snapshot", sweepSnapshots, "aws_db_instance")
	awsv2.Register("aws_db_subnet_group", sweepSubnetGroups, "aws_rds_cluster")
	awsv2.Register("aws_db_instance_automated_backups_replication", sweepInstanceAutomatedBackups, "aws_db_instance")
	awsv2.Register("aws_rds_shard_group", sweepShardGroups)
	awsv2.Register("aws_rds_blue_green_deployment", sweepBlueGreenDeployments, "aws_db_instance") // Pseudo resource.
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

func sweepOptionGroups(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.RDSClient(ctx)
	var input rds.DescribeOptionGroupsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := rds.NewDescribeOptionGroupsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
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

	return sweepResources, nil
}

func sweepParameterGroups(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.RDSClient(ctx)
	var input rds.DescribeDBParameterGroupsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := rds.NewDescribeDBParameterGroupsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
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

	return sweepResources, nil
}

func sweepProxies(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.RDSClient(ctx)
	var input rds.DescribeDBProxiesInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := rds.NewDescribeDBProxiesPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.DBProxies {
			r := resourceProxy()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.DBProxyName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepSnapshots(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.RDSClient(ctx)
	var input rds.DescribeDBSnapshotsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := rds.NewDescribeDBSnapshotsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
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

	return sweepResources, nil
}

func sweepSubnetGroups(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.RDSClient(ctx)
	var input rds.DescribeDBSubnetGroupsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := rds.NewDescribeDBSubnetGroupsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.DBSubnetGroups {
			r := resourceSubnetGroup()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.DBSubnetGroupName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
	}

	return sweepResources, nil
}

func sweepInstanceAutomatedBackups(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.RDSClient(ctx)
	var input rds.DescribeDBInstanceAutomatedBackupsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := rds.NewDescribeDBInstanceAutomatedBackupsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.DBInstanceAutomatedBackups {
			r := resourceInstanceAutomatedBackupsReplication()
			d := r.Data(nil)
			d.SetId(aws.ToString(v.DBInstanceAutomatedBackupsArn))
			d.Set("source_db_instance_arn", v.DBInstanceArn)

			sweepResources = append(sweepResources, newInstanceAutomatedBackupSweeper(ctx, r, d, client))
		}
	}

	return sweepResources, nil
}

type instanceAutomatedBackupSweeper struct {
	conn      *rds.Client
	sweepable sweep.Sweepable
	backupARN string
}

func newInstanceAutomatedBackupSweeper(ctx context.Context, resource *schema.Resource, d *schema.ResourceData, client *conns.AWSClient) sweep.Sweepable {
	return &instanceAutomatedBackupSweeper{
		conn:      client.RDSClient(ctx),
		sweepable: sdk.NewSweepResource(resource, d, client),
		backupARN: d.Id(),
	}
}

func (s instanceAutomatedBackupSweeper) Delete(ctx context.Context, optFns ...tfresource.OptionsFunc) error {
	if err := s.sweepable.Delete(ctx, optFns...); err != nil {
		return err
	}

	// Since there is no resource for automated backups themselves, they are swept here.
	tflog.Info(ctx, "Deleting RDS Instance Automated Backup", map[string]any{
		"backup ARN": s.backupARN,
	})

	_, err := s.conn.DeleteDBInstanceAutomatedBackup(ctx, &rds.DeleteDBInstanceAutomatedBackupInput{
		DBInstanceAutomatedBackupsArn: aws.String(s.backupARN),
	})

	if errs.IsA[*types.DBInstanceAutomatedBackupNotFoundFault](err) || errs.IsA[*types.InvalidDBInstanceAutomatedBackupStateFault](err) {
		err = nil
	}

	if err != nil {
		return fmt.Errorf("deleting RDS Instance Automated Backup (%s): %w", s.backupARN, err)
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

func sweepBlueGreenDeployments(ctx context.Context, client *conns.AWSClient) ([]sweep.Sweepable, error) {
	conn := client.RDSClient(ctx)
	var input rds.DescribeBlueGreenDeploymentsInput
	sweepResources := make([]sweep.Sweepable, 0)

	pages := rds.NewDescribeBlueGreenDeploymentsPaginator(conn, &input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if err != nil {
			return nil, err
		}

		for _, v := range page.BlueGreenDeployments {
			sweepResources = append(sweepResources, newBlueGreenDeploymentSweeper(ctx, aws.ToString(v.BlueGreenDeploymentIdentifier), client))
		}
	}

	return sweepResources, nil
}

type blueGreenDeploymentSweeper struct {
	conn *rds.Client
	id   string
}

func newBlueGreenDeploymentSweeper(ctx context.Context, id string, client *conns.AWSClient) sweep.Sweepable {
	return &blueGreenDeploymentSweeper{
		conn: client.RDSClient(ctx),
		id:   id,
	}
}

func (s blueGreenDeploymentSweeper) Delete(ctx context.Context, optFns ...tfresource.OptionsFunc) error {
	input := rds.DeleteBlueGreenDeploymentInput{
		BlueGreenDeploymentIdentifier: aws.String(s.id),
	}
	_, err := s.conn.DeleteBlueGreenDeployment(ctx, &input)

	if err != nil {
		return fmt.Errorf("deleting RDS Blue/Green Deployment (%s): %w", s.id, err)
	}

	const (
		timeout = 10 * time.Minute
	)
	if _, err := waitBlueGreenDeploymentDeleted(ctx, s.conn, s.id, timeout); err != nil {
		return fmt.Errorf("waiting for RDS Blue/Green Deployment (%s) delete: %w", s.id, err)
	}

	return nil
}
