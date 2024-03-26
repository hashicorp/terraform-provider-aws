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

func FindTableByName(ctx context.Context, conn *dynamodb.DynamoDB, name string) (*dynamodb.TableDescription, error) {
	input := &dynamodb.DescribeTableInput{
		TableName: aws.String(name),
	}

	output, err := conn.DescribeTableWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, dynamodb.ErrCodeResourceNotFoundException) {
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

func findGSIByTwoPartKey(ctx context.Context, conn *dynamodb.DynamoDB, tableName, indexName string) (*dynamodb.GlobalSecondaryIndexDescription, error) {
	table, err := FindTableByName(ctx, conn, tableName)

	if err != nil {
		return nil, err
	}

	for _, v := range table.GlobalSecondaryIndexes {
		if aws.StringValue(v.IndexName) == indexName {
			return v, nil
		}
	}

	return nil, &retry.NotFoundError{}
}

func findPITRDescriptionByTableName(ctx context.Context, conn *dynamodb.DynamoDB, tableName string) (*dynamodb.PointInTimeRecoveryDescription, error) {
	input := &dynamodb.DescribeContinuousBackupsInput{
		TableName: aws.String(tableName),
	}

	output, err := conn.DescribeContinuousBackupsWithContext(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	if output.ContinuousBackupsDescription == nil || output.ContinuousBackupsDescription.PointInTimeRecoveryDescription == nil {
		return nil, nil
	}

	return output.ContinuousBackupsDescription.PointInTimeRecoveryDescription, nil
}

func findTTLRDescriptionByTableName(ctx context.Context, conn *dynamodb.DynamoDB, tableName string) (*dynamodb.TimeToLiveDescription, error) {
	input := &dynamodb.DescribeTimeToLiveInput{
		TableName: aws.String(tableName),
	}

	output, err := conn.DescribeTimeToLiveWithContext(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	if output.TimeToLiveDescription == nil {
		return nil, nil
	}

	return output.TimeToLiveDescription, nil
}

func FindContributorInsights(ctx context.Context, conn *dynamodb.DynamoDB, tableName, indexName string) (*dynamodb.DescribeContributorInsightsOutput, error) {
	input := &dynamodb.DescribeContributorInsightsInput{
		TableName: aws.String(tableName),
	}

	if indexName != "" {
		input.IndexName = aws.String(indexName)
	}

	output, err := conn.DescribeContributorInsightsWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, dynamodb.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if status := aws.StringValue(output.ContributorInsightsStatus); status == dynamodb.ContributorInsightsStatusDisabled {
		return nil, &retry.NotFoundError{
			Message:     status,
			LastRequest: input,
		}
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func FindTableExportByID(ctx context.Context, conn *dynamodb.DynamoDB, id string) (*dynamodb.DescribeExportOutput, error) {
	input := &dynamodb.DescribeExportInput{
		ExportArn: aws.String(id),
	}

	out, err := conn.DescribeExportWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, dynamodb.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if out == nil || out.ExportDescription == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return out, nil
}
