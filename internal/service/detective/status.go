// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package detective

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/detective"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

const (
	// AdminAccountStatus Found
	adminAccountStatusFound = "Found"

	// AdminAccountStatus NotFound
	adminAccountStatusNotFound = "NotFound"

	// AdminAccountStatus Unknown
	adminAccountStatusUnknown = "Unknown"
)

// MemberStatus fetches the Member and its status
func MemberStatus(ctx context.Context, conn *detective.Detective, graphARN, adminAccountID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindMemberByGraphARNAndAccountID(ctx, conn, graphARN, adminAccountID)

		if err != nil {
			return nil, "Unknown", err
		}

		if output == nil {
			return output, "NotFound", nil
		}

		return output, aws.StringValue(output.Status), nil
	}
}

// adminAccountStatus fetches the AdminAccount
func adminAccountStatus(ctx context.Context, conn *detective.Detective, adminAccountID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		adminAccount, err := getOrganizationAdminAccount(ctx, conn, adminAccountID)

		if err != nil {
			return nil, adminAccountStatusUnknown, err
		}

		if adminAccount == nil {
			return adminAccount, adminAccountStatusNotFound, nil
		}

		return adminAccount, adminAccountStatusFound, nil
	}
}

func getOrganizationAdminAccount(ctx context.Context, conn *detective.Detective, adminAccountID string) (*detective.Administrator, error) {
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
