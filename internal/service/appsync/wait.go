package appsync

import (
	"time"

	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	apiCacheAvailableTimeout           = 60 * time.Minute
	apiCacheDeletedTimeout             = 60 * time.Minute
	domainNameApiAssociationTimeout    = 60 * time.Minute
	domainNameApiDisassociationTimeout = 60 * time.Minute
)

func waitApiCacheAvailable(conn *appsync.AppSync, id string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{appsync.ApiCacheStatusCreating, appsync.ApiCacheStatusModifying},
		Target:  []string{appsync.ApiCacheStatusAvailable},
		Refresh: StatusApiCache(conn, id),
		Timeout: apiCacheAvailableTimeout,
	}

	_, err := stateConf.WaitForState()

	return err
}

func waitApiCacheDeleted(conn *appsync.AppSync, id string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{appsync.ApiCacheStatusDeleting},
		Target:  []string{},
		Refresh: StatusApiCache(conn, id),
		Timeout: apiCacheDeletedTimeout,
	}

	_, err := stateConf.WaitForState()

	return err
}

func waitDomainNameApiAssociation(conn *appsync.AppSync, id string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{appsync.AssociationStatusProcessing},
		Target:  []string{appsync.AssociationStatusSuccess},
		Refresh: statusDomainNameApiAssociation(conn, id),
		Timeout: domainNameApiAssociationTimeout,
	}

	_, err := stateConf.WaitForState()

	return err
}

func waitDomainNameApiDisassociation(conn *appsync.AppSync, id string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{appsync.AssociationStatusProcessing},
		Target:  []string{},
		Refresh: statusDomainNameApiAssociation(conn, id),
		Timeout: domainNameApiDisassociationTimeout,
	}

	_, err := stateConf.WaitForState()

	return err
}
