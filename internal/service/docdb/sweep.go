package docdb

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/docdb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)

func init() {
	resource.AddTestSweepers("aws_docdb_global_cluster", &resource.Sweeper{
		Name:         "aws_docdb_global_cluster",
		F:            sweepGlobalClusters,
		Dependencies: []string{
			// "aws_docdb_cluster",
		},
	})
}

func sweepDBClusters(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).DocDBConn
	input := &docdb.DescribeDBClustersInput{}

	err = conn.DescribeDBClustersPages(input, func(out *docdb.DescribeDBClustersOutput, lastPage bool) bool {
		for _, dBCluster := range out.DBClusters {
			id := aws.StringValue(dBCluster.DBClusterIdentifier)
			input := &docdb.DeleteDBClusterInput{
				DBClusterIdentifier: dBCluster.DBClusterIdentifier,
			}

			log.Printf("[INFO] Deleting DocDB Cluster: %s", id)

			_, err := conn.DeleteDBCluster(input)

			if err != nil {
				log.Printf("[ERROR] Failed to delete DocDB Cluster (%s): %s", id, err)
				continue
			}

			if err := WaitForDBClusterDeletion(context.TODO(), conn, id, DBClusterDeleteTimeout); err != nil {
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
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).DocDBConn
	input := &docdb.DescribeDBClusterSnapshotsInput{}

	err = conn.DescribeDBClusterSnapshotsPages(input, func(out *docdb.DescribeDBClusterSnapshotsOutput, lastPage bool) bool {
		for _, dBClusterSnapshot := range out.DBClusterSnapshots {
			name := aws.StringValue(dBClusterSnapshot.DBClusterSnapshotIdentifier)
			input := &docdb.DeleteDBClusterSnapshotInput{
				DBClusterSnapshotIdentifier: dBClusterSnapshot.DBClusterSnapshotIdentifier,
			}

			log.Printf("[INFO] Deleting DocDB Cluster Snapshot: %s", name)

			_, err := conn.DeleteDBClusterSnapshot(input)

			if err != nil {
				log.Printf("[ERROR] Failed to delete DocDB Cluster Snapshot (%s): %s", name, err)
				continue
			}

			if err := WaitForDBClusterSnapshotDeletion(context.TODO(), conn, name, DBClusterSnapshotDeleteTimeout); err != nil {
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

func sweepDBInstances(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).DocDBConn
	input := &docdb.DescribeDBInstancesInput{}

	err = conn.DescribeDBInstancesPages(input, func(out *docdb.DescribeDBInstancesOutput, lastPage bool) bool {
		for _, dBInstance := range out.DBInstances {
			id := aws.StringValue(dBInstance.DBInstanceIdentifier)
			input := &docdb.DeleteDBInstanceInput{
				DBInstanceIdentifier: dBInstance.DBInstanceIdentifier,
			}

			log.Printf("[INFO] Deleting DocDB Instance: %s", id)

			_, err := conn.DeleteDBInstance(input)

			if err != nil {
				log.Printf("[ERROR] Failed to delete DocDB Instance (%s): %s", id, err)
				continue
			}

			if err := WaitForDBInstanceDeletion(context.TODO(), conn, id, DBInstanceDeleteTimeout); err != nil {
				log.Printf("[ERROR] Failure while waiting for DocDB Instance (%s) to be deleted: %s", id, err)
			}
		}
		return !lastPage
	})

	if sweep.SkipSweepError(err) {
		log.Printf("[WARN] Skipping DocDB Instance sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error retrieving DocDB Instances: %w", err)
	}

	return nil
}

func sweepGlobalClusters(region string) error {
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).DocDBConn
	input := &docdb.DescribeGlobalClustersInput{}

	err = conn.DescribeGlobalClustersPages(input, func(out *docdb.DescribeGlobalClustersOutput, lastPage bool) bool {
		for _, globalCluster := range out.GlobalClusters {
			id := aws.StringValue(globalCluster.GlobalClusterIdentifier)
			input := &docdb.DeleteGlobalClusterInput{
				GlobalClusterIdentifier: globalCluster.GlobalClusterIdentifier,
			}

			log.Printf("[INFO] Deleting DocDB Global Cluster: %s", id)

			_, err := conn.DeleteGlobalCluster(input)

			if err != nil {
				log.Printf("[ERROR] Failed to delete DocDB Global Cluster (%s): %s", id, err)
				continue
			}

			if err := WaitForGlobalClusterDeletion(context.TODO(), conn, id, GlobalClusterDeleteTimeout); err != nil {
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
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).DocDBConn
	input := &docdb.DescribeDBSubnetGroupsInput{}

	err = conn.DescribeDBSubnetGroupsPages(input, func(out *docdb.DescribeDBSubnetGroupsOutput, lastPage bool) bool {
		for _, dBSubnetGroup := range out.DBSubnetGroups {
			name := aws.StringValue(dBSubnetGroup.DBSubnetGroupName)
			input := &docdb.DeleteDBSubnetGroupInput{
				DBSubnetGroupName: dBSubnetGroup.DBSubnetGroupName,
			}

			log.Printf("[INFO] Deleting DocDB Subnet Group: %s", name)

			_, err := conn.DeleteDBSubnetGroup(input)

			if err != nil {
				log.Printf("[ERROR] Failed to delete DocDB Subnet Group (%s): %s", name, err)
				continue
			}

			if err := WaitForDBSubnetGroupDeletion(context.TODO(), conn, name, DBSubnetGroupDeleteTimeout); err != nil {
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
	client, err := sweep.SharedRegionalSweepClient(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*conns.AWSClient).DocDBConn
	input := &docdb.DescribeEventSubscriptionsInput{}

	err = conn.DescribeEventSubscriptionsPages(input, func(out *docdb.DescribeEventSubscriptionsOutput, lastPage bool) bool {
		for _, eventSubscription := range out.EventSubscriptionsList {
			id := aws.StringValue(eventSubscription.CustSubscriptionId)
			input := &docdb.DeleteEventSubscriptionInput{
				SubscriptionName: eventSubscription.CustSubscriptionId,
			}

			log.Printf("[INFO] Deleting DocDB Event Subscription: %s", id)

			_, err := conn.DeleteEventSubscription(input)

			if err != nil {
				log.Printf("[ERROR] Failed to delete DocDB Event Subscription (%s): %s", id, err)
				continue
			}

			if _, err := waitEventSubscriptionDeleted(context.TODO(), conn, id, EventSubscriptionDeleteTimeout); err != nil {
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
