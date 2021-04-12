package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func DynamoDBTableByName(conn *dynamodb.DynamoDB, tableName string) (*dynamodb.TableDescription, error) {
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
