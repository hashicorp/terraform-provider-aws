// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamodb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func findTableByName(ctx context.Context, conn *dynamodb.Client, name string, optFns ...func(*dynamodb.Options)) (*awstypes.TableDescription, error) {
	input := &dynamodb.DescribeTableInput{
		TableName: aws.String(name),
	}

	output, err := conn.DescribeTable(ctx, input, optFns...)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Table == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Table, nil
}

func findGSIByTwoPartKey(ctx context.Context, conn *dynamodb.Client, tableName, indexName string) (*awstypes.GlobalSecondaryIndexDescription, error) {
	table, err := findTableByName(ctx, conn, tableName)

	if err != nil {
		return nil, err
	}

	for _, v := range table.GlobalSecondaryIndexes {
		if aws.ToString(v.IndexName) == indexName {
			return &v, nil
		}
	}

	return nil, &retry.NotFoundError{}
}

func findPITRByTableName(ctx context.Context, conn *dynamodb.Client, tableName string, optFns ...func(*dynamodb.Options)) (*awstypes.PointInTimeRecoveryDescription, error) {
	input := &dynamodb.DescribeContinuousBackupsInput{
		TableName: aws.String(tableName),
	}

	output, err := conn.DescribeContinuousBackups(ctx, input, optFns...)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ContinuousBackupsDescription == nil || output.ContinuousBackupsDescription.PointInTimeRecoveryDescription == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ContinuousBackupsDescription.PointInTimeRecoveryDescription, nil
}

func findTTLByTableName(ctx context.Context, conn *dynamodb.Client, tableName string) (*awstypes.TimeToLiveDescription, error) {
	input := &dynamodb.DescribeTimeToLiveInput{
		TableName: aws.String(tableName),
	}

	output, err := conn.DescribeTimeToLive(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.TimeToLiveDescription == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.TimeToLiveDescription, nil
}

func findImportByARN(ctx context.Context, conn *dynamodb.Client, arn string) (*awstypes.ImportTableDescription, error) {
	input := &dynamodb.DescribeImportInput{
		ImportArn: aws.String(arn),
	}

	output, err := conn.DescribeImport(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ImportTableDescription == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ImportTableDescription, nil
}
