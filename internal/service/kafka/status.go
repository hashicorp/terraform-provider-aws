// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kafka

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func statusClusterState(ctx context.Context, conn *kafka.Kafka, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findClusterV2ByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func statusClusterOperationState(ctx context.Context, conn *kafka.Kafka, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindClusterOperationByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.OperationState), nil
	}
}

func statusConfigurationState(ctx context.Context, conn *kafka.Kafka, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindConfigurationByARN(ctx, conn, arn)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}
