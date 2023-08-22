// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/securityhub"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

const (
	// Maximum amount of time to wait for an AdminAccount to return Enabled
	adminAccountEnabledTimeout = 5 * time.Minute

	// Maximum amount of time to wait for an AdminAccount to return NotFound
	adminAccountNotFoundTimeout = 5 * time.Minute
)

// waitAdminAccountEnabled waits for an AdminAccount to return Enabled
func waitAdminAccountEnabled(ctx context.Context, conn *securityhub.SecurityHub, adminAccountID string) (*securityhub.AdminAccount, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{adminStatusNotFound},
		Target:  []string{securityhub.AdminStatusEnabled},
		Refresh: statusAdminAccountAdmin(ctx, conn, adminAccountID),
		Timeout: adminAccountEnabledTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*securityhub.AdminAccount); ok {
		return output, err
	}

	return nil, err
}

// waitAdminAccountNotFound waits for an AdminAccount to return NotFound
func waitAdminAccountNotFound(ctx context.Context, conn *securityhub.SecurityHub, adminAccountID string) (*securityhub.AdminAccount, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{securityhub.AdminStatusDisableInProgress},
		Target:  []string{adminStatusNotFound},
		Refresh: statusAdminAccountAdmin(ctx, conn, adminAccountID),
		Timeout: adminAccountNotFoundTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*securityhub.AdminAccount); ok {
		return output, err
	}

	return nil, err
}
