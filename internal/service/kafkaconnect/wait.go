// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kafkaconnect

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafkaconnect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func waitConnectorCreated(ctx context.Context, conn *kafkaconnect.KafkaConnect, arn string, timeout time.Duration) (*kafkaconnect.DescribeConnectorOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{kafkaconnect.ConnectorStateCreating},
		Target:  []string{kafkaconnect.ConnectorStateRunning},
		Refresh: statusConnectorState(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*kafkaconnect.DescribeConnectorOutput); ok {
		if state, stateDescription := aws.StringValue(output.ConnectorState), output.StateDescription; state == kafkaconnect.ConnectorStateFailed && stateDescription != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(stateDescription.Code), aws.StringValue(stateDescription.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitConnectorDeleted(ctx context.Context, conn *kafkaconnect.KafkaConnect, arn string, timeout time.Duration) (*kafkaconnect.DescribeConnectorOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{kafkaconnect.ConnectorStateDeleting},
		Target:  []string{},
		Refresh: statusConnectorState(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*kafkaconnect.DescribeConnectorOutput); ok {
		if state, stateDescription := aws.StringValue(output.ConnectorState), output.StateDescription; state == kafkaconnect.ConnectorStateFailed && stateDescription != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(stateDescription.Code), aws.StringValue(stateDescription.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitConnectorUpdated(ctx context.Context, conn *kafkaconnect.KafkaConnect, arn string, timeout time.Duration) (*kafkaconnect.DescribeConnectorOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{kafkaconnect.ConnectorStateUpdating},
		Target:  []string{kafkaconnect.ConnectorStateRunning},
		Refresh: statusConnectorState(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*kafkaconnect.DescribeConnectorOutput); ok {
		if state, stateDescription := aws.StringValue(output.ConnectorState), output.StateDescription; state == kafkaconnect.ConnectorStateFailed && stateDescription != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(stateDescription.Code), aws.StringValue(stateDescription.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitCustomPluginCreated(ctx context.Context, conn *kafkaconnect.KafkaConnect, arn string, timeout time.Duration) (*kafkaconnect.DescribeCustomPluginOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{kafkaconnect.CustomPluginStateCreating},
		Target:  []string{kafkaconnect.CustomPluginStateActive},
		Refresh: statusCustomPluginState(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*kafkaconnect.DescribeCustomPluginOutput); ok {
		if state, stateDescription := aws.StringValue(output.CustomPluginState), output.StateDescription; state == kafkaconnect.CustomPluginStateCreateFailed && stateDescription != nil {
			tfresource.SetLastError(err, fmt.Errorf("%s: %s", aws.StringValue(stateDescription.Code), aws.StringValue(stateDescription.Message)))
		}

		return output, err
	}

	return nil, err
}

func waitCustomPluginDeleted(ctx context.Context, conn *kafkaconnect.KafkaConnect, arn string, timeout time.Duration) (*kafkaconnect.DescribeCustomPluginOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{kafkaconnect.CustomPluginStateDeleting},
		Target:  []string{},
		Refresh: statusCustomPluginState(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*kafkaconnect.DescribeCustomPluginOutput); ok {
		return output, err
	}

	return nil, err
}
