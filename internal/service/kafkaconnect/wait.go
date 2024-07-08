// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kafkaconnect

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/kafkaconnect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func waitWorkerConfigurationDeleted(ctx context.Context, conn *kafkaconnect.KafkaConnect, arn string, timeout time.Duration) (*kafkaconnect.DescribeWorkerConfigurationOutput, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{kafkaconnect.WorkerConfigurationStateDeleting},
		Target:  []string{},
		Refresh: statusWorkerConfigurationState(ctx, conn, arn),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*kafkaconnect.DescribeWorkerConfigurationOutput); ok {
		return output, err
	}

	return nil, err
}
