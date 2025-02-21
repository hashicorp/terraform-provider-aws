// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package macie2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/macie2"
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
