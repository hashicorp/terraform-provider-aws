// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package apigateway

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/apigateway"
	awstypes "github.com/aws/aws-sdk-go-v2/service/apigateway/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
)

const (
	// Maximum amount of time for Stage Cache to be available
	stageCacheAvailableTimeout = 90 * time.Minute

	// Maximum amount of time for Stage Cache to update
	stageCacheUpdateTimeout = 30 * time.Minute
)

func waitStageCacheAvailable(ctx context.Context, conn *apigateway.Client, restApiId, name string) (*awstypes.Stage, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.CacheClusterStatusCreateInProgress, awstypes.CacheClusterStatusDeleteInProgress, awstypes.CacheClusterStatusFlushInProgress),
		Target:  enum.Slice(awstypes.CacheClusterStatusAvailable),
		Refresh: stageCacheStatus(ctx, conn, restApiId, name),
		Timeout: stageCacheAvailableTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Stage); ok {
		return output, err
	}

	return nil, err
}

func waitStageCacheUpdated(ctx context.Context, conn *apigateway.Client, restApiId, name string) (*awstypes.Stage, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.CacheClusterStatusCreateInProgress, awstypes.CacheClusterStatusFlushInProgress),
		Target: enum.Slice(awstypes.CacheClusterStatusAvailable,
			// There's an AWS API bug (raised & confirmed in Sep 2016 by support)
			// which causes the stage to remain in deletion state forever
			// TODO: Check if this bug still exists in AWS SDK v2
			awstypes.CacheClusterStatusDeleteInProgress),
		Refresh: stageCacheStatus(ctx, conn, restApiId, name),
		Timeout: stageCacheUpdateTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Stage); ok {
		return output, err
	}

	return nil, err
}
