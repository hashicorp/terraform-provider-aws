package dynamodb

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
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

func waitTableActive(ctx context.Context, conn *dynamodb.DynamoDB, tableName string, timeout time.Duration) (*dynamodb.TableDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{dynamodb.TableStatusCreating, dynamodb.TableStatusUpdating},
		Target:  []string{dynamodb.TableStatusActive},
		Timeout: maxDuration(createTableTimeout, timeout),
		Refresh: statusTable(ctx, conn, tableName),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*dynamodb.TableDescription); ok {
		return output, err
	}

	return nil, err
}

func waitTableDeleted(ctx context.Context, conn *dynamodb.DynamoDB, tableName string, timeout time.Duration) (*dynamodb.TableDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{dynamodb.TableStatusActive, dynamodb.TableStatusDeleting},
		Target:  []string{},
		Timeout: maxDuration(deleteTableTimeout, timeout),
		Refresh: statusTable(ctx, conn, tableName),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*dynamodb.TableDescription); ok {
		return output, err
	}

	return nil, err
}

func waitReplicaActive(ctx context.Context, conn *dynamodb.DynamoDB, tableName, region string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{dynamodb.ReplicaStatusCreating, dynamodb.ReplicaStatusUpdating, dynamodb.ReplicaStatusDeleting},
		Target:  []string{dynamodb.ReplicaStatusActive},
		Timeout: maxDuration(replicaUpdateTimeout, timeout),
		Refresh: statusReplicaUpdate(ctx, conn, tableName, region),
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitReplicaDeleted(ctx context.Context, conn *dynamodb.DynamoDB, tableName, region string, timeout time.Duration) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			dynamodb.ReplicaStatusCreating,
			dynamodb.ReplicaStatusUpdating,
			dynamodb.ReplicaStatusDeleting,
			dynamodb.ReplicaStatusActive,
		},
		Target:  []string{""},
		Timeout: maxDuration(replicaUpdateTimeout, timeout),
		Refresh: statusReplicaDelete(ctx, conn, tableName, region),
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitGSIActive(ctx context.Context, conn *dynamodb.DynamoDB, tableName, indexName string, timeout time.Duration) (*dynamodb.GlobalSecondaryIndexDescription, error) { //nolint:unparam
	stateConf := &resource.StateChangeConf{
		Pending: []string{dynamodb.IndexStatusCreating, dynamodb.IndexStatusUpdating},
		Target:  []string{dynamodb.IndexStatusActive},
		Timeout: maxDuration(updateTableTimeout, timeout),
		Refresh: statusGSI(ctx, conn, tableName, indexName),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*dynamodb.GlobalSecondaryIndexDescription); ok {
		return output, err
	}

	return nil, err
}

func waitGSIDeleted(ctx context.Context, conn *dynamodb.DynamoDB, tableName, indexName string, timeout time.Duration) (*dynamodb.GlobalSecondaryIndexDescription, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{dynamodb.IndexStatusActive, dynamodb.IndexStatusDeleting, dynamodb.IndexStatusUpdating},
		Target:  []string{},
		Timeout: maxDuration(updateTableTimeout, timeout),
		Refresh: statusGSI(ctx, conn, tableName, indexName),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*dynamodb.GlobalSecondaryIndexDescription); ok {
		return output, err
	}

	return nil, err
}

func waitPITRUpdated(ctx context.Context, conn *dynamodb.DynamoDB, tableName string, toEnable bool, timeout time.Duration) (*dynamodb.PointInTimeRecoveryDescription, error) {
	var pending []string
	target := []string{dynamodb.PointInTimeRecoveryStatusDisabled}

	if toEnable {
		pending = []string{
			dynamodb.TimeToLiveStatusEnabling,          // "ENABLING" const not available for PITR
			dynamodb.PointInTimeRecoveryStatusDisabled, // reports say it can get in fast enough to be in this state
		}
		target = []string{dynamodb.PointInTimeRecoveryStatusEnabled}
	}

	stateConf := &resource.StateChangeConf{
		Pending:    pending,
		Target:     target,
		Timeout:    maxDuration(pitrUpdateTimeout, timeout),
		Refresh:    statusPITR(ctx, conn, tableName),
		MinTimeout: 15 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*dynamodb.PointInTimeRecoveryDescription); ok {
		return output, err
	}

	return nil, err
}

func waitTTLUpdated(ctx context.Context, conn *dynamodb.DynamoDB, tableName string, toEnable bool, timeout time.Duration) (*dynamodb.TimeToLiveDescription, error) {
	pending := []string{dynamodb.TimeToLiveStatusEnabled, dynamodb.TimeToLiveStatusDisabling}
	target := []string{dynamodb.TimeToLiveStatusDisabled}

	if toEnable {
		pending = []string{dynamodb.TimeToLiveStatusDisabled, dynamodb.TimeToLiveStatusEnabling}
		target = []string{dynamodb.TimeToLiveStatusEnabled}
	}

	stateConf := &resource.StateChangeConf{
		Pending: pending,
		Target:  target,
		Timeout: maxDuration(ttlUpdateTimeout, timeout),
		Refresh: statusTTL(ctx, conn, tableName),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*dynamodb.TimeToLiveDescription); ok {
		return output, err
	}

	return nil, err
}

func waitSSEUpdated(ctx context.Context, conn *dynamodb.DynamoDB, tableName string, timeout time.Duration) (*dynamodb.TableDescription, error) { //nolint:unparam
	stateConf := &resource.StateChangeConf{
		Delay:   30 * time.Second,
		Pending: []string{dynamodb.SSEStatusDisabling, dynamodb.SSEStatusEnabling, dynamodb.SSEStatusUpdating},
		Target:  []string{dynamodb.SSEStatusDisabled, dynamodb.SSEStatusEnabled},
		Timeout: maxDuration(updateTableTimeout, timeout),
		Refresh: statusTableSES(ctx, conn, tableName),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*dynamodb.TableDescription); ok {
		return output, err
	}

	return nil, err
}

func waitReplicaSSEUpdated(ctx context.Context, client *conns.AWSClient, region string, tableName string, timeout time.Duration) (*dynamodb.TableDescription, error) {
	sess, err := conns.NewSessionForRegion(&client.DynamoDBConn().Config, region, client.TerraformVersion)
	if err != nil {
		return nil, fmt.Errorf("creating session for region %q: %w", region, err)
	}

	conn := dynamodb.New(sess)
	stateConf := &resource.StateChangeConf{
		Delay:   30 * time.Second,
		Pending: []string{dynamodb.SSEStatusDisabling, dynamodb.SSEStatusEnabling, dynamodb.SSEStatusUpdating},
		Target:  []string{dynamodb.SSEStatusDisabled, dynamodb.SSEStatusEnabled},
		Timeout: maxDuration(updateTableTimeout, timeout),
		Refresh: statusTableSES(ctx, conn, tableName),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

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
