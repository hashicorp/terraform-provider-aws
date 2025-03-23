// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appstream

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/appstream"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appstream/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

const (
	// userOperationTimeout Maximum amount of time to wait for User operation eventual consistency
	userOperationTimeout = 4 * time.Minute
	// iamPropagationTimeout Maximum amount of time to wait for an iam resource eventual consistency
	iamPropagationTimeout = 2 * time.Minute
	userAvailable         = "AVAILABLE"
)

// waitUserAvailable waits for a user be available
func waitUserAvailable(ctx context.Context, conn *appstream.Client, username, authType string) (*awstypes.User, error) {
	stateConf := &retry.StateChangeConf{
		Target:  []string{userAvailable},
		Refresh: statusUserAvailable(ctx, conn, username, authType),
		Timeout: userOperationTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.User); ok {
		return output, err
	}

	return nil, err
}
