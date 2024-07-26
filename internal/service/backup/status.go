// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package backup

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/backup"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func statusJobState(ctx context.Context, conn *backup.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findJobByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.State), nil
	}
}

func statusFramework(ctx context.Context, conn *backup.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findFrameworkByName(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.DeploymentStatus), nil
	}
}

func statusRecoveryPoint(ctx context.Context, conn *backup.Client, backupVaultName, recoveryPointARN string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := findRecoveryPointByTwoPartKey(ctx, conn, backupVaultName, recoveryPointARN)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}
