// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package macie2

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/macie2"
	awstypes "github.com/aws/aws-sdk-go-v2/service/macie2/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
)

const (
	// Maximum amount of time to wait for the statusMemberRelationship to be Invited, Enabled, or Paused
	memberInvitedTimeout = 5 * time.Minute
)

// waitMemberInvited waits for an AdminAccount to return Invited, Enabled and Paused
func waitMemberInvited(ctx context.Context, conn *macie2.Client, adminAccountID string) (*awstypes.Member, error) { //nolint:unparam
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.RelationshipStatusCreated, awstypes.RelationshipStatusEmailVerificationInProgress),
		Target:  enum.Slice(awstypes.RelationshipStatusInvited, awstypes.RelationshipStatusEnabled, awstypes.RelationshipStatusPaused),
		Refresh: statusMemberRelationship(ctx, conn, adminAccountID),
		Timeout: memberInvitedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Member); ok {
		return output, err
	}

	return nil, err
}
