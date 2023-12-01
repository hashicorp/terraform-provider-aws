// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package securityhub

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/securityhub"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

const (
	// AdminStatus NotFound
	adminStatusNotFound = "NotFound"

	// AdminStatus Unknown
	adminStatusUnknown = "Unknown"
)

// statusAdminAccountAdmin fetches the AdminAccount and its AdminStatus
func statusAdminAccountAdmin(ctx context.Context, conn *securityhub.Client, adminAccountID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		adminAccount, err := FindAdminAccount(ctx, conn, adminAccountID)

		if err != nil {
			return nil, adminStatusUnknown, err
		}

		if adminAccount == nil {
			return adminAccount, adminStatusNotFound, nil
		}

		return adminAccount, string(adminAccount.Status), nil
	}
}
