// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package gamelift

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/gamelift"
	awstypes "github.com/aws/aws-sdk-go-v2/service/gamelift/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
)

func waitGameServerGroupActive(ctx context.Context, conn *gamelift.Client, name string, timeout time.Duration) (*awstypes.GameServerGroup, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.GameServerGroupStatusNew,
			awstypes.GameServerGroupStatusActivating,
		),
		Target:  enum.Slice(awstypes.GameServerGroupStatusActive),
		Refresh: statusGameServerGroup(ctx, conn, name),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.GameServerGroup); ok {
		return output, err
	}

	return nil, err
}

func waitGameServerGroupTerminated(ctx context.Context, conn *gamelift.Client, name string, timeout time.Duration) error {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(
			awstypes.GameServerGroupStatusDeleteScheduled,
			awstypes.GameServerGroupStatusDeleting,
		),
		Target:  []string{},
		Refresh: statusGameServerGroup(ctx, conn, name),
		Timeout: timeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	if errs.IsA[*awstypes.NotFoundException](err) {
		return nil
	}

	if err != nil {
		return fmt.Errorf("deleting GameLift Game Server Group (%s): %w", name, err)
	}

	return nil
}
