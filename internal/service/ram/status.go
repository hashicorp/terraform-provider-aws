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

	PrincipalAssociationStatusNotFound = "NotFound"
)

func StatusResourceSharePrincipalAssociation(ctx context.Context, conn *ram.RAM, resourceShareArn, principal string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		association, err := findResourceShareAssociationByShareARNAndPrincipal(ctx, conn, resourceShareArn, principal)

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
