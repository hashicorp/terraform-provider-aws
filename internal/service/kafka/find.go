// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kafka

import (
	"context"

	"github.com/aws/aws-sdk-go/service/kafka"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindServerlessClusterByARN(ctx context.Context, conn *kafka.Kafka, arn string) (*kafka.Cluster, error) {
	output, err := findClusterV2ByARN(ctx, conn, arn)

	if err != nil {
		return nil, err
	}

	if output.Serverless == nil {
		return nil, tfresource.NewEmptyResultError(arn)
	}

	return output, nil
}
