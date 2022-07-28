package dynamodb

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindKinesisDataStreamDestination(ctx context.Context, conn *dynamodb.DynamoDB, streamArn, tableName string) (*dynamodb.KinesisDataStreamDestination, error) {
	input := &dynamodb.DescribeKinesisStreamingDestinationInput{
		TableName: aws.String(tableName),
	}

	output, err := conn.DescribeKinesisStreamingDestinationWithContext(ctx, input)

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, nil
	}

	var result *dynamodb.KinesisDataStreamDestination

	for _, destination := range output.KinesisDataStreamDestinations {
		if destination == nil {
			continue
		}

		if aws.StringValue(destination.StreamArn) == streamArn {
			result = destination
			break
		}
	}

	return result, nil
}

func findTableByName(conn *dynamodb.DynamoDB, tableName string) (*dynamodb.TableDescription, error) {
	input := &dynamodb.DescribeTableInput{
		TableName: aws.String(tableName),
	}

	output, err := conn.DescribeTable(input)

	if err != nil {
		return nil, err
	}

	if output == nil || output.Table == nil {
		return nil, nil
	}

	return output.Table, nil
}

func findGSIByTableNameIndexName(conn *dynamodb.DynamoDB, tableName, indexName string) (*dynamodb.GlobalSecondaryIndexDescription, error) {
	table, err := findTableByName(conn, tableName)

	if err != nil {
		return nil, err
	}

	if table == nil {
		return nil, nil
	}

	for _, gsi := range table.GlobalSecondaryIndexes {
		if aws.StringValue(gsi.IndexName) == indexName {
			return gsi, nil
		}
	}

	return nil, nil
}

func findPITRDescriptionByTableName(conn *dynamodb.DynamoDB, tableName string) (*dynamodb.PointInTimeRecoveryDescription, error) {
	input := &dynamodb.DescribeContinuousBackupsInput{
		TableName: aws.String(tableName),
	}

	output, err := conn.DescribeContinuousBackups(input)

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

func findTTLRDescriptionByTableName(conn *dynamodb.DynamoDB, tableName string) (*dynamodb.TimeToLiveDescription, error) {
	input := &dynamodb.DescribeTimeToLiveInput{
		TableName: aws.String(tableName),
	}

	output, err := conn.DescribeTimeToLive(input)

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
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if status := aws.StringValue(output.ContributorInsightsStatus); status == dynamodb.ContributorInsightsStatusDisabled {
		return nil, &resource.NotFoundError{
			Message:     status,
			LastRequest: input,
		}
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
