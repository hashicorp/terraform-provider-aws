// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kafkaconnect

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafkaconnect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func statusConnectorState(ctx context.Context, conn *kafkaconnect.KafkaConnect, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindConnectorByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.ConnectorState), nil
	}
}

func statusCustomPluginState(ctx context.Context, conn *kafkaconnect.KafkaConnect, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindCustomPluginByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.CustomPluginState), nil
	}
}
