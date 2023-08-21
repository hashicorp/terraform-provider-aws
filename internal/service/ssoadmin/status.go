// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssoadmin

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssoadmin"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
)

const (
	permissionSetProvisioningStatusUnknown  = "Unknown"
	permissionSetProvisioningStatusNotFound = "NotFound"
)

func statusPermissionSetProvisioning(ctx context.Context, conn *ssoadmin.SSOAdmin, instanceArn, requestID string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &ssoadmin.DescribePermissionSetProvisioningStatusInput{
			InstanceArn:                     aws.String(instanceArn),
			ProvisionPermissionSetRequestId: aws.String(requestID),
		}

		resp, err := conn.DescribePermissionSetProvisioningStatusWithContext(ctx, input)

		if err != nil {
			return nil, permissionSetProvisioningStatusUnknown, err
		}

		if resp == nil || resp.PermissionSetProvisioningStatus == nil {
			return nil, permissionSetProvisioningStatusNotFound, nil
		}

		return resp.PermissionSetProvisioningStatus, aws.StringValue(resp.PermissionSetProvisioningStatus.Status), nil
	}
}
