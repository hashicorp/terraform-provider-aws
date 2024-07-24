// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package kinesisanalyticsv2

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/kinesisanalyticsv2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/kinesisanalyticsv2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
)

// waitSnapshotCreated waits for a Snapshot to return Created
func waitSnapshotCreated(ctx context.Context, conn *kinesisanalyticsv2.Client, applicationName, snapshotName string, timeout time.Duration) (*awstypes.SnapshotDetails, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.SnapshotStatusCreating),
		Target:  enum.Slice(awstypes.SnapshotStatusReady),
		Refresh: statusSnapshotDetails(ctx, conn, applicationName, snapshotName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*awstypes.SnapshotDetails); ok {
		return v, err
	}

	return nil, err
}

// waitSnapshotDeleted waits for a Snapshot to return Deleted
func waitSnapshotDeleted(ctx context.Context, conn *kinesisanalyticsv2.Client, applicationName, snapshotName string, timeout time.Duration) (*awstypes.SnapshotDetails, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.SnapshotStatusDeleting),
		Target:  []string{},
		Refresh: statusSnapshotDetails(ctx, conn, applicationName, snapshotName),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*awstypes.SnapshotDetails); ok {
		return v, err
	}

	return nil, err
}
