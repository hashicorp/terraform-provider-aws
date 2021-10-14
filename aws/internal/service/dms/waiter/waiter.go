package waiter

import (
	"time"

	dms "github.com/aws/aws-sdk-go/service/databasemigrationservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	tfdms "github.com/hashicorp/terraform-provider-aws/aws/internal/service/dms"
)

const (
	EndpointDeletedTimeout = 5 * time.Minute
)

func EndpointDeleted(conn *dms.DatabaseMigrationService, id string) (*dms.Endpoint, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{tfdms.EndpointStatusDeleting},
		Target:  []string{},
		Refresh: EndpointStatus(conn, id),
		Timeout: EndpointDeletedTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.(*dms.Endpoint); ok {
		return output, err
	}

	return nil, err
}
