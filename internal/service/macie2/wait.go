// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package macie2

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/macie2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

const (
	// Maximum amount of time to wait for the statusMemberRelationship to be Invited, Enabled, or Paused
	memberInvitedTimeout = 5 * time.Minute
)

// waitMemberInvited waits for an AdminAccount to return Invited, Enabled and Paused
func waitMemberInvited(ctx context.Context, conn *macie2.Macie2, adminAccountID string) (*macie2.Member, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: []string{macie2.RelationshipStatusCreated, macie2.RelationshipStatusEmailVerificationInProgress},
		Target:  []string{macie2.RelationshipStatusInvited, macie2.RelationshipStatusEnabled, macie2.RelationshipStatusPaused},
		Refresh: statusMemberRelationship(ctx, conn, adminAccountID),
		Timeout: memberInvitedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*macie2.Member); ok {
		return output, err
	}

	return nil, err
}
