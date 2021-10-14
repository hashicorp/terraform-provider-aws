package ssoadmin

import (
	"time"

	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	awsSSOAdminAccountAssignmentCreateTimeout      = 5 * time.Minute
	awsSSOAdminAccountAssignmentDeleteTimeout      = 5 * time.Minute
	awsSSOAdminAccountAssignmentDelay              = 5 * time.Second
	awsSSOAdminAccountAssignmentMinTimeout         = 3 * time.Second
	awsSSOAdminPermissionSetProvisioningRetryDelay = 5 * time.Second
	awsSSOAdminPermissionSetProvisionTimeout       = 10 * time.Minute
)

func waitAccountAssignmentCreated(conn *ssoadmin.SSOAdmin, instanceArn, requestID string) (*ssoadmin.AccountAssignmentOperationStatus, error) {
	stateConf := &resource.StateChangeConf{
		Pending:    []string{ssoadmin.StatusValuesInProgress},
		Target:     []string{ssoadmin.StatusValuesSucceeded},
		Refresh:    statusAccountAssignmentCreation(conn, instanceArn, requestID),
		Timeout:    awsSSOAdminAccountAssignmentCreateTimeout,
		Delay:      awsSSOAdminAccountAssignmentDelay,
		MinTimeout: awsSSOAdminAccountAssignmentMinTimeout,
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
		Timeout:    awsSSOAdminAccountAssignmentDeleteTimeout,
		Delay:      awsSSOAdminAccountAssignmentDelay,
		MinTimeout: awsSSOAdminAccountAssignmentMinTimeout,
	}

	outputRaw, err := stateConf.WaitForState()
	if v, ok := outputRaw.(*ssoadmin.AccountAssignmentOperationStatus); ok {
		return v, err
	}

	return nil, err
}

func waitPermissionSetProvisioned(conn *ssoadmin.SSOAdmin, instanceArn, requestID string) (*ssoadmin.PermissionSetProvisioningStatus, error) {
	stateConf := resource.StateChangeConf{
		Delay:   awsSSOAdminPermissionSetProvisioningRetryDelay,
		Pending: []string{ssoadmin.StatusValuesInProgress},
		Target:  []string{ssoadmin.StatusValuesSucceeded},
		Refresh: statusPermissionSetProvisioning(conn, instanceArn, requestID),
		Timeout: awsSSOAdminPermissionSetProvisionTimeout,
	}
	outputRaw, err := stateConf.WaitForState()
	if v, ok := outputRaw.(*ssoadmin.PermissionSetProvisioningStatus); ok {
		return v, err
	}
	return nil, err
}
