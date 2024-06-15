// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package macie2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/macie2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/macie2/types"
)

// findMemberNotAssociated Return a list of members not associated and compare with account ID
func findMemberNotAssociated(ctx context.Context, conn *awstypes.Client, accountID string) (*awstypes.Member, error) {
	input := &macie2.ListMembersInput{
		OnlyAssociated: aws.String("false"),
	}
	var result *awstypes.Member

	err := conn.ListMembersPagesWithContext(ctx, input, func(page *macie2.ListMembersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, member := range page.Members {
			if member == nil {
				continue
			}

			if aws.ToString(member.AccountId) == accountID {
				result = member
				return false
			}
		}

		return !lastPage
	})

	return result, err
}
