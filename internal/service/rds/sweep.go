// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep/awsv1"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func RegisterSweepers() {
	resource.AddTestSweepers("aws_rds_cluster_parameter_group", &resource.Sweeper{
		Name: "aws_rds_cluster_parameter_group",
		F:    sweepClusterParameterGroups,
		Dependencies: []string{
			"aws_rds_cluster",
		},
	})

	resource.AddTestSweepers("aws_db_cluster_snapshot", &resource.Sweeper{
		Name: "aws_db_cluster_snapshot",
		F:    sweepClusterSnapshots,
		Dependencies: []string{
			"aws_rds_cluster",
		},
	})

	resource.AddTestSweepers("aws_rds_cluster", &resource.Sweeper{
		Name: "aws_rds_cluster",
		F:    sweepClusters,
		Dependencies: []string{
			"aws_db_instance",
		},
	})

	resource.AddTestSweepers("aws_db_event_subscription", &resource.Sweeper{
		Name: "aws_db_event_subscription",
		F:    sweepEventSubscriptions,
	})

	resource.AddTestSweepers("aws_rds_global_cluster", &resource.Sweeper{
		Name: "aws_rds_global_cluster",
		F:    sweepGlobalClusters,
	})

	resource.AddTestSweepers("aws_db_instance", &resource.Sweeper{
		Name: "aws_db_instance",
		F:    sweepInstances,
		Dependencies: []string{
			"aws_rds_global_cluster",
		},
	})

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
}

func sweepClusterParameterGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.RDSConn(ctx)
	input := &rds.DescribeDBClusterParameterGroupsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.DescribeDBClusterParameterGroupsPagesWithContext(ctx, input, func(page *rds.DescribeDBClusterParameterGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.DBClusterParameterGroups {
			name := aws.StringValue(v.DBClusterParameterGroupName)

			if strings.HasPrefix(name, "default.") {
				continue
			}

			r := resourceClusterParameterGroup()
			d := r.Data(nil)
			d.SetId(name)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping RDS Cluster Parameter Group sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing RDS Cluster Parameter Groups (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping RDS Cluster Parameter Groups (%s): %w", region, err)
	}

	return nil
}

func sweepClusterSnapshots(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.RDSConn(ctx)
	input := &rds.DescribeDBClusterSnapshotsInput{
		// "InvalidDBClusterSnapshotStateFault: Only manual snapshots may be deleted."
		Filters: []*rds.Filter{{
			Name:   aws.String("snapshot-type"),
			Values: aws.StringSlice([]string{"manual"}),
		}},
	}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.DescribeDBClusterSnapshotsPagesWithContext(ctx, input, func(page *rds.DescribeDBClusterSnapshotsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.DBClusterSnapshots {
			r := resourceClusterSnapshot()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.DBClusterSnapshotIdentifier))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping RDS DB Cluster Snapshot sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing RDS DB Cluster Snapshots (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping RDS DB Cluster Snapshots (%s): %w", region, err)
	}

	return nil
}

func sweepClusters(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.RDSConn(ctx)

	var sweeperErrs *multierror.Error
	sweepResources := make([]sweep.Sweepable, 0)

	input := &rds.DescribeDBClustersInput{}
	err = conn.DescribeDBClustersPagesWithContext(ctx, input, func(page *rds.DescribeDBClustersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.DBClusters {
			arn := aws.StringValue(v.DBClusterArn)
			id := aws.StringValue(v.DBClusterIdentifier)
			r := resourceCluster()
			d := r.Data(nil)
			d.SetId(id)
			d.Set(names.AttrApplyImmediately, true)
			d.Set(names.AttrARN, arn)
			d.Set("delete_automated_backups", true)
			d.Set(names.AttrDeletionProtection, false)
			d.Set("skip_final_snapshot", true)

			if engineMode := aws.StringValue(v.EngineMode); engineMode == engineModeGlobal || engineMode == engineModeProvisioned {
				globalCluster, err := FindGlobalClusterByDBClusterARN(ctx, conn, arn)
				if err != nil {
					if !tfresource.NotFound(err) {
						sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("reading RDS Global Cluster information for DB Cluster (%s): %s", id, err))
						continue
					}
				}

				if globalCluster != nil && globalCluster.GlobalClusterIdentifier != nil {
					d.Set("global_cluster_identifier", globalCluster.GlobalClusterIdentifier)
				}
			}

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})
	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping RDS Cluster sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing RDS Clusters (%s): %w", region, err))
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error sweeping RDS Clusters (%s): %w", region, err))
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepEventSubscriptions(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.RDSConn(ctx)
	input := &rds.DescribeEventSubscriptionsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.DescribeEventSubscriptionsPagesWithContext(ctx, input, func(page *rds.DescribeEventSubscriptionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, eventSubscription := range page.EventSubscriptionsList {
			r := resourceEventSubscription()
			d := r.Data(nil)
			d.SetId(aws.StringValue(eventSubscription.CustSubscriptionId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping RDS Event Subscription sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing RDS Event Subscriptions (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping RDS Event Subscriptions (%s): %w", region, err)
	}

	return nil
}

func sweepGlobalClusters(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.RDSConn(ctx)
	input := &rds.DescribeGlobalClustersInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.DescribeGlobalClustersPagesWithContext(ctx, input, func(page *rds.DescribeGlobalClustersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.GlobalClusters {
			r := ResourceGlobalCluster()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.GlobalClusterIdentifier))
			d.Set(names.AttrForceDestroy, true)
			d.Set("global_cluster_members", flattenGlobalClusterMembers(v.GlobalClusterMembers))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping RDS Global Cluster sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing RDS Global Clusters (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping RDS Global Clusters (%s): %w", region, err)
	}

	return nil
}

func sweepInstances(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	input := &rds.DescribeDBInstancesInput{}
	conn := client.RDSConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.DescribeDBInstancesPagesWithContext(ctx, input, func(page *rds.DescribeDBInstancesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.DBInstances {
			r := ResourceInstance()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.DbiResourceId))
			d.Set(names.AttrApplyImmediately, true)
			d.Set("delete_automated_backups", true)
			d.Set(names.AttrDeletionProtection, false)
			d.Set(names.AttrIdentifier, v.DBInstanceIdentifier)
			d.Set("skip_final_snapshot", true)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping RDS DB Instance sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing RDS DB Instances (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping RDS DB Instances (%s): %w", region, err)
	}

	return nil
}

func sweepOptionGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(ctx, region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	input := &rds.DescribeOptionGroupsInput{}
	conn := client.RDSConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.DescribeOptionGroupsPagesWithContext(ctx, input, func(page *rds.DescribeOptionGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.OptionGroupsList {
			name := aws.StringValue(v.OptionGroupName)

			if strings.HasPrefix(name, "default:") {
				continue
			}

			r := ResourceOptionGroup()
			d := r.Data(nil)
			d.SetId(name)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping RDS Option Group sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing RDS Option Groups (%s): %w", region, err)
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
	conn := client.RDSConn(ctx)
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.DescribeDBParameterGroupsPagesWithContext(ctx, input, func(page *rds.DescribeDBParameterGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.DBParameterGroups {
			name := aws.StringValue(v.DBParameterGroupName)

			if strings.HasPrefix(name, "default.") {
				continue
			}

			r := resourceParameterGroup()
			d := r.Data(nil)
			d.SetId(name)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping RDS DB Parameter Group sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing RDS DB Parameter Groups (%s): %w", region, err)
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
	conn := client.RDSConn(ctx)
	input := &rds.DescribeDBProxiesInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.DescribeDBProxiesPagesWithContext(ctx, input, func(page *rds.DescribeDBProxiesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.DBProxies {
			r := resourceProxy()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.DBProxyName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping RDS DB Proxy sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing RDS DB Proxies (%s): %w", region, err)
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
	conn := client.RDSConn(ctx)
	input := &rds.DescribeDBSnapshotsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.DescribeDBSnapshotsPagesWithContext(ctx, input, func(page *rds.DescribeDBSnapshotsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.DBSnapshots {
			id := aws.StringValue(v.DBSnapshotIdentifier)

			if strings.HasPrefix(id, "rds:") {
				continue
			}

			r := resourceSnapshot()
			d := r.Data(nil)
			d.SetId(id)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping RDS DB Snapshot sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing RDS DB Snapshots (%s): %w", region, err)
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
	conn := client.RDSConn(ctx)
	input := &rds.DescribeDBSubnetGroupsInput{}
	sweepResources := make([]sweep.Sweepable, 0)

	err = conn.DescribeDBSubnetGroupsPagesWithContext(ctx, input, func(page *rds.DescribeDBSubnetGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.DBSubnetGroups {
			r := resourceSubnetGroup()
			d := r.Data(nil)
			d.SetId(aws.StringValue(v.DBSubnetGroupName))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping RDS DB Subnet Group sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing RDS DB Subnet Groups (%s): %w", region, err)
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
	conn := client.RDSConn(ctx)
	input := &rds.DescribeDBInstanceAutomatedBackupsInput{}
	sweepResources := make([]sweep.Sweepable, 0)
	var backupARNs []string

	err = conn.DescribeDBInstanceAutomatedBackupsPagesWithContext(ctx, input, func(page *rds.DescribeDBInstanceAutomatedBackupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.DBInstanceAutomatedBackups {
			arn := aws.StringValue(v.DBInstanceAutomatedBackupsArn)
			r := ResourceInstanceAutomatedBackupsReplication()
			d := r.Data(nil)
			d.SetId(arn)
			d.Set("source_db_instance_arn", v.DBInstanceArn)
			backupARNs = append(backupARNs, arn)

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if awsv1.SkipSweepError(err) {
		log.Printf("[WARN] Skipping RDS Instance Automated Backup sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing RDS Instance Automated Backups (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(ctx, sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping RDS Instance Automated Backups (%s): %w", region, err)
	}

	// Since there is no resource for automated backups themselves, they are swept here.
	for _, v := range backupARNs {
		log.Printf("[DEBUG] Deleting RDS Instance Automated Backup: %s", v)
		_, err = conn.DeleteDBInstanceAutomatedBackupWithContext(ctx, &rds.DeleteDBInstanceAutomatedBackupInput{
			DBInstanceAutomatedBackupsArn: aws.String(v),
		})

		if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBInstanceAutomatedBackupNotFoundFault) {
			continue
		}

		if err != nil {
			log.Printf("[ERROR] deleting RDS Instance Automated Backup (%s): %s", v, err)
		}
	}

	return nil
}
