//go:build sweep
// +build sweep

package rds

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
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
		Dependencies: []string{
			"aws_rds_cluster",
		},
	})

	resource.AddTestSweepers("aws_db_instance", &resource.Sweeper{
		Name: "aws_db_instance",
		F:    sweepInstances,
		Dependencies: []string{
			"aws_opsworks_rds_db_instance",
		},
	})

	resource.AddTestSweepers("aws_db_option_group", &resource.Sweeper{
		Name: "aws_db_option_group",
		F:    sweepOptionGroups,
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
			"aws_db_instance",
		},
	})
}

func sweepClusterParameterGroups(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).RDSConn

	input := &rds.DescribeDBClusterParameterGroupsInput{}

	for {
		output, err := conn.DescribeDBClusterParameterGroups(input)

		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping RDS DB Cluster Parameter Group sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error retrieving DB Cluster Parameter Groups: %s", err)
		}

		for _, dbcpg := range output.DBClusterParameterGroups {
			if dbcpg == nil {
				continue
			}

			input := &rds.DeleteDBClusterParameterGroupInput{
				DBClusterParameterGroupName: dbcpg.DBClusterParameterGroupName,
			}
			name := aws.StringValue(dbcpg.DBClusterParameterGroupName)

			if strings.HasPrefix(name, "default.") {
				log.Printf("[INFO] Skipping DB Cluster Parameter Group: %s", name)
				continue
			}

			log.Printf("[INFO] Deleting DB Cluster Parameter Group: %s", name)

			_, err := conn.DeleteDBClusterParameterGroup(input)

			if err != nil {
				log.Printf("[ERROR] Failed to delete DB Cluster Parameter Group %s: %s", name, err)
				continue
			}
		}

		if aws.StringValue(output.Marker) == "" {
			break
		}

		input.Marker = output.Marker
	}

	return nil
}

func sweepClusterSnapshots(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*conns.AWSClient).RDSConn
	input := &rds.DescribeDBClusterSnapshotsInput{
		// "InvalidDBClusterSnapshotStateFault: Only manual snapshots may be deleted."
		Filters: []*rds.Filter{{
			Name:   aws.String("snapshot-type"),
			Values: aws.StringSlice([]string{"manual"}),
		}},
	}
	var sweeperErrs *multierror.Error

	for {
		output, err := conn.DescribeDBClusterSnapshots(input)
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping RDS DB Cluster Snapshots sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
		}
		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving RDS DB Cluster Snapshots: %w", err))
			return sweeperErrs
		}

		for _, dbClusterSnapshot := range output.DBClusterSnapshots {
			id := aws.StringValue(dbClusterSnapshot.DBClusterSnapshotIdentifier)

			log.Printf("[INFO] Deleting RDS DB Cluster Snapshot: %s", id)
			_, err := conn.DeleteDBClusterSnapshot(&rds.DeleteDBClusterSnapshotInput{
				DBClusterSnapshotIdentifier: aws.String(id),
			})
			if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBClusterSnapshotNotFoundFault) {
				continue
			}
			if err != nil {
				sweeperErr := fmt.Errorf("error deleting RDS DB Cluster Snapshot (%s): %w", id, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		if aws.StringValue(output.Marker) == "" {
			break
		}
		input.Marker = output.Marker
	}

	return sweeperErrs.ErrorOrNil()
}

func sweepClusters(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).RDSConn
	input := &rds.DescribeDBClustersInput{}

	err = conn.DescribeDBClustersPages(input, func(out *rds.DescribeDBClustersOutput, lastPage bool) bool {
		for _, cluster := range out.DBClusters {
			id := aws.StringValue(cluster.DBClusterIdentifier)

			// Automatically remove from global cluster to bypass this error on deletion:
			// InvalidDBClusterStateFault: This cluster is a part of a global cluster, please remove it from globalcluster first
			if aws.StringValue(cluster.EngineMode) == "global" {
				globalCluster, err := DescribeGlobalClusterFromClusterARN(conn, aws.StringValue(cluster.DBClusterArn))

				if err != nil {
					log.Printf("[ERROR] Failure reading RDS Global Cluster information for DB Cluster (%s): %s", id, err)
				}

				if globalCluster != nil {
					globalClusterID := aws.StringValue(globalCluster.GlobalClusterIdentifier)
					input := &rds.RemoveFromGlobalClusterInput{
						DbClusterIdentifier:     cluster.DBClusterArn,
						GlobalClusterIdentifier: globalCluster.GlobalClusterIdentifier,
					}

					log.Printf("[INFO] Removing RDS Cluster (%s) from RDS Global Cluster: %s", id, globalClusterID)
					_, err = conn.RemoveFromGlobalCluster(input)

					if err != nil {
						log.Printf("[ERROR] Failure removing RDS Cluster (%s) from RDS Global Cluster (%s): %s", id, globalClusterID, err)
					}
				}
			}

			input := &rds.DeleteDBClusterInput{
				DBClusterIdentifier: cluster.DBClusterIdentifier,
				SkipFinalSnapshot:   aws.Bool(true),
			}

			log.Printf("[INFO] Deleting RDS DB Cluster: %s", id)

			_, err := conn.DeleteDBCluster(input)

			if err != nil {
				log.Printf("[ERROR] Failed to delete RDS DB Cluster (%s): %s", id, err)
				continue
			}

			if err := WaitForClusterDeletion(conn, id, 40*time.Minute); err != nil { //nolint:gomnd
				log.Printf("[ERROR] Failure while waiting for RDS DB Cluster (%s) to be deleted: %s", id, err)
			}
		}
		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping RDS DB Cluster sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error retrieving RDS DB Clusters: %s", err)
	}

	return nil
}

func sweepEventSubscriptions(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).RDSConn
	input := &rds.DescribeEventSubscriptionsInput{}
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.DescribeEventSubscriptionsPages(input, func(page *rds.DescribeEventSubscriptionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, eventSubscription := range page.EventSubscriptionsList {
			r := ResourceEventSubscription()
			d := r.Data(nil)
			d.SetId(aws.StringValue(eventSubscription.CustSubscriptionId))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping RDS Event Subscription sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing RDS Event Subscriptions (%s): %w", region, err)
	}

	err = sweep.SweepOrchestrator(sweepResources)

	if err != nil {
		return fmt.Errorf("error sweeping RDS Event Subscriptions (%s): %w", region, err)
	}

	return nil
}

func sweepGlobalClusters(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).RDSConn
	input := &rds.DescribeGlobalClustersInput{}

	err = conn.DescribeGlobalClustersPages(input, func(out *rds.DescribeGlobalClustersOutput, lastPage bool) bool {
		for _, globalCluster := range out.GlobalClusters {
			id := aws.StringValue(globalCluster.GlobalClusterIdentifier)
			input := &rds.DeleteGlobalClusterInput{
				GlobalClusterIdentifier: globalCluster.GlobalClusterIdentifier,
			}

			log.Printf("[INFO] Deleting RDS Global Cluster: %s", id)

			_, err := conn.DeleteGlobalCluster(input)

			if err != nil {
				log.Printf("[ERROR] Failed to delete RDS Global Cluster (%s): %s", id, err)
				continue
			}

			if err := WaitForGlobalClusterDeletion(conn, id); err != nil {
				log.Printf("[ERROR] Failure while waiting for RDS Global Cluster (%s) to be deleted: %s", id, err)
			}
		}
		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping RDS Global Cluster sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error retrieving RDS Global Clusters: %s", err)
	}

	return nil
}

func sweepInstances(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).RDSConn
	sweepResources := make([]*sweep.SweepResource, 0)

	err = conn.DescribeDBInstancesPages(&rds.DescribeDBInstancesInput{}, func(out *rds.DescribeDBInstancesOutput, lastPage bool) bool {
		for _, dbi := range out.DBInstances {
			r := ResourceInstance()
			d := r.Data(nil)
			d.SetId(aws.StringValue(dbi.DBInstanceIdentifier))
			d.Set("skip_final_snapshot", true)
			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}
		return !lastPage
	})
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping RDS DB Instance sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving DB instances: %s", err)
	}

	return sweep.SweepOrchestrator(sweepResources)
}

func sweepOptionGroups(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).RDSConn

	opts := rds.DescribeOptionGroupsInput{}
	resp, err := conn.DescribeOptionGroups(&opts)
	if err != nil {
		if sweep.SkipSweepError(err) {
			log.Printf("[WARN] Skipping RDS DB Option Group sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("error describing DB Option Groups in Sweeper: %s", err)
	}

	for _, og := range resp.OptionGroupsList {
		if strings.HasPrefix(aws.StringValue(og.OptionGroupName), "default") {
			continue
		}

		log.Printf("[INFO] Deleting RDS Option Group: %s", aws.StringValue(og.OptionGroupName))

		deleteOpts := &rds.DeleteOptionGroupInput{
			OptionGroupName: og.OptionGroupName,
		}

		ret := resource.Retry(1*time.Minute, func() *resource.RetryError {
			_, err := conn.DeleteOptionGroup(deleteOpts)
			if err != nil {
				if tfawserr.ErrCodeEquals(err, rds.ErrCodeInvalidOptionGroupStateFault) {
					log.Printf("[DEBUG] AWS believes the RDS Option Group is still in use, retrying")
					return resource.RetryableError(err)
				}
				return resource.NonRetryableError(err)
			}
			return nil
		})
		if ret != nil {
			return fmt.Errorf("Error Deleting DB Option Group (%s) in Sweeper: %s", *og.OptionGroupName, ret)
		}
	}

	return nil
}

func sweepParameterGroups(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).RDSConn

	err = conn.DescribeDBParameterGroupsPages(&rds.DescribeDBParameterGroupsInput{}, func(out *rds.DescribeDBParameterGroupsOutput, lastPage bool) bool {
		for _, dbpg := range out.DBParameterGroups {
			if dbpg == nil {
				continue
			}

			input := &rds.DeleteDBParameterGroupInput{
				DBParameterGroupName: dbpg.DBParameterGroupName,
			}
			name := aws.StringValue(dbpg.DBParameterGroupName)

			if strings.HasPrefix(name, "default.") {
				log.Printf("[INFO] Skipping DB Parameter Group: %s", name)
				continue
			}

			log.Printf("[INFO] Deleting DB Parameter Group: %s", name)

			_, err := conn.DeleteDBParameterGroup(input)

			if err != nil {
				log.Printf("[ERROR] Failed to delete DB Parameter Group %s: %s", name, err)
				continue
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping RDS DB Parameter Group sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error retrieving DB Parameter Groups: %s", err)
	}

	return nil
}

func sweepProxies(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)
	if err != nil {
		return fmt.Errorf("Error getting client: %s", err)
	}
	conn := client.(*conns.AWSClient).RDSConn

	err = conn.DescribeDBProxiesPages(&rds.DescribeDBProxiesInput{}, func(out *rds.DescribeDBProxiesOutput, lastPage bool) bool {
		for _, dbpg := range out.DBProxies {
			if dbpg == nil {
				continue
			}

			input := &rds.DeleteDBProxyInput{
				DBProxyName: dbpg.DBProxyName,
			}
			name := aws.StringValue(dbpg.DBProxyName)

			log.Printf("[INFO] Deleting DB Proxy: %s", name)

			_, err := conn.DeleteDBProxy(input)

			if err != nil {
				log.Printf("[ERROR] Failed to delete DB Proxy %s: %s", name, err)
				continue
			}
		}

		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping RDS DB Proxy sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error retrieving DB Proxies: %s", err)
	}

	return nil
}

func sweepSnapshots(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).RDSConn
	input := &rds.DescribeDBSnapshotsInput{}
	var sweeperErrs error

	err = conn.DescribeDBSnapshotsPages(input, func(out *rds.DescribeDBSnapshotsOutput, lastPage bool) bool {
		if out == nil {
			return !lastPage
		}

		for _, dbSnapshot := range out.DBSnapshots {
			if dbSnapshot == nil {
				continue
			}

			id := aws.StringValue(dbSnapshot.DBSnapshotIdentifier)
			input := &rds.DeleteDBSnapshotInput{
				DBSnapshotIdentifier: dbSnapshot.DBSnapshotIdentifier,
			}

			if strings.HasPrefix(id, "rds:") {
				log.Printf("[INFO] Skipping RDS Automated DB Snapshot: %s", id)
				continue
			}

			log.Printf("[INFO] Deleting RDS DB Snapshot: %s", id)
			_, err := conn.DeleteDBSnapshot(input)

			if tfawserr.ErrCodeEquals(err, rds.ErrCodeDBSnapshotNotFoundFault) {
				continue
			}

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting RDS DB Snapshot (%s): %w", id, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			}
		}
		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping RDS DB Snapshot sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error describing RDS DB Snapshots: %s", err)
	}

	return sweeperErrs
}

func sweepSubnetGroups(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*conns.AWSClient).RDSConn
	input := &rds.DescribeDBSubnetGroupsInput{}

	err = conn.DescribeDBSubnetGroupsPages(input, func(out *rds.DescribeDBSubnetGroupsOutput, lastPage bool) bool {
		for _, dbSubnetGroup := range out.DBSubnetGroups {
			name := aws.StringValue(dbSubnetGroup.DBSubnetGroupName)
			input := &rds.DeleteDBSubnetGroupInput{
				DBSubnetGroupName: dbSubnetGroup.DBSubnetGroupName,
			}

			log.Printf("[INFO] Deleting RDS DB Subnet Group: %s", name)

			_, err := conn.DeleteDBSubnetGroup(input)

			if err != nil {
				log.Printf("[ERROR] Failed to delete RDS DB Subnet Group (%s): %s", name, err)
			}
		}
		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping RDS DB Subnet Group sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error retrieving RDS DB Subnet Groups: %s", err)
	}

	return nil
}
