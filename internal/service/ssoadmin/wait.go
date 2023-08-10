// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssoadmin

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

const (
	accountAssignmentCreateTimeout      = 5 * time.Minute
	accountAssignmentDeleteTimeout      = 5 * time.Minute
	accountAssignmentDelay              = 10 * time.Second
	accountAssignmentMinTimeout         = 5 * time.Second
	permissionSetProvisioningRetryDelay = 5 * time.Second
	permissionSetProvisionTimeout       = 10 * time.Minute
)

func waitAccountAssignmentCreated(ctx context.Context, conn *ssoadmin.SSOAdmin, instanceArn, requestID string) (*ssoadmin.AccountAssignmentOperationStatus, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{ssoadmin.StatusValuesInProgress},
		Target:     []string{ssoadmin.StatusValuesSucceeded},
		Refresh:    statusAccountAssignmentCreation(ctx, conn, instanceArn, requestID),
		Timeout:    accountAssignmentCreateTimeout,
		Delay:      accountAssignmentDelay,
		MinTimeout: accountAssignmentMinTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if v, ok := outputRaw.(*ssoadmin.AccountAssignmentOperationStatus); ok {
		return v, err
	}

	return nil, err
}

func waitAccountAssignmentDeleted(ctx context.Context, conn *ssoadmin.SSOAdmin, instanceArn, requestID string) (*ssoadmin.AccountAssignmentOperationStatus, error) {
	stateConf := &retry.StateChangeConf{
		Pending:    []string{ssoadmin.StatusValuesInProgress},
		Target:     []string{ssoadmin.StatusValuesSucceeded},
		Refresh:    statusAccountAssignmentDeletion(ctx, conn, instanceArn, requestID),
		Timeout:    accountAssignmentDeleteTimeout,
		Delay:      accountAssignmentDelay,
		MinTimeout: accountAssignmentMinTimeout,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if v, ok := outputRaw.(*ssoadmin.AccountAssignmentOperationStatus); ok {
		return v, err
	}

	return nil, err
}

func waitPermissionSetProvisioned(ctx context.Context, conn *ssoadmin.SSOAdmin, instanceArn, requestID string) (*ssoadmin.PermissionSetProvisioningStatus, error) {
	stateConf := retry.StateChangeConf{
		Delay:   permissionSetProvisioningRetryDelay,
		Pending: []string{ssoadmin.StatusValuesInProgress},
		Target:  []string{ssoadmin.StatusValuesSucceeded},
		Refresh: statusPermissionSetProvisioning(ctx, conn, instanceArn, requestID),
		Timeout: permissionSetProvisionTimeout,
	}
	outputRaw, err := stateConf.WaitForStateContext(ctx)
	if v, ok := outputRaw.(*ssoadmin.PermissionSetProvisioningStatus); ok {
		return v, err
	}
	return nil, err
}
