package dynamodb

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func listBackupsPages(conn *dynamodb.DynamoDB, input *dynamodb.ListBackupsInput, fn func(*dynamodb.ListBackupsOutput, bool) bool) error { //nolint:deadcode // used in sweeper
	return listBackupsPagesWithContext(context.Background(), conn, input, fn)
}

func listBackupsPagesWithContext(ctx context.Context, conn *dynamodb.DynamoDB, input *dynamodb.ListBackupsInput, fn func(*dynamodb.ListBackupsOutput, bool) bool) error {
	for {
		output, err := conn.ListBackupsWithContext(ctx, input)
		if err != nil {
			return err
		}

		lastPage := aws.StringValue(output.LastEvaluatedBackupArn) == ""
		if !fn(output, lastPage) || lastPage {
			break
		}

		input.ExclusiveStartBackupArn = output.LastEvaluatedBackupArn
	}
	return nil
}
