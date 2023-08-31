// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/arn"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

const (
	// Maximum amount of time to wait for IAM changes to propagate
	// This timeout should not be increased without strong consideration
	// as this will negatively impact user experience when configurations
	// have incorrect references or permissions.
	// Reference: https://docs.aws.amazon.com/IAM/latest/UserGuide/troubleshoot_general.html#troubleshoot_general_eventual-consistency
	propagationTimeout = 2 * time.Minute

	RoleStatusARNIsUniqueID = "uniqueid"
	RoleStatusARNIsARN      = "arn"
	RoleStatusNotFound      = "notfound"
)

func waitRoleARNIsNotUniqueID(ctx context.Context, conn *iam.IAM, id string, role *iam.Role) (*iam.Role, error) {
	if arn.IsARN(aws.StringValue(role.Arn)) {
		return role, nil
	}

	stateConf := &retry.StateChangeConf{
		Pending:                   []string{RoleStatusARNIsUniqueID, RoleStatusNotFound},
		Target:                    []string{RoleStatusARNIsARN},
		Refresh:                   statusRoleCreate(ctx, conn, id),
		Timeout:                   propagationTimeout,
		NotFoundChecks:            10,
		ContinuousTargetOccurence: 5,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*iam.Role); ok {
		return output, err
	}

	return nil, err
}

func statusRoleCreate(ctx context.Context, conn *iam.IAM, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		role, err := FindRoleByName(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, RoleStatusNotFound, nil
		}

		if err != nil {
			return nil, "", err
		}

		if arn.IsARN(aws.StringValue(role.Arn)) {
			return role, RoleStatusARNIsARN, nil
		}

		return role, RoleStatusARNIsUniqueID, nil
	}
}
