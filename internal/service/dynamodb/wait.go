package dynamodb

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	kinesisStreamingDestinationActiveTimeout   = 5 * time.Minute
	kinesisStreamingDestinationDisabledTimeout = 5 * time.Minute
	createTableTimeout                         = 20 * time.Minute
	updateTableTimeoutTotal                    = 60 * time.Minute
	replicaUpdateTimeout                       = 30 * time.Minute
	updateTableTimeout                         = 20 * time.Minute
	updateTableContinuousBackupsTimeout        = 20 * time.Minute
	deleteTableTimeout                         = 10 * time.Minute
	pitrUpdateTimeout                          = 30 * time.Second
	ttlUpdateTimeout                           = 30 * time.Second
)

func waitDynamoDBKinesisStreamingDestinationActive(ctx context.Context, conn *dynamodb.DynamoDB, streamArn, tableName string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{dynamodb.DestinationStatusDisabled, dynamodb.DestinationStatusEnabling},
		Target:  []string{dynamodb.DestinationStatusActive},
		Timeout: kinesisStreamingDestinationActiveTimeout,
		Refresh: statusDynamoDBKinesisStreamingDestination(ctx, conn, streamArn, tableName),
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitDynamoDBKinesisStreamingDestinationDisabled(ctx context.Context, conn *dynamodb.DynamoDB, streamArn, tableName string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{dynamodb.DestinationStatusActive, dynamodb.DestinationStatusDisabling},
		Target:  []string{dynamodb.DestinationStatusDisabled},
		Timeout: kinesisStreamingDestinationDisabledTimeout,
		Refresh: statusDynamoDBKinesisStreamingDestination(ctx, conn, streamArn, tableName),
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitDynamoDBTableActive(conn *dynamodb.DynamoDB, tableName string) (*dynamodb.TableDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			dynamodb.TableStatusCreating,
			dynamodb.TableStatusUpdating,
		},
		Target: []string{
			dynamodb.TableStatusActive,
		},
		Timeout: createTableTimeout,
		Refresh: statusDynamoDBTable(conn, tableName),
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*dynamodb.TableDescription); ok {
		return output, err
	}

	return nil, err
}

func waitDynamoDBTableDeleted(conn *dynamodb.DynamoDB, tableName string) (*dynamodb.TableDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			dynamodb.TableStatusActive,
			dynamodb.TableStatusDeleting,
		},
		Target:  []string{},
		Timeout: deleteTableTimeout,
		Refresh: statusDynamoDBTable(conn, tableName),
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*dynamodb.TableDescription); ok {
		return output, err
	}

	return nil, err
}

func waitDynamoDBReplicaActive(conn *dynamodb.DynamoDB, tableName, region string) (*dynamodb.DescribeTableOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			dynamodb.ReplicaStatusCreating,
			dynamodb.ReplicaStatusUpdating,
			dynamodb.ReplicaStatusDeleting,
		},
		Target: []string{
			dynamodb.ReplicaStatusActive,
		},
		Timeout: replicaUpdateTimeout,
		Refresh: statusDynamoDBReplicaUpdate(conn, tableName, region),
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*dynamodb.DescribeTableOutput); ok {
		return output, err
	}

	return nil, err
}

func waitDynamoDBReplicaDeleted(conn *dynamodb.DynamoDB, tableName, region string) (*dynamodb.DescribeTableOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			dynamodb.ReplicaStatusCreating,
			dynamodb.ReplicaStatusUpdating,
			dynamodb.ReplicaStatusDeleting,
			dynamodb.ReplicaStatusActive,
		},
		Target:  []string{""},
		Timeout: replicaUpdateTimeout,
		Refresh: statusDynamoDBReplicaDelete(conn, tableName, region),
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*dynamodb.DescribeTableOutput); ok {
		return output, err
	}

	return nil, err
}

func waitDynamoDBGSIActive(conn *dynamodb.DynamoDB, tableName, indexName string) (*dynamodb.GlobalSecondaryIndexDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			dynamodb.IndexStatusCreating,
			dynamodb.IndexStatusUpdating,
		},
		Target: []string{
			dynamodb.IndexStatusActive,
		},
		Timeout: updateTableTimeout,
		Refresh: statusDynamoDBGSI(conn, tableName, indexName),
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*dynamodb.GlobalSecondaryIndexDescription); ok {
		return output, err
	}

	return nil, err
}

func waitDynamoDBGSIDeleted(conn *dynamodb.DynamoDB, tableName, indexName string) (*dynamodb.GlobalSecondaryIndexDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			dynamodb.IndexStatusActive,
			dynamodb.IndexStatusDeleting,
			dynamodb.IndexStatusUpdating,
		},
		Target:  []string{},
		Timeout: updateTableTimeout,
		Refresh: statusDynamoDBGSI(conn, tableName, indexName),
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*dynamodb.GlobalSecondaryIndexDescription); ok {
		return output, err
	}

	return nil, err
}

func waitDynamoDBPITRUpdated(conn *dynamodb.DynamoDB, tableName string, toEnable bool) (*dynamodb.PointInTimeRecoveryDescription, error) {
	var pending []string
	target := []string{dynamodb.TimeToLiveStatusDisabled}

	if toEnable {
		pending = []string{
			"ENABLING",
		}
		target = []string{dynamodb.PointInTimeRecoveryStatusEnabled}
	}

	stateConf := &resource.StateChangeConf{
		Pending: pending,
		Target:  target,
		Timeout: pitrUpdateTimeout,
		Refresh: statusDynamoDBPITR(conn, tableName),
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*dynamodb.PointInTimeRecoveryDescription); ok {
		return output, err
	}

	return nil, err
}

func waitDynamoDBTTLUpdated(conn *dynamodb.DynamoDB, tableName string, toEnable bool) (*dynamodb.TimeToLiveDescription, error) {
	pending := []string{
		dynamodb.TimeToLiveStatusEnabled,
		dynamodb.TimeToLiveStatusDisabling,
	}
	target := []string{dynamodb.TimeToLiveStatusDisabled}

	if toEnable {
		pending = []string{
			dynamodb.TimeToLiveStatusDisabled,
			dynamodb.TimeToLiveStatusEnabling,
		}
		target = []string{dynamodb.TimeToLiveStatusEnabled}
	}

	stateConf := &resource.StateChangeConf{
		Pending: pending,
		Target:  target,
		Timeout: ttlUpdateTimeout,
		Refresh: statusDynamoDBTTL(conn, tableName),
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*dynamodb.TimeToLiveDescription); ok {
		return output, err
	}

	return nil, err
}

func waitDynamoDBSSEUpdated(conn *dynamodb.DynamoDB, tableName string) (*dynamodb.TableDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			dynamodb.SSEStatusDisabling,
			dynamodb.SSEStatusEnabling,
			dynamodb.SSEStatusUpdating,
		},
		Target: []string{
			dynamodb.SSEStatusDisabled,
			dynamodb.SSEStatusEnabled,
		},
		Timeout: updateTableTimeout,
		Refresh: statusDynamoDBTableSES(conn, tableName),
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*dynamodb.TableDescription); ok {
		return output, err
	}

	return nil, err
}
