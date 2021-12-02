package dms

import (
	"time"

	dms "github.com/aws/aws-sdk-go/service/databasemigrationservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	endpointDeletedTimeout = 5 * time.Minute
)

func waitEndpointDeleted(conn *dms.DatabaseMigrationService, id string) (*dms.Endpoint, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{endpointStatusDeleting},
		Target:  []string{},
		Refresh: statusEndpoint(conn, id),
		Timeout: endpointDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*dms.Endpoint); ok {
		return output, err
	}

	return nil, err
}
