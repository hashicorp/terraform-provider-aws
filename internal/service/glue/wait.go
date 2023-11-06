// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	// Maximum amount of time to wait for an Operation to return Deleted
	mlTransformDeleteTimeout      = 2 * time.Minute
	registryDeleteTimeout         = 2 * time.Minute
	schemaAvailableTimeout        = 2 * time.Minute
	schemaDeleteTimeout           = 2 * time.Minute
	schemaVersionAvailableTimeout = 2 * time.Minute
	triggerCreateTimeout          = 5 * time.Minute
	triggerDeleteTimeout          = 5 * time.Minute
)

// waitMLTransformDeleted waits for an MLTransform to return Deleted
func waitMLTransformDeleted(ctx context.Context, conn *glue.Glue, transformId string) (*glue.GetMLTransformOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{glue.TransformStatusTypeNotReady, glue.TransformStatusTypeReady, glue.TransformStatusTypeDeleting},
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
func waitRegistryDeleted(ctx context.Context, conn *glue.Glue, registryID string) (*glue.GetRegistryOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{glue.RegistryStatusDeleting},
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
func waitSchemaAvailable(ctx context.Context, conn *glue.Glue, registryID string) (*glue.GetSchemaOutput, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: []string{glue.SchemaStatusPending},
		Target:  []string{glue.SchemaStatusAvailable},
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
func waitSchemaDeleted(ctx context.Context, conn *glue.Glue, registryID string) (*glue.GetSchemaOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{glue.SchemaStatusDeleting},
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
func waitSchemaVersionAvailable(ctx context.Context, conn *glue.Glue, registryID string) (*glue.GetSchemaVersionOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{glue.SchemaVersionStatusPending},
		Target:  []string{glue.SchemaVersionStatusAvailable},
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
func waitTriggerCreated(ctx context.Context, conn *glue.Glue, triggerName string) (*glue.GetTriggerOutput, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: []string{
			glue.TriggerStateActivating,
			glue.TriggerStateCreating,
			glue.TriggerStateUpdating,
		},
		Target: []string{
			glue.TriggerStateActivated,
			glue.TriggerStateCreated,
		},
		Refresh: statusTrigger(ctx, conn, triggerName),
		Timeout: triggerCreateTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*glue.GetTriggerOutput); ok {
		return output, err
	}

	return nil, err
}

// waitTriggerDeleted waits for a Trigger to return Deleted
func waitTriggerDeleted(ctx context.Context, conn *glue.Glue, triggerName string) (*glue.GetTriggerOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{glue.TriggerStateDeleting},
		Target:  []string{},
		Refresh: statusTrigger(ctx, conn, triggerName),
		Timeout: triggerDeleteTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*glue.GetTriggerOutput); ok {
		return output, err
	}

	return nil, err
}

func waitDevEndpointCreated(ctx context.Context, conn *glue.Glue, name string) (*glue.DevEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{devEndpointStatusProvisioning},
		Target:  []string{devEndpointStatusReady},
		Refresh: statusDevEndpoint(ctx, conn, name),
		Timeout: 15 * time.Minute,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*glue.DevEndpoint); ok {
		if status := aws.StringValue(output.Status); status == devEndpointStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.FailureReason)))
		}

		return output, err
	}

	return nil, err
}

func waitDevEndpointDeleted(ctx context.Context, conn *glue.Glue, name string) (*glue.DevEndpoint, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{devEndpointStatusTerminating},
		Target:  []string{},
		Refresh: statusDevEndpoint(ctx, conn, name),
		Timeout: 15 * time.Minute,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*glue.DevEndpoint); ok {
		if status := aws.StringValue(output.Status); status == devEndpointStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.FailureReason)))
		}

		return output, err
	}

	return nil, err
}

func waitPartitionIndexCreated(ctx context.Context, conn *glue.Glue, id string, timeout time.Duration) (*glue.PartitionIndexDescriptor, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{glue.PartitionIndexStatusCreating},
		Target:  []string{glue.PartitionIndexStatusActive},
		Refresh: statusPartitionIndex(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*glue.PartitionIndexDescriptor); ok {
		return output, err
	}

	return nil, err
}

func waitPartitionIndexDeleted(ctx context.Context, conn *glue.Glue, id string, timeout time.Duration) (*glue.PartitionIndexDescriptor, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{glue.PartitionIndexStatusDeleting},
		Target:  []string{},
		Refresh: statusPartitionIndex(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*glue.PartitionIndexDescriptor); ok {
		return output, err
	}

	return nil, err
}
