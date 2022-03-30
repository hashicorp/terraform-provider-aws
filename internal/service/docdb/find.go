package docdb

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/docdb"
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
