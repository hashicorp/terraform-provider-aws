// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindStandardsControlByStandardsSubscriptionARNAndStandardsControlARN(ctx context.Context, conn *securityhub.Client, standardsSubscriptionARN, standardsControlARN string) (*types.StandardsControl, error) {
	input := &securityhub.DescribeStandardsControlsInput{
		StandardsSubscriptionArn: aws.String(standardsSubscriptionARN),
	}

	output, err := conn.DescribeStandardsControls(ctx, input)

	var result types.StandardsControl
	for _, control := range output.Controls {
		if aws.ToString(control.StandardsControlArn) == standardsControlARN {
			result = control
			break
		}
	}

	if errs.IsA[*types.ResourceNotFoundException](err) {
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

	return &result, nil
}
