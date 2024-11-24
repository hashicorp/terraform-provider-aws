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
