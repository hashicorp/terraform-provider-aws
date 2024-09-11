// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/glue"
	awstypes "github.com/aws/aws-sdk-go-v2/service/glue/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	// Maximum amount of time to wait for an Operation to return Deleted
	mlTransformDeleteTimeout      = 2 * time.Minute
	iamPropagationTimeout         = 2 * time.Minute
	registryDeleteTimeout         = 2 * time.Minute
	schemaAvailableTimeout        = 2 * time.Minute
	schemaDeleteTimeout           = 2 * time.Minute
	schemaVersionAvailableTimeout = 2 * time.Minute
)

// waitMLTransformDeleted waits for an MLTransform to return Deleted
func waitMLTransformDeleted(ctx context.Context, conn *glue.Client, transformId string) (*glue.GetMLTransformOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.TransformStatusTypeNotReady, awstypes.TransformStatusTypeReady, awstypes.TransformStatusTypeDeleting),
		Target:  []string{},
		Refresh: statusMLTransform(ctx, conn, transformId),
		Timeout: mlTransformDeleteTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*glue.GetMLTransformOutput); ok {
		return output, err
	}

	return nil, err
}

// waitRegistryDeleted waits for a Registry to return Deleted
func waitRegistryDeleted(ctx context.Context, conn *glue.Client, registryID string) (*glue.GetRegistryOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.RegistryStatusDeleting),
		Target:  []string{},
		Refresh: statusRegistry(ctx, conn, registryID),
		Timeout: registryDeleteTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*glue.GetRegistryOutput); ok {
		return output, err
	}

	return nil, err
}

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

func waitDevEndpointCreated(ctx context.Context, conn *glue.Client, name string) (*awstypes.DevEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{devEndpointStatusProvisioning},
		Target:  []string{devEndpointStatusReady},
		Refresh: statusDevEndpoint(ctx, conn, name),
		Timeout: 15 * time.Minute,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DevEndpoint); ok {
		if status := aws.ToString(output.Status); status == devEndpointStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.FailureReason)))
		}

		return output, err
	}

	return nil, err
}

func waitDevEndpointDeleted(ctx context.Context, conn *glue.Client, name string) (*awstypes.DevEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{devEndpointStatusTerminating},
		Target:  []string{},
		Refresh: statusDevEndpoint(ctx, conn, name),
		Timeout: 15 * time.Minute,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DevEndpoint); ok {
		if status := aws.ToString(output.Status); status == devEndpointStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.ToString(output.FailureReason)))
		}

		return output, err
	}

	return nil, err
}

func waitPartitionIndexCreated(ctx context.Context, conn *glue.Client, id string, timeout time.Duration) (*awstypes.PartitionIndexDescriptor, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.PartitionIndexStatusCreating),
		Target:  enum.Slice(awstypes.PartitionIndexStatusActive),
		Refresh: statusPartitionIndex(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.PartitionIndexDescriptor); ok {
		return output, err
	}

	return nil, err
}

func waitPartitionIndexDeleted(ctx context.Context, conn *glue.Client, id string, timeout time.Duration) (*awstypes.PartitionIndexDescriptor, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.PartitionIndexStatusDeleting),
		Target:  []string{},
		Refresh: statusPartitionIndex(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.PartitionIndexDescriptor); ok {
		return output, err
	}

	return nil, err
}
