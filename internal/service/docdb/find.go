package docdb

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/docdb"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func findGlobalClusterByArn(ctx context.Context, conn *docdb.DocDB, dbClusterARN string) (*docdb.GlobalCluster, error) {
	var globalCluster *docdb.GlobalCluster

	input := &docdb.DescribeGlobalClustersInput{
		Filters: []*docdb.Filter{
			{
				Name:   aws.String("db-cluster-id"),
				Values: []*string{aws.String(dbClusterARN)},
			},
		},
	}

	log.Printf("[DEBUG] Reading DocDB Global Clusters: %s", input)
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

func FindGlobalClusterById(ctx context.Context, conn *docdb.DocDB, globalClusterID string) (*docdb.GlobalCluster, error) {
	var globalCluster *docdb.GlobalCluster

	input := &docdb.DescribeGlobalClustersInput{
		GlobalClusterIdentifier: aws.String(globalClusterID),
	}

	log.Printf("[DEBUG] Reading DocDB Global Cluster (%s): %s", globalClusterID, input)
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

func findGlobalClusterIdByArn(ctx context.Context, conn *docdb.DocDB, arn string) string {
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

func FindEventSubscriptionByID(ctx context.Context, conn *docdb.DocDB, id string) (*docdb.EventSubscription, error) {
	var eventSubscription *docdb.EventSubscription

	input := &docdb.DescribeEventSubscriptionsInput{
		SubscriptionName: aws.String(id),
	}

	log.Printf("[DEBUG] Reading DocDB Event Subscription (%s): %s", id, input)
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
		return nil, &resource.NotFoundError{
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
