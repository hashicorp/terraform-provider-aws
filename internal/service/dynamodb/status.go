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

func statusReplicaUpdate(ctx context.Context, conn *dynamodb.Client, tableName, region string, optFns ...func(*dynamodb.Options)) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findTableByName(ctx, conn, tableName, optFns...)

		if tfresource.NotFound(err) {
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

func statusReplicaDelete(ctx context.Context, conn *dynamodb.Client, tableName, region string, optFns ...func(*dynamodb.Options)) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findTableByName(ctx, conn, tableName, optFns...)

		if tfresource.NotFound(err) {
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

func statusPITR(ctx context.Context, conn *dynamodb.Client, tableName string, optFns ...func(*dynamodb.Options)) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findPITRByTableName(ctx, conn, tableName, optFns...)

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

func statusSSE(ctx context.Context, conn *dynamodb.Client, tableName string, optFns ...func(*dynamodb.Options)) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findTableByName(ctx, conn, tableName, optFns...)

		if tfresource.NotFound(err) {
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
