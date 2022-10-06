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
	createTableTimeout                         = 30 * time.Minute
	updateTableTimeoutTotal                    = 60 * time.Minute
	replicaUpdateTimeout                       = 30 * time.Minute
	updateTableTimeout                         = 20 * time.Minute
	updateTableContinuousBackupsTimeout        = 20 * time.Minute
	deleteTableTimeout                         = 10 * time.Minute
	pitrUpdateTimeout                          = 30 * time.Second
	ttlUpdateTimeout                           = 30 * time.Second
)

func maxDuration(a, b time.Duration) time.Duration {
	if a >= b {
		return a
	}

	return b
}

func waitKinesisStreamingDestinationActive(ctx context.Context, conn *dynamodb.DynamoDB, streamArn, tableName string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{dynamodb.DestinationStatusDisabled, dynamodb.DestinationStatusEnabling},
		Target:  []string{dynamodb.DestinationStatusActive},
		Timeout: kinesisStreamingDestinationActiveTimeout,
		Refresh: statusKinesisStreamingDestination(ctx, conn, streamArn, tableName),
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitKinesisStreamingDestinationDisabled(ctx context.Context, conn *dynamodb.DynamoDB, streamArn, tableName string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{dynamodb.DestinationStatusActive, dynamodb.DestinationStatusDisabling},
		Target:  []string{dynamodb.DestinationStatusDisabled},
		Timeout: kinesisStreamingDestinationDisabledTimeout,
		Refresh: statusKinesisStreamingDestination(ctx, conn, streamArn, tableName),
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitTableActive(conn *dynamodb.DynamoDB, tableName string, timeout time.Duration) (*dynamodb.TableDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			dynamodb.TableStatusCreating,
			dynamodb.TableStatusUpdating,
		},
		Target: []string{
			dynamodb.TableStatusActive,
		},
		Timeout: maxDuration(createTableTimeout, timeout),
		Refresh: statusTable(conn, tableName),
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*dynamodb.TableDescription); ok {
		return output, err
	}

	return nil, err
}

func waitTableDeleted(conn *dynamodb.DynamoDB, tableName string, timeout time.Duration) (*dynamodb.TableDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			dynamodb.TableStatusActive,
			dynamodb.TableStatusDeleting,
		},
		Target:  []string{},
		Timeout: maxDuration(deleteTableTimeout, timeout),
		Refresh: statusTable(conn, tableName),
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*dynamodb.TableDescription); ok {
		return output, err
	}

	return nil, err
}

func waitReplicaActive(conn *dynamodb.DynamoDB, tableName, region string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			dynamodb.ReplicaStatusCreating,
			dynamodb.ReplicaStatusUpdating,
			dynamodb.ReplicaStatusDeleting,
		},
		Target: []string{
			dynamodb.ReplicaStatusActive,
		},
		Timeout: maxDuration(replicaUpdateTimeout, timeout),
		Refresh: statusReplicaUpdate(conn, tableName, region),
	}

	_, err := stateConf.WaitForState()

	return err
}

func waitReplicaDeleted(conn *dynamodb.DynamoDB, tableName, region string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			dynamodb.ReplicaStatusCreating,
			dynamodb.ReplicaStatusUpdating,
			dynamodb.ReplicaStatusDeleting,
			dynamodb.ReplicaStatusActive,
		},
		Target:  []string{""},
		Timeout: maxDuration(replicaUpdateTimeout, timeout),
		Refresh: statusReplicaDelete(conn, tableName, region),
	}

	_, err := stateConf.WaitForState()

	return err
}

func waitGSIActive(conn *dynamodb.DynamoDB, tableName, indexName string, timeout time.Duration) (*dynamodb.GlobalSecondaryIndexDescription, error) { //nolint:unparam
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			dynamodb.IndexStatusCreating,
			dynamodb.IndexStatusUpdating,
		},
		Target: []string{
			dynamodb.IndexStatusActive,
		},
		Timeout: maxDuration(updateTableTimeout, timeout),
		Refresh: statusGSI(conn, tableName, indexName),
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*dynamodb.GlobalSecondaryIndexDescription); ok {
		return output, err
	}

	return nil, err
}

func waitGSIDeleted(conn *dynamodb.DynamoDB, tableName, indexName string, timeout time.Duration) (*dynamodb.GlobalSecondaryIndexDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			dynamodb.IndexStatusActive,
			dynamodb.IndexStatusDeleting,
			dynamodb.IndexStatusUpdating,
		},
		Target:  []string{},
		Timeout: maxDuration(updateTableTimeout, timeout),
		Refresh: statusGSI(conn, tableName, indexName),
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*dynamodb.GlobalSecondaryIndexDescription); ok {
		return output, err
	}

	return nil, err
}

func waitPITRUpdated(conn *dynamodb.DynamoDB, tableName string, toEnable bool, timeout time.Duration) (*dynamodb.PointInTimeRecoveryDescription, error) {
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
		Timeout: maxDuration(pitrUpdateTimeout, timeout),
		Refresh: statusPITR(conn, tableName),
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*dynamodb.PointInTimeRecoveryDescription); ok {
		return output, err
	}

	return nil, err
}

func waitTTLUpdated(conn *dynamodb.DynamoDB, tableName string, toEnable bool, timeout time.Duration) (*dynamodb.TimeToLiveDescription, error) {
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
		Timeout: maxDuration(ttlUpdateTimeout, timeout),
		Refresh: statusTTL(conn, tableName),
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*dynamodb.TimeToLiveDescription); ok {
		return output, err
	}

	return nil, err
}

func waitSSEUpdated(conn *dynamodb.DynamoDB, tableName string, timeout time.Duration) (*dynamodb.TableDescription, error) {
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
		Timeout: maxDuration(updateTableTimeout, timeout),
		Refresh: statusTableSES(conn, tableName),
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*dynamodb.TableDescription); ok {
		return output, err
	}

	return nil, err
}

func waitContributorInsightsCreated(ctx context.Context, conn *dynamodb.DynamoDB, tableName, indexName string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{dynamodb.ContributorInsightsStatusEnabling},
		Target:  []string{dynamodb.ContributorInsightsStatusEnabled},
		Timeout: timeout,
		Refresh: statusContributorInsights(ctx, conn, tableName, indexName),
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitContributorInsightsDeleted(ctx context.Context, conn *dynamodb.DynamoDB, tableName, indexName string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{dynamodb.ContributorInsightsStatusDisabling},
		Target:  []string{},
		Timeout: timeout,
		Refresh: statusContributorInsights(ctx, conn, tableName, indexName),
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}
