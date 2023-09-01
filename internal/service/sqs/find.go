// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package sqs

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindQueueAttributesByURL(ctx context.Context, conn *sqs.SQS, url string) (map[string]string, error) {
	input := &sqs.GetQueueAttributesInput{
		AttributeNames: aws.StringSlice([]string{sqs.QueueAttributeNameAll}),
		QueueUrl:       aws.String(url),
	}

	output, err := conn.GetQueueAttributesWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, sqs.ErrCodeQueueDoesNotExist) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.Attributes == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return aws.StringValueMap(output.Attributes), nil
}

func FindQueueAttributeByURL(ctx context.Context, conn *sqs.SQS, url string, attributeName string) (string, error) {
	input := &sqs.GetQueueAttributesInput{
		AttributeNames: aws.StringSlice([]string{attributeName}),
		QueueUrl:       aws.String(url),
	}

	output, err := conn.GetQueueAttributesWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, sqs.ErrCodeQueueDoesNotExist) {
		return "", &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return "", err
	}

	if output == nil || output.Attributes == nil {
		return "", tfresource.NewEmptyResultError(input)
	}

	v, ok := output.Attributes[attributeName]

	if !ok || aws.StringValue(v) == "" {
		return "", tfresource.NewEmptyResultError(input)
	}

	return aws.StringValue(v), nil
}
