package kafkaconnect

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafkaconnect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func waitCustomPluginCreated(conn *kafkaconnect.KafkaConnect, arn string, timeout time.Duration) (*kafkaconnect.DescribeCustomPluginOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{kafkaconnect.CustomPluginStateCreating},
		Target:  []string{kafkaconnect.CustomPluginStateActive},
		Refresh: statusCustomPluginState(conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*kafkaconnect.DescribeCustomPluginOutput); ok {
		return output, err
	}

	return nil, err
}

func waitConnectorCreatedWithContext(ctx context.Context, conn *kafkaconnect.KafkaConnect, arn string, timeout time.Duration) (*kafkaconnect.DescribeConnectorOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{kafkaconnect.ConnectorStateCreating},
		Target:  []string{kafkaconnect.ConnectorStateRunning},
		Refresh: statusConnectorState(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*kafkaconnect.DescribeConnectorOutput); ok {
		return output, err
	}

	return nil, err
}

func waitConnectorDeletedWithContext(ctx context.Context, conn *kafkaconnect.KafkaConnect, arn string, timeout time.Duration) (*kafkaconnect.DescribeConnectorOutput, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{kafkaconnect.ConnectorStateDeleting},
		Target:  []string{},
		Refresh: statusConnectorState(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*kafkaconnect.DescribeConnectorOutput); ok {
		return output, err
	}

	return nil, err
}

func waitConnectorOperationCompletedWithContext(ctx context.Context, conn *kafkaconnect.KafkaConnect, arn string, timeout time.Duration) (*kafkaconnect.ConnectorSummary, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{kafkaconnect.ConnectorStateUpdating},
		Target:  []string{kafkaconnect.ConnectorStateRunning},
		Refresh: statusConnectorState(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*kafkaconnect.ConnectorSummary); ok {
		if state := aws.StringValue(output.ConnectorState); state == kafkaconnect.ConnectorStateFailed {
			tfresource.SetLastError(err, fmt.Errorf("connector (%s) state update failed", arn))
		}

		return output, err
	}

	return nil, err
}
