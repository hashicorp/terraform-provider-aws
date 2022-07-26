package appsync

import (
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

func waitAPICacheAvailable(conn *appsync.AppSync, id string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{appsync.ApiCacheStatusCreating, appsync.ApiCacheStatusModifying},
		Target:  []string{appsync.ApiCacheStatusAvailable},
		Refresh: StatusAPICache(conn, id),
		Timeout: apiCacheAvailableTimeout,
	}

	_, err := stateConf.WaitForState()

	return err
}

func waitAPICacheDeleted(conn *appsync.AppSync, id string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{appsync.ApiCacheStatusDeleting},
		Target:  []string{},
		Refresh: StatusAPICache(conn, id),
		Timeout: apiCacheDeletedTimeout,
	}

	_, err := stateConf.WaitForState()

	return err
}

func waitDomainNameAPIAssociation(conn *appsync.AppSync, id string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{appsync.AssociationStatusProcessing},
		Target:  []string{appsync.AssociationStatusSuccess},
		Refresh: statusDomainNameAPIAssociation(conn, id),
		Timeout: domainNameAPIAssociationTimeout,
	}

	_, err := stateConf.WaitForState()

	return err
}

func waitDomainNameAPIDisassociation(conn *appsync.AppSync, id string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{appsync.AssociationStatusProcessing},
		Target:  []string{},
		Refresh: statusDomainNameAPIAssociation(conn, id),
		Timeout: domainNameAPIDisassociationTimeout,
	}

	_, err := stateConf.WaitForState()

	return err
}
