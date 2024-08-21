// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package macie2

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/service/macie2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

// statusMemberRelationship fetches the Member and its relationship status
func statusMemberRelationship(ctx context.Context, conn *macie2.Client, adminAccountID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		adminAccount, err := findMemberNotAssociated(ctx, conn, adminAccountID)

		if err != nil {
			return nil, "Unknown", err
		}

		if adminAccount == nil {
			return adminAccount, "NotFound", nil
		}

		return adminAccount, string(adminAccount.RelationshipStatus), nil
	}
}
