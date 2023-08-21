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
	permissionSetProvisioningRetryDelay = 5 * time.Second
	permissionSetProvisionTimeout       = 10 * time.Minute
)

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
