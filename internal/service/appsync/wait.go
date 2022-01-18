package appsync

import (
	"time"

	"github.com/aws/aws-sdk-go/service/appsync"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	ApiCacheAvailableTimeout = 60 * time.Minute
	ApiCacheDeletedTimeout   = 60 * time.Minute
)

func waitApiCacheAvailable(conn *appsync.AppSync, id string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{appsync.ApiCacheStatusCreating, appsync.ApiCacheStatusModifying},
		Target:  []string{appsync.ApiCacheStatusAvailable},
		Refresh: StatusApiCache(conn, id),
		Timeout: ApiCacheAvailableTimeout,
	}

	_, err := stateConf.WaitForState()

	return err
}

func waitApiCacheDeleted(conn *appsync.AppSync, id string) error {
	stateConf := &resource.StateChangeConf{
		Pending: []string{appsync.ApiCacheStatusDeleting},
		Target:  []string{},
		Refresh: StatusApiCache(conn, id),
		Timeout: ApiCacheDeletedTimeout,
	}

	_, err := stateConf.WaitForState()

	return err
}
