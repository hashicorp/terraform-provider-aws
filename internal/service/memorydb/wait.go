package memorydb

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/memorydb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	aclActiveTimeout   = 5 * time.Minute
	aclDeletedTimeout  = 5 * time.Minute
	userActiveTimeout  = 5 * time.Minute
	userDeletedTimeout = 5 * time.Minute
)

// waitACLActive waits for MemoryDB ACL to reach an active state after modifications.
func waitACLActive(ctx context.Context, conn *memorydb.MemoryDB, aclId string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{aclStatusCreating, aclStatusModifying},
		Target:  []string{aclStatusActive},
		Refresh: statusACL(ctx, conn, aclId),
		Timeout: aclActiveTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

// waitACLDeleted waits for MemoryDB ACL to be deleted.
func waitACLDeleted(ctx context.Context, conn *memorydb.MemoryDB, aclId string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{aclStatusDeleting},
		Target:  []string{},
		Refresh: statusACL(ctx, conn, aclId),
		Timeout: aclDeletedTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

// waitUserActive waits for MemoryDB user to reach an active state after modifications.
func waitUserActive(ctx context.Context, conn *memorydb.MemoryDB, userId string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{userStatusModifying},
		Target:  []string{userStatusActive},
		Refresh: statusUser(ctx, conn, userId),
		Timeout: userActiveTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

// waitUserDeleted waits for MemoryDB user to be deleted.
func waitUserDeleted(ctx context.Context, conn *memorydb.MemoryDB, userId string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{userStatusDeleting},
		Target:  []string{},
		Refresh: statusUser(ctx, conn, userId),
		Timeout: userDeletedTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}
