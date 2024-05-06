// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamodb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func statusTable(ctx context.Context, conn *dynamodb.Client, tableName string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findTableByName(ctx, conn, tableName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if output == nil {
			return nil, "", nil
		}

		return output, string(output.TableStatus), nil
	}
}

func statusImport(ctx context.Context, conn *dynamodb.Client, importARN string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findImportByARN(ctx, conn, importARN)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if output == nil {
			return nil, "", nil
		}

		return output, string(output.ImportStatus), nil
	}
}

func statusReplicaUpdate(ctx context.Context, conn *dynamodb.Client, tableName, region string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		result, err := conn.DescribeTableWithContext(ctx, &dynamodb.DescribeTableInput{
			TableName: aws.String(tableName),
		})
		if err != nil {
			return nil, "", err
		}

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

func statusReplicaDelete(ctx context.Context, conn *dynamodb.DynamoDB, tableName, region string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		result, err := conn.DescribeTableWithContext(ctx, &dynamodb.DescribeTableInput{
			TableName: aws.String(tableName),
		})
		if err != nil {
			return nil, "", err
		}

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

func statusGSI(ctx context.Context, conn *dynamodb.Client, tableName, indexName string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findGSIByTwoPartKey(ctx, conn, tableName, indexName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if output == nil {
			return nil, "", nil
		}

		return output, string(output.IndexStatus), nil
	}
}

func statusPITR(ctx context.Context, conn *dynamodb.Client, tableName string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findPITRByTableName(ctx, conn, tableName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if output == nil {
			return nil, "", nil
		}

		return output, string(output.PointInTimeRecoveryStatus), nil
	}
}

func statusTTL(ctx context.Context, conn *dynamodb.Client, tableName string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findTTLByTableName(ctx, conn, tableName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if output == nil {
			return nil, "", nil
		}

		return output, string(output.TimeToLiveStatus), nil
	}
}

func statusTableSES(ctx context.Context, conn *dynamodb.Client, tableName string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findTableByName(ctx, conn, tableName)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if output == nil {
			return nil, "", nil
		}

		// Disabling SSE returns null SSEDescription
		if output.SSEDescription == nil {
			return output, string(awstypes.SSEStatusDisabled), nil
		}

		return output, string(output.SSEDescription.Status), nil
	}
}
