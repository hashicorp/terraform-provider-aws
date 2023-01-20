package transfer

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/transfer"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	serverDeletedTimeout = 10 * time.Minute
	userDeletedTimeout   = 10 * time.Minute
)

func waitServerCreated(ctx context.Context, conn *transfer.Transfer, id string, timeout time.Duration) (*transfer.DescribedServer, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{transfer.StateStarting},
		Target:  []string{transfer.StateOnline},
		Refresh: statusServerState(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*transfer.DescribedServer); ok {
		return output, err
	}

	return nil, err
}

func waitServerDeleted(ctx context.Context, conn *transfer.Transfer, id string) (*transfer.DescribedServer, error) {
	stateConf := &resource.StateChangeConf{
		Pending: transfer.State_Values(),
		Target:  []string{},
		Refresh: statusServerState(ctx, conn, id),
		Timeout: serverDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*transfer.DescribedServer); ok {
		return output, err
	}

	return nil, err
}

func waitServerStarted(ctx context.Context, conn *transfer.Transfer, id string, timeout time.Duration) (*transfer.DescribedServer, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{transfer.StateStarting, transfer.StateOffline, transfer.StateStopping},
		Target:  []string{transfer.StateOnline},
		Refresh: statusServerState(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*transfer.DescribedServer); ok {
		return output, err
	}

	return nil, err
}

func waitServerStopped(ctx context.Context, conn *transfer.Transfer, id string, timeout time.Duration) (*transfer.DescribedServer, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{transfer.StateStarting, transfer.StateOnline, transfer.StateStopping},
		Target:  []string{transfer.StateOffline},
		Refresh: statusServerState(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*transfer.DescribedServer); ok {
		return output, err
	}

	return nil, err
}

func waitUserDeleted(ctx context.Context, conn *transfer.Transfer, serverID, userName string) (*transfer.DescribedUser, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{userStateExists},
		Target:  []string{},
		Refresh: statusUserState(ctx, conn, serverID, userName),
		Timeout: userDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*transfer.DescribedUser); ok {
		return output, err
	}

	return nil, err
}
