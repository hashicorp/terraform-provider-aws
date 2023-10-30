// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package detective

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/detective"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

const (
	// Maximum amount of time to wait for an Administrator to return Found
	adminAccountFoundTimeout = 5 * time.Minute
	// Maximum amount of time to wait for an Administrator to return NotFound
	adminAccountNotFoundTimeout = 5 * time.Minute
)

// waitAdminAccountFound waits for an AdminAccount to return Found
func waitAdminAccountFound(ctx context.Context, conn *detective.Detective, adminAccountID string) (*detective.Administrator, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{adminAccountStatusNotFound},
		Target:  []string{adminAccountStatusFound},
		Refresh: adminAccountStatus(ctx, conn, adminAccountID),
		Timeout: adminAccountFoundTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*detective.Administrator); ok {
		return output, err
	}

	return nil, err
}

// waitAdminAccountNotFound waits for an AdminAccount to return NotFound
func waitAdminAccountNotFound(ctx context.Context, conn *detective.Detective, adminAccountID string) (*detective.Administrator, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{adminAccountStatusFound},
		Target:  []string{adminAccountStatusNotFound},
		Refresh: adminAccountStatus(ctx, conn, adminAccountID),
		Timeout: adminAccountNotFoundTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*detective.Administrator); ok {
		return output, err
	}

	return nil, err
}
