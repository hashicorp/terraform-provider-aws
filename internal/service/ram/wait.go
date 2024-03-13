// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ram

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/ram"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

const (
	PrincipalAssociationTimeout    = 3 * time.Minute
	PrincipalDisassociationTimeout = 3 * time.Minute
)

func WaitResourceSharePrincipalAssociated(ctx context.Context, conn *ram.RAM, resourceShareARN, principal string) (*ram.ResourceShareAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ram.ResourceShareAssociationStatusAssociating, PrincipalAssociationStatusNotFound},
		Target:  []string{ram.ResourceShareAssociationStatusAssociated},
		Refresh: StatusResourceSharePrincipalAssociation(ctx, conn, resourceShareARN, principal),
		Timeout: PrincipalAssociationTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*ram.ResourceShareAssociation); ok {
		return v, err
	}

	return nil, err
}

func WaitResourceSharePrincipalDisassociated(ctx context.Context, conn *ram.RAM, resourceShareARN, principal string) (*ram.ResourceShareAssociation, error) {
	stateConf := &retry.StateChangeConf{
		Pending: []string{ram.ResourceShareAssociationStatusAssociated, ram.ResourceShareAssociationStatusDisassociating},
		Target:  []string{ram.ResourceShareAssociationStatusDisassociated, PrincipalAssociationStatusNotFound},
		Refresh: StatusResourceSharePrincipalAssociation(ctx, conn, resourceShareARN, principal),
		Timeout: PrincipalDisassociationTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if v, ok := outputRaw.(*ram.ResourceShareAssociation); ok {
		return v, err
	}

	return nil, err
}
