// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kafkaconnect

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kafkaconnect"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindWorkerConfigurationByARN(ctx context.Context, conn *kafkaconnect.KafkaConnect, arn string) (*kafkaconnect.DescribeWorkerConfigurationOutput, error) {
	input := &kafkaconnect.DescribeWorkerConfigurationInput{
		WorkerConfigurationArn: aws.String(arn),
	}

	output, err := conn.DescribeWorkerConfigurationWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, kafkaconnect.ErrCodeNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}
