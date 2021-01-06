package waiter

import (
	"time"

	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	AWSSSOAdminPermissionSetDeleteTimeout          = 5 * time.Minute
	AWSSSOAdminPermissionSetProvisioningRetryDelay = 5 * time.Second
	AWSSSOAdminPermissionSetProvisionTimeout       = 10 * time.Minute
)

func InlinePolicyDeleted(conn *ssoadmin.SSOAdmin, instanceArn, permissionSetArn string) (*string, error) {
	stateConf := resource.StateChangeConf{
		Pending: []string{InlinePolicyDeleteStatusExists},
		Target:  []string{InlinePolicyDeleteStatusNotFound},
		Refresh: InlinePolicyDeletedStatus(conn, instanceArn, permissionSetArn),
		Timeout: AWSSSOAdminPermissionSetDeleteTimeout,
	}
	outputRaw, err := stateConf.WaitForState()
	if v, ok := outputRaw.(*string); ok {
		return v, err
	}
	return nil, err
}

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
