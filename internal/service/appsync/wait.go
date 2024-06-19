// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appsync

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

const (
	domainNameAPIAssociationTimeout    = 60 * time.Minute
	domainNameAPIDisassociationTimeout = 60 * time.Minute
)

func waitDomainNameAPIAssociation(ctx context.Context, conn *appsync.AppSync, id string) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{appsync.AssociationStatusProcessing},
		Target:  []string{appsync.AssociationStatusSuccess},
		Refresh: statusDomainNameAPIAssociation(ctx, conn, id),
		Timeout: domainNameAPIAssociationTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitDomainNameAPIDisassociation(ctx context.Context, conn *appsync.AppSync, id string) error {
	stateConf := &retry.StateChangeConf{
		Pending: []string{appsync.AssociationStatusProcessing},
		Target:  []string{},
		Refresh: statusDomainNameAPIAssociation(ctx, conn, id),
		Timeout: domainNameAPIDisassociationTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}
