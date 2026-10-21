// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ssmincidents

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssmincidents"
	"github.com/aws/aws-sdk-go-v2/service/ssmincidents/types"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindResponsePlanByID(context context.Context, client *ssmincidents.Client, arn string) (*ssmincidents.GetResponsePlanOutput, error) {
	input := &ssmincidents.GetResponsePlanInput{
		Arn: aws.String(arn),
	}
	output, err := client.GetResponsePlan(context, input)

	if err != nil {
		if errs.IsA[*types.ResourceNotFoundException](err) {
			return nil, &retry.NotFoundError{
				LastError: err,
			}
		}

		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError()
	}

	return output, nil
}
