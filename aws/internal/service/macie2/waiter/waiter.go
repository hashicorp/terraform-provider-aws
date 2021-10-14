package waiter

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/macie2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	// Maximum amount of time to wait for the statusMemberRelationship to be Invited, Enabled, or Paused
	memberInvitedTimeout = 5 * time.Minute
)

// waitMemberInvited waits for an AdminAccount to return Invited, Enabled and Paused
func waitMemberInvited(ctx context.Context, conn *macie2.Macie2, adminAccountID string) (*macie2.Member, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{macie2.RelationshipStatusCreated, macie2.RelationshipStatusEmailVerificationInProgress},
		Target:  []string{macie2.RelationshipStatusInvited, macie2.RelationshipStatusEnabled, macie2.RelationshipStatusPaused},
		Refresh: statusMemberRelationship(conn, adminAccountID),
		Timeout: memberInvitedTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*macie2.Member); ok {
		return output, err
	}

	return nil, err
}
