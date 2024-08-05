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
	lagDeletedTimeout = 10 * time.Minute
)

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
