// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package applicationinsights

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/applicationinsights"
	awstypes "github.com/aws/aws-sdk-go-v2/service/applicationinsights/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindApplicationByName(ctx context.Context, conn *applicationinsights.Client, name string) (*awstypes.ApplicationInfo, error) {
	input := applicationinsights.DescribeApplicationInput{
		ResourceGroupName: aws.String(name),
	}

	output, err := conn.DescribeApplication(ctx, &input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ApplicationInfo == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ApplicationInfo, nil
}
