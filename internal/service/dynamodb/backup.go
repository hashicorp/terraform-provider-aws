package dynamodb

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func listBackupsPages(ctx context.Context, conn *dynamodb.DynamoDB, input *dynamodb.ListBackupsInput, fn func(*dynamodb.ListBackupsOutput, bool) bool) error { //nolint:unused // This function is called from a sweeper.
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
