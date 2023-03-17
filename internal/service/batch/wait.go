package batch

import (
	"context"
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func waitComputeEnvironmentCreated(ctx context.Context, conn *batch.Batch, name string, timeout time.Duration) (*batch.ComputeEnvironmentDetail, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{batch.CEStatusCreating},
		Target:  []string{batch.CEStatusValid},
		Refresh: statusComputeEnvironment(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*batch.ComputeEnvironmentDetail); ok {
		if status := aws.StringValue(output.Status); status == batch.CEStatusInvalid {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusReason)))
		}

		return output, err
	}

	return nil, err
}

func waitComputeEnvironmentDeleted(ctx context.Context, conn *batch.Batch, name string, timeout time.Duration) (*batch.ComputeEnvironmentDetail, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{batch.CEStatusDeleting},
		Target:  []string{},
		Refresh: statusComputeEnvironment(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*batch.ComputeEnvironmentDetail); ok {
		if status := aws.StringValue(output.Status); status == batch.CEStatusInvalid {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusReason)))
		}

		return output, err
	}

	return nil, err
}

func waitComputeEnvironmentDisabled(ctx context.Context, conn *batch.Batch, name string, timeout time.Duration) (*batch.ComputeEnvironmentDetail, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{batch.CEStatusUpdating},
		Target:  []string{batch.CEStatusValid},
		Refresh: statusComputeEnvironment(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*batch.ComputeEnvironmentDetail); ok {
		if status := aws.StringValue(output.Status); status == batch.CEStatusInvalid {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusReason)))
		}

		return output, err
	}

	return nil, err
}

func waitComputeEnvironmentUpdated(ctx context.Context, conn *batch.Batch, name string, timeout time.Duration) (*batch.ComputeEnvironmentDetail, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{batch.CEStatusUpdating},
		Target:  []string{batch.CEStatusValid},
		Refresh: statusComputeEnvironment(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*batch.ComputeEnvironmentDetail); ok {
		return v, err
	}

	return nil, err
}
