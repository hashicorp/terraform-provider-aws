package waiter

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ram"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ram/finder"
)

const (
	resourceShareInvitationStatusNotFound = "NotFound"
	resourceShareInvitationStatusUnknown  = "Unknown"

	resourceShareStatusNotFound = "NotFound"
	resourceShareStatusUnknown  = "Unknown"
)

// ResourceShareInvitationStatus fetches the ResourceShareInvitation and its Status
func ResourceShareInvitationStatus(conn *ram.RAM, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		invitation, err := finder.ResourceShareInvitationByArn(conn, arn)

		if err != nil {
			return nil, resourceShareInvitationStatusUnknown, err
		}

		if invitation == nil {
			return nil, resourceShareInvitationStatusNotFound, nil
		}

		return invitation, aws.StringValue(invitation.Status), nil
	}
}

// ResourceShareOwnerSelfStatus fetches the ResourceShare and its Status
func ResourceShareOwnerSelfStatus(conn *ram.RAM, arn string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		share, err := finder.ResourceShareOwnerSelfByArn(conn, arn)

		if err != nil {
			if tfawserr.ErrCodeEquals(err, ram.ErrCodeUnknownResourceException) {
				return nil, resourceShareStatusNotFound, nil
			}

			return nil, resourceShareStatusUnknown, err
		}

		if share == nil {
			return nil, resourceShareStatusNotFound, nil
		}

		return share, aws.StringValue(share.Status), nil
	}
}
