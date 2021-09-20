package appstream

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/appstream"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	// stackOperationTimeout Maximum amount of time to wait for Stack operation eventual consistency
	stackOperationTimeout = 4 * time.Minute
)

// waitStackStateDeleted waits for a deleted stack
func waitStackStateDeleted(ctx context.Context, conn *appstream.AppStream, name string) (*appstream.Stack, error) {
	stateConf := &resource.StateChangeConf{
		Target:  []string{"NotFound", "Unknown"},
		Refresh: statusStackState(ctx, conn, name),
		Timeout: stackOperationTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*appstream.Stack); ok {
		return output, err
	}

	return nil, err
}
