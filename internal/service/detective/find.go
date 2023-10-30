// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package detective

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/detective"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
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

func FindInvitationByGraphARN(ctx context.Context, conn *detective.Detective, graphARN string) (*string, error) {
	input := &detective.ListInvitationsInput{}

	var result *string

	err := conn.ListInvitationsPagesWithContext(ctx, input, func(page *detective.ListInvitationsOutput, lastPage bool) bool {
		for _, invitation := range page.Invitations {
			if aws.StringValue(invitation.GraphArn) == graphARN {
				result = invitation.GraphArn
				return false
			}
		}
		return !lastPage
	})
	if tfawserr.ErrCodeEquals(err, detective.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, &retry.NotFoundError{
			Message:     fmt.Sprintf("No member found with arn %q ", graphARN),
			LastRequest: input,
		}
	}

	return result, nil
}

func FindMemberByGraphARNAndAccountID(ctx context.Context, conn *detective.Detective, graphARN string, accountID string) (*detective.MemberDetail, error) {
	input := &detective.ListMembersInput{
		GraphArn: aws.String(graphARN),
	}

	var result *detective.MemberDetail

	err := conn.ListMembersPagesWithContext(ctx, input, func(page *detective.ListMembersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, member := range page.MemberDetails {
			if member == nil {
				continue
			}

			if aws.StringValue(member.AccountId) == accountID {
				result = member
				return false
			}
		}

		return !lastPage
	})
	if tfawserr.ErrCodeEquals(err, detective.ErrCodeResourceNotFoundException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, &retry.NotFoundError{
			Message:     fmt.Sprintf("No member found with arn %q and accountID %q", graphARN, accountID),
			LastRequest: input,
		}
	}

	return result, nil
}
