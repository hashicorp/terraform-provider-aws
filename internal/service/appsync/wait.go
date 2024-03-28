// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appsync

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/appsync"
	awstypes "github.com/aws/aws-sdk-go-v2/service/appsync/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/enum"
)

const (
	apiCacheAvailableTimeout           = 60 * time.Minute
	apiCacheDeletedTimeout             = 60 * time.Minute
	domainNameAPIAssociationTimeout    = 60 * time.Minute
	domainNameAPIDisassociationTimeout = 60 * time.Minute
)

func waitAPICacheAvailable(ctx context.Context, conn *appsync.Client, id string) error {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ApiCacheStatusCreating, awstypes.ApiCacheStatusModifying),
		Target:  enum.Slice(awstypes.ApiCacheStatusAvailable),
		Refresh: StatusAPICache(ctx, conn, id),
		Timeout: apiCacheAvailableTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitAPICacheDeleted(ctx context.Context, conn *appsync.Client, id string) error {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.ApiCacheStatusDeleting),
		Target:  []string{},
		Refresh: StatusAPICache(ctx, conn, id),
		Timeout: apiCacheDeletedTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitDomainNameAPIAssociation(ctx context.Context, conn *appsync.Client, id string) error {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AssociationStatusProcessing),
		Target:  enum.Slice(awstypes.AssociationStatusSuccess),
		Refresh: statusDomainNameAPIAssociation(ctx, conn, id),
		Timeout: domainNameAPIAssociationTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitDomainNameAPIDisassociation(ctx context.Context, conn *appsync.Client, id string) error {
	stateConf := &retry.StateChangeConf{
		Pending: enum.Slice(awstypes.AssociationStatusProcessing),
		Target:  []string{},
		Refresh: statusDomainNameAPIAssociation(ctx, conn, id),
		Timeout: domainNameAPIDisassociationTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}
