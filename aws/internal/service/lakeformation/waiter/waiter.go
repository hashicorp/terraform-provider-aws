package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/lakeformation"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

const (
	PermissionsReadyTimeout       = 1 * time.Minute
	PermissionsDeleteRetryTimeout = 30 * time.Second

	StatusAvailable = "AVAILABLE"
	StatusNotFound  = "NOT FOUND"
	StatusFailed    = "FAILED"
	StatusIAMDelay  = "IAM DELAY"
)

func PermissionsReady(conn *lakeformation.LakeFormation, input *lakeformation.ListPermissionsInput, tableType string, columnNames []*string, excludedColumnNames []*string, columnWildcard bool) ([]*lakeformation.PrincipalResourcePermissions, error) {
	stateConf := &resource.StateChangeConf{
		Pending: []string{StatusNotFound, StatusIAMDelay},
		Target:  []string{StatusAvailable},
		Refresh: PermissionsStatus(conn, input, tableType, columnNames, excludedColumnNames, columnWildcard),
		Timeout: PermissionsReadyTimeout,
	}

	outputRaw, err := stateConf.WaitForState()

	if output, ok := outputRaw.([]*lakeformation.PrincipalResourcePermissions); ok {
		return output, err
	}

	return nil, err
}
