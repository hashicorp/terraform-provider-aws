// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package detective

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/detective"
)

func FindAdminAccount(ctx context.Context, conn *detective.Detective, adminAccountID string) (*detective.Administrator, error) {
	input := &detective.ListOrganizationAdminAccountsInput{}
	var result *detective.Administrator

	err := conn.ListOrganizationAdminAccountsPagesWithContext(ctx, input, func(page *detective.ListOrganizationAdminAccountsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, adminAccount := range page.Administrators {
			if adminAccount == nil {
				continue
			}

			if aws.StringValue(adminAccount.AccountId) == adminAccountID {
				result = adminAccount
				return false
			}
		}

		return !lastPage
	})

	return result, err
}
