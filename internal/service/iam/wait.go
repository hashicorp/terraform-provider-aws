// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iam

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/aws/arn"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	awstypes "github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

const (
	// Maximum amount of time to wait for IAM changes to propagate
	// This timeout should not be increased without strong consideration
	// as this will negatively impact user experience when configurations
	// have incorrect references or permissions.
	// Reference: https://docs.aws.amazon.com/IAM/latest/UserGuide/troubleshoot_general.html#troubleshoot_general_eventual-consistency
	propagationTimeout = 2 * time.Minute

	RoleStatusARNIsUniqueID = "uniqueid"
	RoleStatusARNIsARN      = names.AttrARN
	RoleStatusNotFound      = "notfound"
)

func waitRoleARNIsNotUniqueID(ctx context.Context, conn *iam.Client, id string, role *awstypes.Role) (*awstypes.Role, error) {
	if arn.IsARN(aws.ToString(role.Arn)) {
		return role, nil
	}

	stateConf := &retry.StateChangeConf{
		Pending:                   []string{RoleStatusARNIsUniqueID, RoleStatusNotFound},
		Target:                    []string{names.AttrARN},
		Refresh:                   statusRoleCreate(ctx, conn, id),
		Timeout:                   propagationTimeout,
		NotFoundChecks:            10,
		ContinuousTargetOccurence: 5,
	}

	outputRaw, err := stateConf.WaitForStateContext(ctx)

	if output, ok := outputRaw.(*awstypes.Role); ok {
		return output, err
	}

	return nil, err
}

func statusRoleCreate(ctx context.Context, conn *iam.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		role, err := findRoleByName(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, RoleStatusNotFound, nil
		}

		if err != nil {
			return nil, "", err
		}

		if arn.IsARN(aws.ToString(role.Arn)) {
			return role, names.AttrARN, nil
		}

		return role, RoleStatusARNIsUniqueID, nil
	}
}
