package appsync

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	apiCacheAvailableTimeout           = 60 * time.Minute
	apiCacheDeletedTimeout             = 60 * time.Minute
	domainNameAPIAssociationTimeout    = 60 * time.Minute
	domainNameAPIDisassociationTimeout = 60 * time.Minute
)

func waitAPICacheAvailable(ctx context.Context, conn *appsync.AppSync, id string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{appsync.ApiCacheStatusCreating, appsync.ApiCacheStatusModifying},
		Target:  []string{appsync.ApiCacheStatusAvailable},
		Refresh: StatusAPICache(ctx, conn, id),
		Timeout: apiCacheAvailableTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitAPICacheDeleted(ctx context.Context, conn *appsync.AppSync, id string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{appsync.ApiCacheStatusDeleting},
		Target:  []string{},
		Refresh: StatusAPICache(ctx, conn, id),
		Timeout: apiCacheDeletedTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitDomainNameAPIAssociation(ctx context.Context, conn *appsync.AppSync, id string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{appsync.AssociationStatusProcessing},
		Target:  []string{appsync.AssociationStatusSuccess},
		Refresh: statusDomainNameAPIAssociation(ctx, conn, id),
		Timeout: domainNameAPIAssociationTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}

func waitDomainNameAPIDisassociation(ctx context.Context, conn *appsync.AppSync, id string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{appsync.AssociationStatusProcessing},
		Target:  []string{},
		Refresh: statusDomainNameAPIAssociation(ctx, conn, id),
		Timeout: domainNameAPIDisassociationTimeout,
	}

	_, err := stateConf.WaitForStateContext(ctx)

	return err
}
