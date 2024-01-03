// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package amp

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/amp"
	"github.com/aws/aws-sdk-go-v2/service/amp/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
)

const (
	// Maximum amount of time to wait for a Workspace to be created, updated, or deleted
	workspaceTimeout = 5 * time.Minute
)

func waitScraperCreated(ctx context.Context, conn *amp.Client, id string, timeout time.Duration) (*types.ScraperDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.ScraperStatusCodeCreating),
		Target:  enum.Slice(types.ScraperStatusCodeActive),
		Refresh: statusScraper(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if out, ok := outputRaw.(*types.ScraperDescription); ok {
		return out, err
	}

	return nil, err
}

func waitScraperDeleted(ctx context.Context, conn *amp.Client, id string, timeout time.Duration) (*types.ScraperDescription, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.ScraperStatusCodeActive, types.ScraperStatusCodeDeleting),
		Target:  []string{},
		Refresh: statusScraper(ctx, conn, id),
		Timeout: timeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.ScraperDescription); ok {
		return output, err
	}

	return nil, err
}
