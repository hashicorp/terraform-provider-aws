// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kinesisanalytics

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kinesisanalytics"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kinesisanalytics/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func findApplicationDetailByName(ctx context.Context, conn *kinesisanalytics.Client, name string) (*awstypes.ApplicationDetail, error) {
	input := &kinesisanalytics.DescribeApplicationInput{
		ApplicationName: aws.String(name),
	}

	return findApplicationDetail(ctx, conn, input)
}

func findApplicationDetail(ctx context.Context, conn *kinesisanalytics.Client, input *kinesisanalytics.DescribeApplicationInput) (*awstypes.ApplicationDetail, error) {
	output, err := conn.DescribeApplication(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.ApplicationDetail == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.ApplicationDetail, nil
}
