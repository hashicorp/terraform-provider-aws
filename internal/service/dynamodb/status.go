package dynamodb

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func statusKinesisStreamingDestination(ctx context.Context, conn *dynamodb.DynamoDB, streamArn, tableName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		result, err := FindKinesisDataStreamDestination(ctx, conn, streamArn, tableName)

		if err != nil {
			return nil, "", err
		}

		if result == nil {
			return nil, "", nil
		}

		return result, aws.StringValue(result.DestinationStatus), nil
	}
}

func statusTable(conn *dynamodb.DynamoDB, tableName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		table, err := findTableByName(conn, tableName)

		if tfawserr.ErrCodeEquals(err, dynamodb.ErrCodeResourceNotFoundException) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if table == nil {
			return nil, "", nil
		}

		return table, aws.StringValue(table.TableStatus), nil
	}
}

func statusReplicaUpdate(conn *dynamodb.DynamoDB, tableName, region string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		result, err := conn.DescribeTable(&dynamodb.DescribeTableInput{
			TableName: aws.String(tableName),
		})
		if err != nil {
			return 42, "", err
		}
		log.Printf("[DEBUG] DynamoDB replicas: %s", result.Table.Replicas)

		var targetReplica *dynamodb.ReplicaDescription

		for _, replica := range result.Table.Replicas {
			if aws.StringValue(replica.RegionName) == region {
				targetReplica = replica
				break
			}
		}

		if targetReplica == nil {
			return result, dynamodb.ReplicaStatusCreating, nil
		}

		return result, aws.StringValue(targetReplica.ReplicaStatus), nil
	}
}

func statusReplicaDelete(conn *dynamodb.DynamoDB, tableName, region string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		result, err := conn.DescribeTable(&dynamodb.DescribeTableInput{
			TableName: aws.String(tableName),
		})
		if err != nil {
			return 42, "", err
		}

		log.Printf("[DEBUG] all replicas for waiting: %s", result.Table.Replicas)
		var targetReplica *dynamodb.ReplicaDescription

		for _, replica := range result.Table.Replicas {
			if aws.StringValue(replica.RegionName) == region {
				targetReplica = replica
				break
			}
		}

		if targetReplica == nil {
			return result, "", nil
		}

		return result, aws.StringValue(targetReplica.ReplicaStatus), nil
	}
}

func statusGSI(conn *dynamodb.DynamoDB, tableName, indexName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		gsi, err := findGSIByTableNameIndexName(conn, tableName, indexName)

		if tfawserr.ErrCodeEquals(err, dynamodb.ErrCodeResourceNotFoundException) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if gsi == nil {
			return nil, "", nil
		}

		return gsi, aws.StringValue(gsi.IndexStatus), nil
	}
}

func statusPITR(conn *dynamodb.DynamoDB, tableName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		pitr, err := findPITRDescriptionByTableName(conn, tableName)

		if tfawserr.ErrCodeEquals(err, dynamodb.ErrCodeResourceNotFoundException) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if pitr == nil {
			return nil, "", nil
		}

		return pitr, aws.StringValue(pitr.PointInTimeRecoveryStatus), nil
	}
}

func statusTTL(conn *dynamodb.DynamoDB, tableName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		ttl, err := findTTLRDescriptionByTableName(conn, tableName)

		if tfawserr.ErrCodeEquals(err, dynamodb.ErrCodeResourceNotFoundException) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if ttl == nil {
			return nil, "", nil
		}

		return ttl, aws.StringValue(ttl.TimeToLiveStatus), nil
	}
}

func statusTableSES(conn *dynamodb.DynamoDB, tableName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		table, err := findTableByName(conn, tableName)

		if tfawserr.ErrCodeEquals(err, dynamodb.ErrCodeResourceNotFoundException) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if table == nil {
			return nil, "", nil
		}

		// Disabling SSE returns null SSEDescription
		if table.SSEDescription == nil {
			return table, dynamodb.SSEStatusDisabled, nil
		}

		return table, aws.StringValue(table.SSEDescription.Status), nil
	}
}

func statusContributorInsights(ctx context.Context, conn *dynamodb.DynamoDB, tableName, indexName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		insight, err := FindContributorInsights(ctx, conn, tableName, indexName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if insight == nil {
			return nil, "", nil
		}

		return insight, aws.StringValue(insight.ContributorInsightsStatus), nil
	}
}
