package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	CreateTableTimeout                  = 2 * time.Minute
	UpdateTableTimeoutTotal             = 60 * time.Minute
	UpdateTableTimeout                  = 20 * time.Minute
	UpdateTableContinuousBackupsTimeout = 20 * time.Minute
	DeleteTableTimeout                  = 10 * time.Minute
)

func DynamoDBTableDeleted(conn *dynamodb.DynamoDB, tableName string) (*dynamodb.TableDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			dynamodb.TableStatusActive,
			dynamodb.TableStatusDeleting,
		},
		Target:  []string{},
		Timeout: DeleteTableTimeout,
		Refresh: DynamoDBTableStatus(conn, tableName),
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*dynamodb.TableDescription); ok {
		return output, err
	}

	return nil, err
}

func DynamoDBReplicaUpdateComplete(conn *dynamodb.DynamoDB, tableName, region string) (*dynamodb.ReplicaDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			dynamodb.ReplicaStatusCreating,
			dynamodb.ReplicaStatusUpdating,
			dynamodb.ReplicaStatusDeleting,
			ReplicaStatusEmptyResult,
			ReplicaStatusNotFound,
		},
		Target: []string{
			dynamodb.ReplicaStatusActive,
		},
		Timeout: UpdateTableTimeout,
		Refresh: DynamoDBReplicaStatus(conn, tableName, region),
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*dynamodb.ReplicaDescription); ok {
		return output, err
	}

	return nil, err
}
