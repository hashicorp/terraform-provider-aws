// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package transfer

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/transfer"
	awstypes "github.com/aws/aws-sdk-go-v2/service/transfer/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
)

const (
	serverDeletedTimeout = 10 * time.Minute
)

func waitServerCreated(ctx context.Context, conn *transfer.Client, id string, timeout time.Duration) (*awstypes.DescribedServer, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.StateStarting),
		Target:  enum.Slice(awstypes.StateOnline),
		Refresh: statusServerState(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DescribedServer); ok {
		return output, err
	}

	return nil, err
}

func waitServerDeleted(ctx context.Context, conn *transfer.Client, id string) (*awstypes.DescribedServer, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.StateOffline, awstypes.StateOnline, awstypes.StateStarting, awstypes.StateStopping, awstypes.StateStartFailed, awstypes.StateStopFailed),
		Target:  []string{},
		Refresh: statusServerState(ctx, conn, id),
		Timeout: serverDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DescribedServer); ok {
		return output, err
	}

	return nil, err
}

func waitServerStarted(ctx context.Context, conn *transfer.Client, id string, timeout time.Duration) (*awstypes.DescribedServer, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.StateStarting, awstypes.StateOffline, awstypes.StateStopping),
		Target:  enum.Slice(awstypes.StateOnline),
		Refresh: statusServerState(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DescribedServer); ok {
		return output, err
	}

	return nil, err
}

func waitServerStopped(ctx context.Context, conn *transfer.Client, id string, timeout time.Duration) (*awstypes.DescribedServer, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.StateStarting, awstypes.StateOnline, awstypes.StateStopping),
		Target:  enum.Slice(awstypes.StateOffline),
		Refresh: statusServerState(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.DescribedServer); ok {
		return output, err
	}

	return nil, err
}
