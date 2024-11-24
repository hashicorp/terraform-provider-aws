// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package memorydb

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/memorydb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

const (
	userActiveTimeout  = 5 * time.Minute
	userDeletedTimeout = 5 * time.Minute

	snapshotAvailableTimeout = 120 * time.Minute
	snapshotDeletedTimeout   = 120 * time.Minute
)

// waitUserActive waits for MemoryDB user to reach an active state after modifications.
func waitUserActive(ctx context.Context, conn *memorydb.Client, userId string) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{userStatusModifying},
		Target:  []string{userStatusActive},
		Refresh: statusUser(ctx, conn, userId),
		Timeout: userActiveTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

// waitUserDeleted waits for MemoryDB user to be deleted.
func waitUserDeleted(ctx context.Context, conn *memorydb.Client, userId string) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{userStatusDeleting},
		Target:  []string{},
		Refresh: statusUser(ctx, conn, userId),
		Timeout: userDeletedTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

// waitSnapshotAvailable waits for MemoryDB snapshot to reach the available state.
func waitSnapshotAvailable(ctx context.Context, conn *memorydb.Client, snapshotId string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{snapshotStatusCreating},
		Target:  []string{snapshotStatusAvailable},
		Refresh: statusSnapshot(ctx, conn, snapshotId),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

// waitSnapshotDeleted waits for MemoryDB snapshot to be deleted.
func waitSnapshotDeleted(ctx context.Context, conn *memorydb.Client, snapshotId string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{snapshotStatusDeleting},
		Target:  []string{},
		Refresh: statusSnapshot(ctx, conn, snapshotId),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}
