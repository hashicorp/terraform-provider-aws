package waiter

import (
	"errors"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/batch"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func ComputeEnvironmentCreated(conn *batch.Batch, name string, timeout time.Duration) (*batch.ComputeEnvironmentDetail, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{batch.CEStatusCreating},
		Target:  []string{batch.CEStatusValid},
		Refresh: ComputeEnvironmentStatus(conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*batch.ComputeEnvironmentDetail); ok {
		if status := aws.StringValue(output.Status); status == batch.CEStatusInvalid {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusReason)))
		}

		return output, err
	}

	return nil, err
}

func ComputeEnvironmentDeleted(conn *batch.Batch, name string, timeout time.Duration) (*batch.ComputeEnvironmentDetail, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{batch.CEStatusDeleting},
		Target:  []string{},
		Refresh: ComputeEnvironmentStatus(conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*batch.ComputeEnvironmentDetail); ok {
		if status := aws.StringValue(output.Status); status == batch.CEStatusInvalid {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusReason)))
		}

		return output, err
	}

	return nil, err
}

func ComputeEnvironmentDisabled(conn *batch.Batch, name string, timeout time.Duration) (*batch.ComputeEnvironmentDetail, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{batch.CEStatusUpdating},
		Target:  []string{batch.CEStatusValid},
		Refresh: ComputeEnvironmentStatus(conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*batch.ComputeEnvironmentDetail); ok {
		if status := aws.StringValue(output.Status); status == batch.CEStatusInvalid {
			tfresource.SetLastError(err, errors.New(aws.StringValue(output.StatusReason)))
		}

		return output, err
	}

	return nil, err
}

func ComputeEnvironmentUpdated(conn *batch.Batch, name string, timeout time.Duration) (*batch.ComputeEnvironmentDetail, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{batch.CEStatusUpdating},
		Target:  []string{batch.CEStatusValid},
		Refresh: ComputeEnvironmentStatus(conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if v, ok := outputRaw.(*batch.ComputeEnvironmentDetail); ok {
		return v, err
	}

	return nil, err
}
