// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/glue"
	awstypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
)

const (
	// Maximum amount of time to wait for an Operation to return Deleted
	iamPropagationTimeout         = 2 * time.Minute
	schemaAvailableTimeout        = 2 * time.Minute
	schemaDeleteTimeout           = 2 * time.Minute
	schemaVersionAvailableTimeout = 2 * time.Minute
)

// waitSchemaAvailable waits for a Schema to return Available
func waitSchemaAvailable(ctx context.Context, conn *glue.Client, registryID string) (*glue.GetSchemaOutput, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.SchemaStatusPending),
		Target:  enum.Slice(awstypes.SchemaStatusAvailable),
		Refresh: statusSchema(ctx, conn, registryID),
		Timeout: schemaAvailableTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*glue.GetSchemaOutput); ok {
		return output, err
	}

	return nil, err
}

// waitSchemaDeleted waits for a Schema to return Deleted
func waitSchemaDeleted(ctx context.Context, conn *glue.Client, registryID string) (*glue.GetSchemaOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.SchemaStatusDeleting),
		Target:  []string{},
		Refresh: statusSchema(ctx, conn, registryID),
		Timeout: schemaDeleteTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*glue.GetSchemaOutput); ok {
		return output, err
	}

	return nil, err
}

// waitSchemaVersionAvailable waits for a Schema to return Available
func waitSchemaVersionAvailable(ctx context.Context, conn *glue.Client, registryID string) (*glue.GetSchemaVersionOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.SchemaVersionStatusPending),
		Target:  enum.Slice(awstypes.SchemaVersionStatusAvailable),
		Refresh: statusSchemaVersion(ctx, conn, registryID),
		Timeout: schemaVersionAvailableTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*glue.GetSchemaVersionOutput); ok {
		return output, err
	}

	return nil, err
}

// waitTriggerCreated waits for a Trigger to return Created
func waitTriggerCreated(ctx context.Context, conn *glue.Client, triggerName string, timeout time.Duration) (*glue.GetTriggerOutput, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.TriggerStateActivating,
			awstypes.TriggerStateCreating,
			awstypes.TriggerStateUpdating,
		),
		Target: enum.Slice(
			awstypes.TriggerStateActivated,
			awstypes.TriggerStateCreated,
		),
		Refresh: statusTrigger(ctx, conn, triggerName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*glue.GetTriggerOutput); ok {
		return output, err
	}

	return nil, err
}

// waitTriggerDeleted waits for a Trigger to return Deleted
func waitTriggerDeleted(ctx context.Context, conn *glue.Client, triggerName string, timeout time.Duration) (*glue.GetTriggerOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.TriggerStateDeleting),
		Target:  []string{},
		Refresh: statusTrigger(ctx, conn, triggerName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*glue.GetTriggerOutput); ok {
		return output, err
	}

	return nil, err
}
