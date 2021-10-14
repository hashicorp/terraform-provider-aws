package waiter

import (
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/synthetics"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	CanaryCreatedTimeout = 5 * time.Minute
	CanaryRunningTimeout = 5 * time.Minute
	CanaryStoppedTimeout = 5 * time.Minute
	CanaryDeletedTimeout = 5 * time.Minute
)

func CanaryReady(conn *synthetics.Synthetics, name string) (*synthetics.Canary, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{synthetics.CanaryStateCreating, synthetics.CanaryStateUpdating},
		Target:  []string{synthetics.CanaryStateReady},
		Refresh: CanaryState(conn, name),
		Timeout: CanaryCreatedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*synthetics.Canary); ok {
		if status := output.Status; aws.StringValue(status.State) == synthetics.CanaryStateError {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(status.StateReasonCode), aws.StringValue(status.StateReason)))
		}

		return output, err
	}

	return nil, err
}

func CanaryStopped(conn *synthetics.Synthetics, name string) (*synthetics.Canary, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			synthetics.CanaryStateStopping,
			synthetics.CanaryStateUpdating,
			synthetics.CanaryStateRunning,
			synthetics.CanaryStateReady,
			synthetics.CanaryStateStarting,
		},
		Target:  []string{synthetics.CanaryStateStopped},
		Refresh: CanaryState(conn, name),
		Timeout: CanaryStoppedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*synthetics.Canary); ok {
		if status := output.Status; aws.StringValue(status.State) == synthetics.CanaryStateError {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(status.StateReasonCode), aws.StringValue(status.StateReason)))
		}

		return output, err
	}

	return nil, err
}

func CanaryRunning(conn *synthetics.Synthetics, name string) (*synthetics.Canary, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			synthetics.CanaryStateStarting,
			synthetics.CanaryStateUpdating,
			synthetics.CanaryStateStarting,
			synthetics.CanaryStateReady,
		},
		Target:  []string{synthetics.CanaryStateRunning},
		Refresh: CanaryState(conn, name),
		Timeout: CanaryRunningTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*synthetics.Canary); ok {
		if status := output.Status; aws.StringValue(status.State) == synthetics.CanaryStateError {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(status.StateReasonCode), aws.StringValue(status.StateReason)))
		}

		return output, err
	}

	return nil, err
}

func CanaryDeleted(conn *synthetics.Synthetics, name string) (*synthetics.Canary, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{synthetics.CanaryStateDeleting, synthetics.CanaryStateStopped},
		Target:  []string{},
		Refresh: CanaryState(conn, name),
		Timeout: CanaryDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*synthetics.Canary); ok {
		if status := output.Status; aws.StringValue(status.State) == synthetics.CanaryStateError {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(status.StateReasonCode), aws.StringValue(status.StateReason)))
		}

		return output, err
	}

	return nil, err
}
