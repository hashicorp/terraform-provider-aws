package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	eventSourceMappingCreateTimeout      = 10 * time.Minute
	eventSourceMappingUpdateTimeout      = 10 * time.Minute
	eventSourceMappingDeleteTimeout      = 5 * time.Minute
	lambdaFunctionCreateTimeout          = 5 * time.Minute
	lambdaFunctionUpdateTimeout          = 5 * time.Minute
	lambdaFunctionPublishTimeout         = 5 * time.Minute
	lambdaFunctionPutConcurrencyTimeout  = 1 * time.Minute
	lambdaFunctionExtraThrottlingTimeout = 9 * time.Minute

	eventSourceMappingPropagationTimeout = 5 * time.Minute
)

func waitEventSourceMappingCreate(conn *lambda.Lambda, id string) (*lambda.EventSourceMappingConfiguration, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			eventSourceMappingStateCreating,
			eventSourceMappingStateDisabling,
			eventSourceMappingStateEnabling,
		},
		Target: []string{
			eventSourceMappingStateDisabled,
			eventSourceMappingStateEnabled,
		},
		Refresh: statusEventSourceMappingState(conn, id),
		Timeout: eventSourceMappingCreateTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*lambda.EventSourceMappingConfiguration); ok {
		return output, err
	}

	return nil, err
}

func waitEventSourceMappingDelete(conn *lambda.Lambda, id string) (*lambda.EventSourceMappingConfiguration, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{eventSourceMappingStateDeleting},
		Target:  []string{},
		Refresh: statusEventSourceMappingState(conn, id),
		Timeout: eventSourceMappingDeleteTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*lambda.EventSourceMappingConfiguration); ok {
		return output, err
	}

	return nil, err
}

func waitEventSourceMappingUpdate(conn *lambda.Lambda, id string) (*lambda.EventSourceMappingConfiguration, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{
			eventSourceMappingStateDisabling,
			eventSourceMappingStateEnabling,
			eventSourceMappingStateUpdating,
		},
		Target: []string{
			eventSourceMappingStateDisabled,
			eventSourceMappingStateEnabled,
		},
		Refresh: statusEventSourceMappingState(conn, id),
		Timeout: eventSourceMappingUpdateTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*lambda.EventSourceMappingConfiguration); ok {
		return output, err
	}

	return nil, err
}
