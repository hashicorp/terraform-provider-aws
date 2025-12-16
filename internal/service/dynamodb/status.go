// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package dynamodb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
)

func statusTable(conn *dynamodb.Client, tableName string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findTableByName(ctx, conn, tableName)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.TableStatus), nil
	}
}

func statusTableWarmThroughput(conn *dynamodb.Client, tableName string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findTableByName(ctx, conn, tableName)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if output == nil || output.WarmThroughput == nil {
			return nil, "", nil
		}

		return output, string(output.WarmThroughput.Status), nil
	}
}

func statusImport(conn *dynamodb.Client, importARN string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findImportByARN(ctx, conn, importARN)

		if retry.NotFound(err) {
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

func statusReplicaUpdate(conn *dynamodb.Client, tableName, region string, optFns ...func(*dynamodb.Options)) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findTableByName(ctx, conn, tableName, optFns...)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		for _, v := range output.Replicas {
			if aws.ToString(v.RegionName) == region {
				return output, string(v.ReplicaStatus), nil
			}
		}

		return output, string(awstypes.ReplicaStatusCreating), nil
	}
}

func statusReplicaDelete(conn *dynamodb.Client, tableName, region string, optFns ...func(*dynamodb.Options)) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findTableByName(ctx, conn, tableName, optFns...)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		for _, v := range output.Replicas {
			if aws.ToString(v.RegionName) == region {
				return output, string(v.ReplicaStatus), nil
			}
		}

		return nil, "", nil
	}
}

func statusGSI(conn *dynamodb.Client, tableName, indexName string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findGSIByTwoPartKey(ctx, conn, tableName, indexName)

		if retry.NotFound(err) {
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

func statusAllGSI(conn *dynamodb.Client, tableName string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findTableByName(ctx, conn, tableName)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if output == nil {
			return nil, "", nil
		}

		for _, g := range output.GlobalSecondaryIndexes {
			if g.IndexStatus != awstypes.IndexStatusActive {
				return output, string(g.IndexStatus), nil
			}
		}

		return output, string(awstypes.IndexStatusActive), nil
	}
}

func statusGSIWarmThroughput(conn *dynamodb.Client, tableName, indexName string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findGSIByTwoPartKey(ctx, conn, tableName, indexName)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if output == nil || output.WarmThroughput == nil {
			return nil, "", nil
		}

		return output, string(output.WarmThroughput.Status), nil
	}
}

func statusPITR(conn *dynamodb.Client, tableName string, optFns ...func(*dynamodb.Options)) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findPITRByTableName(ctx, conn, tableName, optFns...)

		if retry.NotFound(err) {
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

func statusTTL(conn *dynamodb.Client, tableName string) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findTTLByTableName(ctx, conn, tableName)

		if retry.NotFound(err) {
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

func statusSSE(conn *dynamodb.Client, tableName string, optFns ...func(*dynamodb.Options)) retry.StateRefreshFunc {
	return func(ctx context.Context) (any, string, error) {
		output, err := findTableByName(ctx, conn, tableName, optFns...)

		if retry.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		// Disabling SSE returns null SSEDescription.
		if output.SSEDescription == nil {
			return output, string(awstypes.SSEStatusDisabled), nil
		}

		return output, string(output.SSEDescription.Status), nil
	}
}
