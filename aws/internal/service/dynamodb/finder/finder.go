package finder

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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

func DynamoDBReplicaByTableNameRegion(conn *dynamodb.DynamoDB, tableName, region string) (*dynamodb.ReplicaDescription, error) {
	table, err := DynamoDBTableByName(conn, tableName)

	if err != nil {
		return nil, err
	}

	if table == nil {
		return nil, &resource.NotFoundError{
			Message: dynamodb.ErrCodeTableNotFoundException,
		}
	}

	for _, replica := range table.Replicas {
		if aws.StringValue(replica.RegionName) == region {
			return replica, nil
		}
	}

	return nil, nil
}
