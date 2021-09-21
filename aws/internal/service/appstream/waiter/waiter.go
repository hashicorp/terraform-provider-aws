package waiter

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/appstream"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	// StackOperationTimeout Maximum amount of time to wait for Stack operation eventual consistency
	StackOperationTimeout = 4 * time.Minute
)

// StackStateDeleted waits for a deleted stack
func StackStateDeleted(ctx context.Context, conn *appstream.AppStream, name string) (*appstream.Stack, error) {
	stateConf := &resource.StateChangeConf{
		Target:  []string{"NotFound", "Unknown"},
		Refresh: StackState(ctx, conn, name),
		Timeout: StackOperationTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*appstream.Stack); ok {
		return output, err
	}

	return nil, err
}
