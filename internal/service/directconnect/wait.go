// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package directconnect

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/directconnect"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directconnect/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
)

const (
	connectionDisassociatedTimeout = 1 * time.Minute
	hostedConnectionDeletedTimeout = 10 * time.Minute
	lagDeletedTimeout              = 10 * time.Minute
)

func waitHostedConnectionDeleted(ctx context.Context, conn *directconnect.Client, id string) (*awstypes.Connection, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ConnectionStatePending, awstypes.ConnectionStateOrdering, awstypes.ConnectionStateAvailable, awstypes.ConnectionStateRequested, awstypes.ConnectionStateDeleting),
		Target:  []string{},
		Refresh: statusHostedConnectionState(ctx, conn, id),
		Timeout: hostedConnectionDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Connection); ok {
		return output, err
	}

	return nil, err
}

func waitLagDeleted(ctx context.Context, conn *directconnect.Client, id string) (*awstypes.Lag, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.LagStateAvailable, awstypes.LagStateRequested, awstypes.LagStatePending, awstypes.LagStateDeleting),
		Target:  []string{},
		Refresh: statusLagState(ctx, conn, id),
		Timeout: lagDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Lag); ok {
		return output, err
	}

	return nil, err
}
