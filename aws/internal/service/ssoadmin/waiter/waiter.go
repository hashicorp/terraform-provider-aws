package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	AWSSSOAdminPermissionSetProvisioningRetryDelay = 5 * time.Second
	AWSSSOAdminPermissionSetProvisionTimeout       = 10 * time.Minute
)

func PermissionSetProvisioned(conn *ssoadmin.SSOAdmin, instanceArn, requestID string) (*ssoadmin.PermissionSetProvisioningStatus, error) {
	stateConf := resource.StateChangeConf{
		Delay:   AWSSSOAdminPermissionSetProvisioningRetryDelay,
		Pending: []string{ssoadmin.StatusValuesInProgress},
		Target:  []string{ssoadmin.StatusValuesSucceeded},
		Refresh: PermissionSetProvisioningStatus(conn, instanceArn, requestID),
		Timeout: AWSSSOAdminPermissionSetProvisionTimeout,
	}
	outputRaw, err := stateConf.WaitForState()
	if v, ok := outputRaw.(*ssoadmin.PermissionSetProvisioningStatus); ok {
		return v, err
	}
	return nil, err
}
