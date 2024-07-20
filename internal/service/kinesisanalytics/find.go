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
)

// FindApplicationDetailByName returns the application corresponding to the specified name.
// Returns NotFoundError if no application is found.
func FindApplicationDetailByName(ctx context.Context, conn *kinesisanalytics.Client, name string) (*awstypes.ApplicationDetail, error) {
	input := &kinesisanalytics.DescribeApplicationInput{
		ApplicationName: aws.String(name),
	}

	return FindApplicationDetail(ctx, conn, input)
}

// FindApplicationDetail returns the application details corresponding to the specified name.
// Returns NotFoundError if no application is found.
func FindApplicationDetail(ctx context.Context, conn *kinesisanalytics.Client, input *kinesisanalytics.DescribeApplicationInput) (*awstypes.ApplicationDetail, error) {
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
		return nil, &retry.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return output.ApplicationDetail, nil
}
