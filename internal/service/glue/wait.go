package glue

import (
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
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
func waitMLTransformDeleted(conn *glue.Glue, transformId string) (*glue.GetMLTransformOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{glue.TransformStatusTypeNotReady, glue.TransformStatusTypeReady, glue.TransformStatusTypeDeleting},
		Target:  []string{},
		Refresh: statusMLTransform(conn, transformId),
		Timeout: mlTransformDeleteTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*glue.GetMLTransformOutput); ok {
		return output, err
	}

	return nil, err
}

// waitRegistryDeleted waits for a Registry to return Deleted
func waitRegistryDeleted(conn *glue.Glue, registryID string) (*glue.GetRegistryOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{glue.RegistryStatusDeleting},
		Target:  []string{},
		Refresh: statusRegistry(conn, registryID),
		Timeout: registryDeleteTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*glue.GetRegistryOutput); ok {
		return output, err
	}

	return nil, err
}

// waitSchemaAvailable waits for a Schema to return Available
func waitSchemaAvailable(conn *glue.Glue, registryID string) (*glue.GetSchemaOutput, error) { //nolint:unparam
	stateConf := &resource.StateChangeConf{
		Pending: []string{glue.SchemaStatusPending},
		Target:  []string{glue.SchemaStatusAvailable},
		Refresh: statusSchema(conn, registryID),
		Timeout: schemaAvailableTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*glue.GetSchemaOutput); ok {
		return output, err
	}

	return nil, err
}

// waitSchemaDeleted waits for a Schema to return Deleted
func waitSchemaDeleted(conn *glue.Glue, registryID string) (*glue.GetSchemaOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{glue.SchemaStatusDeleting},
		Target:  []string{},
		Refresh: statusSchema(conn, registryID),
		Timeout: schemaDeleteTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*glue.GetSchemaOutput); ok {
		return output, err
	}

	return nil, err
}

// waitSchemaVersionAvailable waits for a Schema to return Available
func waitSchemaVersionAvailable(conn *glue.Glue, registryID string) (*glue.GetSchemaVersionOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{glue.SchemaVersionStatusPending},
		Target:  []string{glue.SchemaVersionStatusAvailable},
		Refresh: statusSchemaVersion(conn, registryID),
		Timeout: schemaVersionAvailableTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*glue.GetSchemaVersionOutput); ok {
		return output, err
	}

	return nil, err
}

// waitTriggerCreated waits for a Trigger to return Created
func waitTriggerCreated(conn *glue.Glue, triggerName string) (*glue.GetTriggerOutput, error) { //nolint:unparam
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			glue.TriggerStateActivating,
			glue.TriggerStateCreating,
			glue.TriggerStateUpdating,
		},
		Target: []string{
			glue.TriggerStateActivated,
			glue.TriggerStateCreated,
		},
		Refresh: statusTrigger(conn, triggerName),
		Timeout: triggerCreateTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*glue.GetTriggerOutput); ok {
		return output, err
	}

	return nil, err
}

// waitTriggerDeleted waits for a Trigger to return Deleted
func waitTriggerDeleted(conn *glue.Glue, triggerName string) (*glue.GetTriggerOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{glue.TriggerStateDeleting},
		Target:  []string{},
		Refresh: statusTrigger(conn, triggerName),
		Timeout: triggerDeleteTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*glue.GetTriggerOutput); ok {
		return output, err
	}

	return nil, err
}

func waitDevEndpointCreated(conn *glue.Glue, name string) (*glue.DevEndpoint, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{devEndpointStatusProvisioning},
		Target:  []string{devEndpointStatusReady},
		Refresh: statusDevEndpoint(conn, name),
		Timeout: 15 * time.Minute,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*glue.DevEndpoint); ok {
		if status := aws.StringValue(output.Status); status == devEndpointStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.FailureReason)))
		}

		return output, err
	}

	return nil, err
}

func waitDevEndpointDeleted(conn *glue.Glue, name string) (*glue.DevEndpoint, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{devEndpointStatusTerminating},
		Target:  []string{},
		Refresh: statusDevEndpoint(conn, name),
		Timeout: 15 * time.Minute,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*glue.DevEndpoint); ok {
		if status := aws.StringValue(output.Status); status == devEndpointStatusFailed {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.FailureReason)))
		}

		return output, err
	}

	return nil, err
}

func waitPartitionIndexCreated(conn *glue.Glue, id string, timeout time.Duration) (*glue.PartitionIndexDescriptor, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{glue.PartitionIndexStatusCreating},
		Target:  []string{glue.PartitionIndexStatusActive},
		Refresh: statusPartitionIndex(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*glue.PartitionIndexDescriptor); ok {
		return output, err
	}

	return nil, err
}

func waitPartitionIndexDeleted(conn *glue.Glue, id string, timeout time.Duration) (*glue.PartitionIndexDescriptor, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{glue.PartitionIndexStatusDeleting},
		Target:  []string{},
		Refresh: statusPartitionIndex(conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*glue.PartitionIndexDescriptor); ok {
		return output, err
	}

	return nil, err
}
