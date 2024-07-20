// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kinesisanalyticsv2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/kinesisanalyticsv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kinesisanalyticsv2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
)

// FindApplicationDetailByName returns the application corresponding to the specified name.
// Returns NotFoundError if no application is found.
func FindApplicationDetailByName(ctx context.Context, conn *kinesisanalyticsv2.Client, name string) (*awstypes.ApplicationDetail, error) {
	input := &kinesisanalyticsv2.DescribeApplicationInput{
		ApplicationName: aws.String(name),
	}

	return FindApplicationDetail(ctx, conn, input)
}

// FindApplicationDetail returns the application details corresponding to the specified input.
// Returns NotFoundError if no application is found.
func FindApplicationDetail(ctx context.Context, conn *kinesisanalyticsv2.Client, input *kinesisanalyticsv2.DescribeApplicationInput) (*awstypes.ApplicationDetail, error) {
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

// FindSnapshotDetailsByApplicationAndSnapshotNames returns the application snapshot details corresponding to the specified application and snapshot names.
// Returns NotFoundError if no application snapshot is found.
func FindSnapshotDetailsByApplicationAndSnapshotNames(ctx context.Context, conn *kinesisanalyticsv2.Client, applicationName, snapshotName string) (*awstypes.SnapshotDetails, error) {
	input := &kinesisanalyticsv2.DescribeApplicationSnapshotInput{
		ApplicationName: aws.String(applicationName),
		SnapshotName:    aws.String(snapshotName),
	}

	return FindSnapshotDetails(ctx, conn, input)
}

// FindSnapshotDetails returns the application snapshot details corresponding to the specified input.
// Returns NotFoundError if no application snapshot is found.
func FindSnapshotDetails(ctx context.Context, conn *kinesisanalyticsv2.Client, input *kinesisanalyticsv2.DescribeApplicationSnapshotInput) (*awstypes.SnapshotDetails, error) {
	output, err := conn.DescribeApplicationSnapshot(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if errs.IsAErrorMessageContains[*awstypes.InvalidArgumentException](err, "does not exist") {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.SnapshotDetails == nil {
		return nil, &retry.NotFoundError{
			Message:     "Empty result",
			LastRequest: input,
		}
	}

	return output.SnapshotDetails, nil
}
