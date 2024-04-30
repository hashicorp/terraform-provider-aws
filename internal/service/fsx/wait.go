// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fsx

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

func waitSnapshotCreated(ctx context.Context, conn *fsx.FSx, id string, timeout time.Duration) (*fsx.Snapshot, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{fsx.SnapshotLifecycleCreating, fsx.SnapshotLifecyclePending},
		Target:  []string{fsx.SnapshotLifecycleAvailable},
		Refresh: statusSnapshot(ctx, conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*fsx.Snapshot); ok {
		return output, err
	}

	return nil, err
}

func waitSnapshotUpdated(ctx context.Context, conn *fsx.FSx, id string, timeout time.Duration) (*fsx.Snapshot, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{fsx.SnapshotLifecyclePending},
		Target:  []string{fsx.SnapshotLifecycleAvailable},
		Refresh: statusSnapshot(ctx, conn, id),
		Timeout: timeout,
		Delay:   150 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*fsx.Snapshot); ok {
		return output, err
	}

	return nil, err
}

func waitSnapshotDeleted(ctx context.Context, conn *fsx.FSx, id string, timeout time.Duration) (*fsx.Snapshot, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{fsx.SnapshotLifecyclePending, fsx.SnapshotLifecycleDeleting},
		Target:  []string{},
		Refresh: statusSnapshot(ctx, conn, id),
		Timeout: timeout,
		Delay:   30 * time.Second,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*fsx.Snapshot); ok {
		return output, err
	}

	return nil, err
}
