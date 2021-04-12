package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/dynamodb/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

const (
	ReplicaStatusNotFound    = "NOT_FOUND"
	ReplicaStatusEmptyResult = "EMPTY_RESULT"
)

func DynamoDBTableStatus(conn *dynamodb.DynamoDB, tableName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		table, err := finder.DynamoDBTableByName(conn, tableName)

		if tfawserr.ErrCodeEquals(err, dynamodb.ErrCodeResourceNotFoundException) {
			return nil, "", nil
		}

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

func DynamoDBReplicaStatus(conn *dynamodb.DynamoDB, tableName, region string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		replica, err := finder.DynamoDBReplicaByTableNameRegion(conn, tableName, region)

		if tfresource.NotFound(err) {
			return nil, ReplicaStatusNotFound, nil
		}

		if err != nil {
			return nil, "", err
		}

		if replica == nil {
			return nil, ReplicaStatusEmptyResult, nil
		}

		return replica, aws.StringValue(replica.ReplicaStatus), nil
	}
}
