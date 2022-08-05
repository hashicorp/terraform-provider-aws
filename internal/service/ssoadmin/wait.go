package ssoadmin

import (
	"time"

	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	accountAssignmentCreateTimeout      = 5 * time.Minute
	accountAssignmentDeleteTimeout      = 5 * time.Minute
	accountAssignmentDelay              = 5 * time.Second
	accountAssignmentMinTimeout         = 3 * time.Second
	permissionSetProvisioningRetryDelay = 5 * time.Second
	permissionSetProvisionTimeout       = 10 * time.Minute
)

func waitAccountAssignmentCreated(conn *ssoadmin.SSOAdmin, instanceArn, requestID string) (*ssoadmin.AccountAssignmentOperationStatus, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{ssoadmin.StatusValuesInProgress},
		Target:     []string{ssoadmin.StatusValuesSucceeded},
		Refresh:    statusAccountAssignmentCreation(conn, instanceArn, requestID),
		Timeout:    accountAssignmentCreateTimeout,
		Delay:      accountAssignmentDelay,
		MinTimeout: accountAssignmentMinTimeout,
	}

	outputRaw, err := stateConf.WaitForState()
	if v, ok := outputRaw.(*ssoadmin.AccountAssignmentOperationStatus); ok {
		return v, err
	}

	return nil, err
}

func waitAccountAssignmentDeleted(conn *ssoadmin.SSOAdmin, instanceArn, requestID string) (*ssoadmin.AccountAssignmentOperationStatus, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{ssoadmin.StatusValuesInProgress},
		Target:     []string{ssoadmin.StatusValuesSucceeded},
		Refresh:    statusAccountAssignmentDeletion(conn, instanceArn, requestID),
		Timeout:    accountAssignmentDeleteTimeout,
		Delay:      accountAssignmentDelay,
		MinTimeout: accountAssignmentMinTimeout,
	}

	outputRaw, err := stateConf.WaitForState()
	if v, ok := outputRaw.(*ssoadmin.AccountAssignmentOperationStatus); ok {
		return v, err
	}

	return nil, err
}

func waitPermissionSetProvisioned(conn *ssoadmin.SSOAdmin, instanceArn, requestID string) (*ssoadmin.PermissionSetProvisioningStatus, error) {
	stateConf := resource.StateChangeConf{
		Delay:   permissionSetProvisioningRetryDelay,
		Pending: []string{ssoadmin.StatusValuesInProgress},
		Target:  []string{ssoadmin.StatusValuesSucceeded},
		Refresh: statusPermissionSetProvisioning(conn, instanceArn, requestID),
		Timeout: permissionSetProvisionTimeout,
	}
	outputRaw, err := stateConf.WaitForState()
	if v, ok := outputRaw.(*ssoadmin.PermissionSetProvisioningStatus); ok {
		return v, err
	}
	return nil, err
}
