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
