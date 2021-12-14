package memorydb

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/memorydb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	UserActiveTimeout  = 5 * time.Minute
	UserDeletedTimeout = 5 * time.Minute
)

// WaitUserActive waits for MemoryDB user to reach an active state after modifications.
func WaitUserActive(ctx context.Context, conn *memorydb.MemoryDB, userId string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{UserStatusModifying},
		Target:  []string{UserStatusActive},
		Refresh: StatusUser(ctx, conn, userId),
		Timeout: UserActiveTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

// WaitUserDeleted waits for MemoryDB user to reach an active state after modifications.
func WaitUserDeleted(ctx context.Context, conn *memorydb.MemoryDB, userId string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{UserStatusDeleting},
		Target:  []string{},
		Refresh: StatusUser(ctx, conn, userId),
		Timeout: UserDeletedTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}
