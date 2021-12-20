package memorydb

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/memorydb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	userActiveTimeout  = 5 * time.Minute
	userDeletedTimeout = 5 * time.Minute
)

// waitUserActive waits for MemoryDB user to reach an active state after modifications.
func waitUserActive(ctx context.Context, conn *memorydb.MemoryDB, userId string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{UserStatusModifying},
		Target:  []string{UserStatusActive},
		Refresh: statusUser(ctx, conn, userId),
		Timeout: userActiveTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

// waitUserDeleted waits for MemoryDB user to reach an active state after modifications.
func waitUserDeleted(ctx context.Context, conn *memorydb.MemoryDB, userId string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{UserStatusDeleting},
		Target:  []string{},
		Refresh: statusUser(ctx, conn, userId),
		Timeout: userDeletedTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}
