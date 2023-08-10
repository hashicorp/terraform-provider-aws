// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssmincidents

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssmincidents"
	"github.com/aws/aws-sdk-go-v2/service/ssmincidents/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindResponsePlanByID(context context.Context, client *ssmincidents.Client, arn string) (*ssmincidents.GetResponsePlanOutput, error) {
	input := &ssmincidents.GetResponsePlanInput{
		Arn: aws.String(arn),
	}
	output, err := client.GetResponsePlan(context, input)

	if err != nil {
		var nfe *types.ResourceNotFoundException
		if errors.As(err, &nfe) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func FindReplicationSetByID(context context.Context, client *ssmincidents.Client, arn string) (*types.ReplicationSet, error) {
	input := &ssmincidents.GetReplicationSetInput{
		Arn: aws.String(arn),
	}
	output, err := client.GetReplicationSet(context, input)
	if err != nil {
		var notFoundError *types.ResourceNotFoundException
		if errors.As(err, &notFoundError) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		return nil, err
	}

	if output == nil || output.ReplicationSet == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ReplicationSet, nil
}
