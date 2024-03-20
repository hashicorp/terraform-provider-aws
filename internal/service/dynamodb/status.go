// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamodb

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func statusTable(ctx context.Context, conn *dynamodb.DynamoDB, tableName string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		table, err := FindTableByName(ctx, conn, tableName)

		if tfresource.NotFound(err) {
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

func statusImport(ctx context.Context, conn *dynamodb.DynamoDB, importArn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		describeImportInput := &dynamodb.DescribeImportInput{
			ImportArn: &importArn,
		}
		output, err := conn.DescribeImportWithContext(ctx, describeImportInput)

		if tfawserr.ErrCodeEquals(err, dynamodb.ErrCodeResourceNotFoundException) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.ImportTableDescription.ImportStatus), nil
	}
}

func statusReplicaUpdate(ctx context.Context, conn *dynamodb.DynamoDB, tableName, region string) retry.StateRefreshFunc {
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

func statusGSI(ctx context.Context, conn *dynamodb.DynamoDB, tableName, indexName string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		gsi, err := findGSIByTwoPartKey(ctx, conn, tableName, indexName)

		if tfresource.NotFound(err) {
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

func statusPITR(ctx context.Context, conn *dynamodb.DynamoDB, tableName string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		pitr, err := findPITRDescriptionByTableName(ctx, conn, tableName)

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

func statusTTL(ctx context.Context, conn *dynamodb.DynamoDB, tableName string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		ttl, err := findTTLRDescriptionByTableName(ctx, conn, tableName)

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

func statusTableSES(ctx context.Context, conn *dynamodb.DynamoDB, tableName string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		table, err := FindTableByName(ctx, conn, tableName)

		if tfresource.NotFound(err) {
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

func statusContributorInsights(ctx context.Context, conn *dynamodb.DynamoDB, tableName, indexName string) retry.StateRefreshFunc {
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

func statusTableExport(ctx context.Context, conn *dynamodb.DynamoDB, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		out, err := FindTableExportByID(ctx, conn, id)
		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		if out.ExportDescription == nil {
			return nil, "", nil
		}

		return out, aws.StringValue(out.ExportDescription.ExportStatus), nil
	}
}
