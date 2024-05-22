// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dynamodb

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	awstypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
)

const (
	createTableTimeout                         = 30 * time.Minute
	deleteTableTimeout                         = 10 * time.Minute
	kinesisStreamingDestinationActiveTimeout   = 5 * time.Minute
	kinesisStreamingDestinationDisabledTimeout = 5 * time.Minute
	pitrUpdateTimeout                          = 30 * time.Second
	replicaUpdateTimeout                       = 30 * time.Minute
	ttlUpdateTimeout                           = 30 * time.Second
	updateTableContinuousBackupsTimeout        = 20 * time.Minute
	updateTableTimeout                         = 20 * time.Minute
	updateTableTimeoutTotal                    = 60 * time.Minute
)

func waitTableActive(ctx context.Context, conn *dynamodb.Client, tableName string, timeout time.Duration) (*awstypes.TableDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.TableStatusCreating, awstypes.TableStatusUpdating),
		Target:  enum.Slice(awstypes.TableStatusActive),
		Refresh: statusTable(ctx, conn, tableName),
		Timeout: max(createTableTimeout, timeout),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TableDescription); ok {
		return output, err
	}

	return nil, err
}

func waitTableDeleted(ctx context.Context, conn *dynamodb.Client, tableName string, timeout time.Duration) (*awstypes.TableDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.TableStatusActive, awstypes.TableStatusDeleting),
		Target:  []string{},
		Refresh: statusTable(ctx, conn, tableName),
		Timeout: max(deleteTableTimeout, timeout),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TableDescription); ok {
		return output, err
	}

	return nil, err
}

func waitImportComplete(ctx context.Context, conn *dynamodb.Client, importARN string, timeout time.Duration) (*awstypes.ImportTableDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ImportStatusInProgress),
		Target:  enum.Slice(awstypes.ImportStatusCompleted),
		Refresh: statusImport(ctx, conn, importARN),
		Timeout: max(createTableTimeout, timeout),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.ImportTableDescription); ok {
		return output, err
	}

	return nil, err
}

func waitReplicaActive(ctx context.Context, conn *dynamodb.Client, tableName, region string, timeout time.Duration, optFns ...func(*dynamodb.Options)) (*awstypes.TableDescription, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ReplicaStatusCreating, awstypes.ReplicaStatusUpdating, awstypes.ReplicaStatusDeleting),
		Target:  enum.Slice(awstypes.ReplicaStatusActive),
		Refresh: statusReplicaUpdate(ctx, conn, tableName, region, optFns...),
		Timeout: max(replicaUpdateTimeout, timeout),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TableDescription); ok {
		return output, err
	}

	return nil, err
}

func waitReplicaDeleted(ctx context.Context, conn *dynamodb.Client, tableName, region string, timeout time.Duration, optFns ...func(*dynamodb.Options)) (*awstypes.TableDescription, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.ReplicaStatusCreating,
			awstypes.ReplicaStatusUpdating,
			awstypes.ReplicaStatusDeleting,
			awstypes.ReplicaStatusActive,
		),
		Target:  []string{},
		Refresh: statusReplicaDelete(ctx, conn, tableName, region, optFns...),
		Timeout: max(replicaUpdateTimeout, timeout),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TableDescription); ok {
		return output, err
	}

	return nil, err
}

func waitGSIActive(ctx context.Context, conn *dynamodb.Client, tableName, indexName string, timeout time.Duration) (*awstypes.GlobalSecondaryIndexDescription, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.IndexStatusCreating, awstypes.IndexStatusUpdating),
		Target:  enum.Slice(awstypes.IndexStatusActive),
		Refresh: statusGSI(ctx, conn, tableName, indexName),
		Timeout: max(updateTableTimeout, timeout),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.GlobalSecondaryIndexDescription); ok {
		return output, err
	}

	return nil, err
}

func waitGSIDeleted(ctx context.Context, conn *dynamodb.Client, tableName, indexName string, timeout time.Duration) (*awstypes.GlobalSecondaryIndexDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.IndexStatusActive, awstypes.IndexStatusDeleting, awstypes.IndexStatusUpdating),
		Target:  []string{},
		Refresh: statusGSI(ctx, conn, tableName, indexName),
		Timeout: max(updateTableTimeout, timeout),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.GlobalSecondaryIndexDescription); ok {
		return output, err
	}

	return nil, err
}

func waitPITRUpdated(ctx context.Context, conn *dynamodb.Client, tableName string, toEnable bool, timeout time.Duration, optFns ...func(*dynamodb.Options)) (*awstypes.PointInTimeRecoveryDescription, error) {
	var pending []string
	target := enum.Slice(awstypes.PointInTimeRecoveryStatusDisabled)

	if toEnable {
		pending = enum.Slice(
			awstypes.PointInTimeRecoveryStatus("ENABLING"), // "ENABLING" const not available for PITR
			awstypes.PointInTimeRecoveryStatusDisabled,     // reports say it can get in fast enough to be in this state
		)
		target = enum.Slice(awstypes.PointInTimeRecoveryStatusEnabled)
	}

	stateConf := &retry.StateChangeConf{
		Pending:    pending,
		Target:     target,
		Refresh:    statusPITR(ctx, conn, tableName, optFns...),
		Timeout:    max(pitrUpdateTimeout, timeout),
		MinTimeout: 15 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.PointInTimeRecoveryDescription); ok {
		return output, err
	}

	return nil, err
}

func waitTTLUpdated(ctx context.Context, conn *dynamodb.Client, tableName string, toEnable bool, timeout time.Duration) (*awstypes.TimeToLiveDescription, error) {
	pending := enum.Slice(awstypes.TimeToLiveStatusEnabled, awstypes.TimeToLiveStatusDisabling)
	target := enum.Slice(awstypes.TimeToLiveStatusDisabled)

	if toEnable {
		pending = enum.Slice(awstypes.TimeToLiveStatusDisabled, awstypes.TimeToLiveStatusEnabling)
		target = enum.Slice(awstypes.TimeToLiveStatusEnabled)
	}

	stateConf := &retry.StateChangeConf{
		Pending: pending,
		Target:  target,
		Timeout: max(ttlUpdateTimeout, timeout),
		Refresh: statusTTL(ctx, conn, tableName),
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TimeToLiveDescription); ok {
		return output, err
	}

	return nil, err
}

func waitSSEUpdated(ctx context.Context, conn *dynamodb.Client, tableName string, timeout time.Duration) (*awstypes.TableDescription, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.SSEStatusDisabling, awstypes.SSEStatusEnabling, awstypes.SSEStatusUpdating),
		Target:  enum.Slice(awstypes.SSEStatusDisabled, awstypes.SSEStatusEnabled),
		Refresh: statusSSE(ctx, conn, tableName),
		Timeout: max(updateTableTimeout, timeout),
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TableDescription); ok {
		return output, err
	}

	return nil, err
}

func waitReplicaSSEUpdated(ctx context.Context, conn *dynamodb.Client, region, tableName string, timeout time.Duration) (*awstypes.TableDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.SSEStatusDisabling, awstypes.SSEStatusEnabling, awstypes.SSEStatusUpdating),
		Target:  enum.Slice(awstypes.SSEStatusDisabled, awstypes.SSEStatusEnabled),
		Refresh: statusSSE(ctx, conn, tableName, func(o *dynamodb.Options) {
			o.Region = region
		}),
		Timeout: max(updateTableTimeout, timeout),
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.TableDescription); ok {
		return output, err
	}

	return nil, err
}
