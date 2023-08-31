// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package docdb

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/docdb"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func findGlobalClusterByARN(ctx context.Context, conn *docdb.DocDB, dbClusterARN string) (*docdb.GlobalCluster, error) {
	var globalCluster *docdb.GlobalCluster

	input := &docdb.DescribeGlobalClustersInput{
		Filters: []*docdb.Filter{
			{
				Name:   aws.String("db-cluster-id"),
				Values: []*string{aws.String(dbClusterARN)},
			},
		},
	}

	err := conn.DescribeGlobalClustersPagesWithContext(ctx, input, func(page *docdb.DescribeGlobalClustersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, gc := range page.GlobalClusters {
			if gc == nil {
				continue
			}

			for _, globalClusterMember := range gc.GlobalClusterMembers {
				if aws.StringValue(globalClusterMember.DBClusterArn) == dbClusterARN {
					globalCluster = gc
					return false
				}
			}
		}

		return !lastPage
	})

	return globalCluster, err
}

func findGlobalClusterIDByARN(ctx context.Context, conn *docdb.DocDB, arn string) string {
	result, err := conn.DescribeDBClustersWithContext(ctx, &docdb.DescribeDBClustersInput{})
	if err != nil {
		return ""
	}
	for _, cluster := range result.DBClusters {
		if aws.StringValue(cluster.DBClusterArn) == arn {
			return aws.StringValue(cluster.DBClusterIdentifier)
		}
	}
	return ""
}

func FindDBClusterById(ctx context.Context, conn *docdb.DocDB, dBClusterID string) (*docdb.DBCluster, error) {
	var dBCluster *docdb.DBCluster

	input := &docdb.DescribeDBClustersInput{
		DBClusterIdentifier: aws.String(dBClusterID),
	}

	err := conn.DescribeDBClustersPagesWithContext(ctx, input, func(page *docdb.DescribeDBClustersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, dbc := range page.DBClusters {
			if dbc == nil {
				continue
			}

			if aws.StringValue(dbc.DBClusterIdentifier) == dBClusterID {
				dBCluster = dbc
				return false
			}
		}

		return !lastPage
	})

	return dBCluster, err
}

func FindDBClusterSnapshotById(ctx context.Context, conn *docdb.DocDB, dBClusterSnapshotID string) (*docdb.DBClusterSnapshot, error) {
	var dBClusterSnapshot *docdb.DBClusterSnapshot

	input := &docdb.DescribeDBClusterSnapshotsInput{
		DBClusterIdentifier: aws.String(dBClusterSnapshotID),
	}

	err := conn.DescribeDBClusterSnapshotsPagesWithContext(ctx, input, func(page *docdb.DescribeDBClusterSnapshotsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, dbcss := range page.DBClusterSnapshots {
			if dbcss == nil {
				continue
			}

			if aws.StringValue(dbcss.DBClusterIdentifier) == dBClusterSnapshotID {
				dBClusterSnapshot = dbcss
				return false
			}
		}

		return !lastPage
	})

	return dBClusterSnapshot, err
}

func FindDBInstanceById(ctx context.Context, conn *docdb.DocDB, dBInstanceID string) (*docdb.DBInstance, error) {
	var dBInstance *docdb.DBInstance

	input := &docdb.DescribeDBInstancesInput{
		DBInstanceIdentifier: aws.String(dBInstanceID),
	}

	err := conn.DescribeDBInstancesPagesWithContext(ctx, input, func(page *docdb.DescribeDBInstancesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, dbi := range page.DBInstances {
			if dbi == nil {
				continue
			}

			if aws.StringValue(dbi.DBInstanceIdentifier) == dBInstanceID {
				dBInstance = dbi
				return false
			}
		}

		return !lastPage
	})

	return dBInstance, err
}

func FindGlobalClusterById(ctx context.Context, conn *docdb.DocDB, globalClusterID string) (*docdb.GlobalCluster, error) {
	var globalCluster *docdb.GlobalCluster

	input := &docdb.DescribeGlobalClustersInput{
		GlobalClusterIdentifier: aws.String(globalClusterID),
	}

	err := conn.DescribeGlobalClustersPagesWithContext(ctx, input, func(page *docdb.DescribeGlobalClustersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, gc := range page.GlobalClusters {
			if gc == nil {
				continue
			}

			if aws.StringValue(gc.GlobalClusterIdentifier) == globalClusterID {
				globalCluster = gc
				return false
			}
		}

		return !lastPage
	})

	return globalCluster, err
}

func FindDBSubnetGroupByName(ctx context.Context, conn *docdb.DocDB, dBSubnetGroupName string) (*docdb.DBSubnetGroup, error) {
	var dBSubnetGroup *docdb.DBSubnetGroup

	input := &docdb.DescribeDBSubnetGroupsInput{
		DBSubnetGroupName: aws.String(dBSubnetGroupName),
	}

	err := conn.DescribeDBSubnetGroupsPagesWithContext(ctx, input, func(page *docdb.DescribeDBSubnetGroupsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, sg := range page.DBSubnetGroups {
			if sg == nil {
				continue
			}

			if aws.StringValue(sg.DBSubnetGroupName) == dBSubnetGroupName {
				dBSubnetGroup = sg
				return false
			}
		}

		return !lastPage
	})

	return dBSubnetGroup, err
}

func FindEventSubscriptionByID(ctx context.Context, conn *docdb.DocDB, id string) (*docdb.EventSubscription, error) {
	var eventSubscription *docdb.EventSubscription

	input := &docdb.DescribeEventSubscriptionsInput{
		SubscriptionName: aws.String(id),
	}

	err := conn.DescribeEventSubscriptionsPagesWithContext(ctx, input, func(page *docdb.DescribeEventSubscriptionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, es := range page.EventSubscriptionsList {
			if es == nil {
				continue
			}

			if aws.StringValue(es.CustSubscriptionId) == id {
				eventSubscription = es
				return false
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, docdb.ErrCodeSubscriptionNotFoundFault) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if eventSubscription == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return eventSubscription, nil
}
