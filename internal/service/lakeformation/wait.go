package lakeformation

import (
	"time"

	"github.com/aws/aws-sdk-go/service/lakeformation"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	permissionsReadyTimeout       = 1 * time.Minute
	permissionsDeleteRetryTimeout = 30 * time.Second

	statusAvailable = "AVAILABLE"
	statusNotFound  = "NOT FOUND"
	statusFailed    = "FAILED"
	statusIAMDelay  = "IAM DELAY"
)

func waitPermissionsReady(conn *lakeformation.LakeFormation, input *lakeformation.ListPermissionsInput, tableType string, columnNames []*string, excludedColumnNames []*string, columnWildcard bool) ([]*lakeformation.PrincipalResourcePermissions, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{statusNotFound, statusIAMDelay},
		Target:  []string{statusAvailable},
		Refresh: statusPermissions(conn, input, tableType, columnNames, excludedColumnNames, columnWildcard),
		Timeout: permissionsReadyTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.([]*lakeformation.PrincipalResourcePermissions); ok {
		return output, err
	}

	return nil, err
}
