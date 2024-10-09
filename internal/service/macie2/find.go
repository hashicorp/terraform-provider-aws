// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package macie2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/macie2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/macie2/types"
)

func findInvitationByAccount(ctx context.Context, conn *macie2.Client, accountID string) (string, error) {
	input := &macie2.ListInvitationsInput{}

	pages := macie2.NewListInvitationsPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return "", err
		}

		for _, invitation := range page.Invitations {
			if aws.ToString(invitation.AccountId) == accountID {
				return aws.ToString(invitation.InvitationId), nil
			}
		}
	}

	return "", nil
}

// findMemberNotAssociated Return a list of members not associated and compare with account ID
func findMemberNotAssociated(ctx context.Context, conn *macie2.Client, accountID string) (*awstypes.Member, error) {
	input := &macie2.ListMembersInput{
		OnlyAssociated: aws.String("false"),
	}

	pages := macie2.NewListMembersPaginator(conn, input)
	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)
		if err != nil {
			return nil, err
		}

		for _, member := range page.Members {
			if aws.ToString(member.AccountId) == accountID {
				return &member, nil
			}
		}
	}

	return nil, nil
}
