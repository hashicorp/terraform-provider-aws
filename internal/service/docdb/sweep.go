//go:build sweep
// +build sweep

package docdb

import (
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/docdb"
	multierror "github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_docdb_global_cluster", &resource.Sweeper{
		Name: "aws_docdb_global_cluster",
		F:    sweepGlobalClusters,
		Dependencies: []string{
			"aws_docdb_cluster",
		},
	})

	resource.AddTestSweepers("aws_docdb_subnet_group", &resource.Sweeper{
		Name: "aws_docdb_subnet_group",
		F:    sweepDBSubnetGroups,
		Dependencies: []string{
			"aws_docdb_cluster_instance",
		},
	})

	resource.AddTestSweepers("aws_docdb_event_subscription", &resource.Sweeper{
		Name: "aws_docdb_event_subscription",
		F:    sweepEventSubscriptions,
	})

	resource.AddTestSweepers("aws_docdb_cluster", &resource.Sweeper{
		Name: "aws_docdb_cluster",
		F:    sweepDBClusters,
		Dependencies: []string{
			"aws_docdb_cluster_instance",
			"aws_docdb_cluster_snapshot",
		},
	})

	resource.AddTestSweepers("aws_docdb_cluster_snapshot", &resource.Sweeper{
		Name: "aws_docdb_cluster_snapshot",
		F:    sweepDBClusterSnapshots,
		Dependencies: []string{
			"aws_docdb_cluster_instance",
		},
	})

	resource.AddTestSweepers("aws_docdb_cluster_instance", &resource.Sweeper{
		Name: "aws_docdb_cluster_instance",
		F:    sweepDBInstances,
	})

	resource.AddTestSweepers("aws_docdb_cluster_parameter_group", &resource.Sweeper{
		Name: "aws_docdb_cluster_parameter_group",
		F:    sweepDBClusterParameterGroups,
		Dependencies: []string{
			"aws_docdb_cluster_instance",
		},
	})
}

func sweepDBClusters(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).DocDBConn()
	input := &docdb.DescribeDBClustersInput{}

	err = conn.DescribeDBClustersPagesWithContext(ctx, input, func(out *docdb.DescribeDBClustersOutput, lastPage bool) bool {
		for _, dBCluster := range out.DBClusters {
			id := aws.StringValue(dBCluster.DBClusterIdentifier)
			input := &docdb.DeleteDBClusterInput{
				DBClusterIdentifier: dBCluster.DBClusterIdentifier,
				SkipFinalSnapshot:   aws.Bool(true),
			}

			log.Printf("[INFO] Deleting DocDB Cluster: %s", id)

			_, err := conn.DeleteDBClusterWithContext(ctx, input)

			if err != nil {
				log.Printf("[ERROR] Failed to delete DocDB Cluster (%s): %s", id, err)
				continue
			}

			if err := WaitForDBClusterDeletion(ctx, conn, id, DBClusterDeleteTimeout); err != nil {
				log.Printf("[ERROR] Failure while waiting for DocDB Cluster (%s) to be deleted: %s", id, err)
			}
		}
		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping DocDB Cluster sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error retrieving DocDB Clusters: %w", err)
	}

	return nil
}

func sweepDBClusterSnapshots(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).DocDBConn()
	input := &docdb.DescribeDBClusterSnapshotsInput{}

	err = conn.DescribeDBClusterSnapshotsPagesWithContext(ctx, input, func(out *docdb.DescribeDBClusterSnapshotsOutput, lastPage bool) bool {
		for _, dBClusterSnapshot := range out.DBClusterSnapshots {
			name := aws.StringValue(dBClusterSnapshot.DBClusterSnapshotIdentifier)
			input := &docdb.DeleteDBClusterSnapshotInput{
				DBClusterSnapshotIdentifier: dBClusterSnapshot.DBClusterSnapshotIdentifier,
			}

			log.Printf("[INFO] Deleting DocDB Cluster Snapshot: %s", name)

			_, err := conn.DeleteDBClusterSnapshotWithContext(ctx, input)

			if err != nil {
				log.Printf("[ERROR] Failed to delete DocDB Cluster Snapshot (%s): %s", name, err)
				continue
			}

			if err := WaitForDBClusterSnapshotDeletion(ctx, conn, name, DBClusterSnapshotDeleteTimeout); err != nil {
				log.Printf("[ERROR] Failure while waiting for DocDB Cluster Snapshot (%s) to be deleted: %s", name, err)
			}
		}
		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping DocDB Cluster Snapshot sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error retrieving DocDB Cluster Snapshots: %w", err)
	}

	return nil
}

func sweepDBClusterParameterGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).DocDBConn()
	input := &docdb.DescribeDBClusterParameterGroupsInput{}

	err = conn.DescribeDBClusterParameterGroupsPagesWithContext(ctx, input, func(out *docdb.DescribeDBClusterParameterGroupsOutput, lastPage bool) bool {
		for _, dBClusterParameterGroup := range out.DBClusterParameterGroups {
			name := aws.StringValue(dBClusterParameterGroup.DBClusterParameterGroupName)
			input := &docdb.DeleteDBClusterParameterGroupInput{
				DBClusterParameterGroupName: dBClusterParameterGroup.DBClusterParameterGroupName,
			}

			if strings.HasPrefix(name, "default.") {
				log.Printf("[INFO] Skipping Document DB Parameter Group: %s", name)
				continue
			}

			log.Printf("[INFO] Deleting DocDB Cluster Parameter Group: %s", name)

			_, err := conn.DeleteDBClusterParameterGroupWithContext(ctx, input)

			if err != nil {
				log.Printf("[ERROR] Failed to delete DocDB Parameter Group (%s): %s", name, err)
				continue
			}
		}
		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping DocDB Cluster Parameter Group sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error retrieving DocDB Cluster Parameter Groups: %w", err)
	}

	return nil
}

func sweepDBInstances(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).DocDBConn()
	sweepResources := make([]sweep.Sweepable, 0)
	var errs *multierror.Error
	input := &docdb.DescribeDBInstancesInput{}

	err = conn.DescribeDBInstancesPagesWithContext(ctx, input, func(page *docdb.DescribeDBInstancesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, dBInstance := range page.DBInstances {
			r := ResourceClusterInstance()
			d := r.Data(nil)
			d.SetId(aws.StringValue(dBInstance.DBInstanceIdentifier))

			sweepResources = append(sweepResources, sweep.NewSweepResource(r, d, client))
		}

		return !lastPage
	})

	if err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error listing DocDB Instances for %s: %w", region, err))
	}

	if err = sweep.SweepOrchestratorWithContext(ctx, sweepResources); err != nil {
		errs = multierror.Append(errs, fmt.Errorf("error sweeping DocDB Instances for %s: %w", region, err))
	}

	if sweep.SkipSweepError(errs.ErrorOrNil()) {
		log.Printf("[WARN] Skipping DocDB Instance sweep for %s: %s", region, errs)
		return nil
	}

	return errs.ErrorOrNil()
}

func sweepGlobalClusters(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).DocDBConn()
	input := &docdb.DescribeGlobalClustersInput{}

	err = conn.DescribeGlobalClustersPagesWithContext(ctx, input, func(out *docdb.DescribeGlobalClustersOutput, lastPage bool) bool {
		for _, globalCluster := range out.GlobalClusters {
			id := aws.StringValue(globalCluster.GlobalClusterIdentifier)
			input := &docdb.DeleteGlobalClusterInput{
				GlobalClusterIdentifier: globalCluster.GlobalClusterIdentifier,
			}

			log.Printf("[INFO] Deleting DocDB Global Cluster: %s", id)

			_, err := conn.DeleteGlobalClusterWithContext(ctx, input)

			if err != nil {
				log.Printf("[ERROR] Failed to delete DocDB Global Cluster (%s): %s", id, err)
				continue
			}

			if err := WaitForGlobalClusterDeletion(ctx, conn, id, GlobalClusterDeleteTimeout); err != nil {
				log.Printf("[ERROR] Failure while waiting for DocDB Global Cluster (%s) to be deleted: %s", id, err)
			}
		}
		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping DocDB Global Cluster sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error retrieving DocDB Global Clusters: %w", err)
	}

	return nil
}

func sweepDBSubnetGroups(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).DocDBConn()
	input := &docdb.DescribeDBSubnetGroupsInput{}

	err = conn.DescribeDBSubnetGroupsPagesWithContext(ctx, input, func(out *docdb.DescribeDBSubnetGroupsOutput, lastPage bool) bool {
		for _, dBSubnetGroup := range out.DBSubnetGroups {
			name := aws.StringValue(dBSubnetGroup.DBSubnetGroupName)
			input := &docdb.DeleteDBSubnetGroupInput{
				DBSubnetGroupName: dBSubnetGroup.DBSubnetGroupName,
			}

			log.Printf("[INFO] Deleting DocDB Subnet Group: %s", name)

			_, err := conn.DeleteDBSubnetGroupWithContext(ctx, input)

			if err != nil {
				log.Printf("[ERROR] Failed to delete DocDB Subnet Group (%s): %s", name, err)
				continue
			}

			if err := WaitForDBSubnetGroupDeletion(ctx, conn, name, DBSubnetGroupDeleteTimeout); err != nil {
				log.Printf("[ERROR] Failure while waiting for DocDB Subnet Group (%s) to be deleted: %s", name, err)
			}
		}
		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping DocDB Subnet Group sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error retrieving DocDB Subnet Groups: %w", err)
	}

	return nil
}

func sweepEventSubscriptions(region string) error {
	ctx := sweep.Context(region)
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).DocDBConn()
	input := &docdb.DescribeEventSubscriptionsInput{}

	err = conn.DescribeEventSubscriptionsPagesWithContext(ctx, input, func(out *docdb.DescribeEventSubscriptionsOutput, lastPage bool) bool {
		for _, eventSubscription := range out.EventSubscriptionsList {
			id := aws.StringValue(eventSubscription.CustSubscriptionId)
			input := &docdb.DeleteEventSubscriptionInput{
				SubscriptionName: eventSubscription.CustSubscriptionId,
			}

			log.Printf("[INFO] Deleting DocDB Event Subscription: %s", id)

			_, err := conn.DeleteEventSubscriptionWithContext(ctx, input)

			if err != nil {
				log.Printf("[ERROR] Failed to delete DocDB Event Subscription (%s): %s", id, err)
				continue
			}

			if _, err := waitEventSubscriptionDeleted(ctx, conn, id, EventSubscriptionDeleteTimeout); err != nil {
				log.Printf("[ERROR] Failure while waiting for DocDB Event Subscription (%s) to be deleted: %s", id, err)
			}
		}
		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping DocDB Event Subscription sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error retrieving DocDB Event Subscriptions: %w", err)
	}

	return nil
}
