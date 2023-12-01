// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/aws/aws-sdk-go-v2/service/securityhub/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
)

const (
	// Maximum amount of time to wait for an AdminAccount to return Enabled
	adminAccountEnabledTimeout = 5 * time.Minute

	// Maximum amount of time to wait for an AdminAccount to return NotFound
	adminAccountNotFoundTimeout = 5 * time.Minute
)

// waitAdminAccountEnabled waits for an AdminAccount to return Enabled
func waitAdminAccountEnabled(ctx context.Context, conn *securityhub.Client, adminAccountID string) (*types.AdminAccount, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{adminStatusNotFound},
		Target:  enum.Slice(types.AdminStatusEnabled),
		Refresh: statusAdminAccountAdmin(ctx, conn, adminAccountID),
		Timeout: adminAccountEnabledTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.AdminAccount); ok {
		return output, err
	}

	return nil, err
}

// waitAdminAccountNotFound waits for an AdminAccount to return NotFound
func waitAdminAccountNotFound(ctx context.Context, conn *securityhub.Client, adminAccountID string) (*types.AdminAccount, error) {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(types.AdminStatusDisableInProgress),
		Target:  []string{adminStatusNotFound},
		Refresh: statusAdminAccountAdmin(ctx, conn, adminAccountID),
		Timeout: adminAccountNotFoundTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*types.AdminAccount); ok {
		return output, err
	}

	return nil, err
}
