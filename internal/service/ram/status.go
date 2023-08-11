// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ram

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ram"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

const (
	ResourceShareInvitationStatusNotFound = "NotFound"
	ResourceShareInvitationStatusUnknown  = "Unknown"

	ResourceShareStatusNotFound = "NotFound"
	ResourceShareStatusUnknown  = "Unknown"

	PrincipalAssociationStatusNotFound = "NotFound"
)

// StatusResourceShareInvitation fetches the ResourceShareInvitation and its Status
func StatusResourceShareInvitation(ctx context.Context, conn *ram.RAM, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		invitation, err := FindResourceShareInvitationByARN(ctx, conn, arn)

		if err != nil {
			return nil, ResourceShareInvitationStatusUnknown, err
		}

		if invitation == nil {
			return nil, ResourceShareInvitationStatusNotFound, nil
		}

		return invitation, aws.StringValue(invitation.Status), nil
	}
}

// StatusResourceShareOwnerSelf fetches the ResourceShare and its Status
func StatusResourceShareOwnerSelf(ctx context.Context, conn *ram.RAM, arn string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		share, err := FindResourceShareOwnerSelfByARN(ctx, conn, arn)

		if tfawserr.ErrCodeEquals(err, ram.ErrCodeUnknownResourceException) {
			return nil, ResourceShareStatusNotFound, nil
		}

		if err != nil {
			return nil, ResourceShareStatusUnknown, err
		}

		if share == nil {
			return nil, ResourceShareStatusNotFound, nil
		}

		return share, aws.StringValue(share.Status), nil
	}
}

func StatusResourceSharePrincipalAssociation(ctx context.Context, conn *ram.RAM, resourceShareArn, principal string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		association, err := FindResourceSharePrincipalAssociationByShareARNPrincipal(ctx, conn, resourceShareArn, principal)

		if tfawserr.ErrCodeEquals(err, ram.ErrCodeUnknownResourceException) {
			return nil, PrincipalAssociationStatusNotFound, err
		}

		if err != nil {
			return nil, ram.ResourceShareAssociationStatusFailed, err
		}

		if association == nil {
			return nil, ram.ResourceShareAssociationStatusDisassociated, nil
		}

		if aws.StringValue(association.Status) == ram.ResourceShareAssociationStatusFailed {
			extendedErr := fmt.Errorf("association status message: %s", aws.StringValue(association.StatusMessage))
			return association, aws.StringValue(association.Status), extendedErr
		}

		return association, aws.StringValue(association.Status), nil
	}
}
