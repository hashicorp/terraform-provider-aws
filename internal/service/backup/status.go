// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package backup

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/backup"
	awstypes "github.com/aws/aws-sdk-go-v2/service/backup/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func statusJobState(ctx context.Context, conn *backup.Client, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindJobByID(ctx, conn, id)

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
		input := &backup.DescribeFrameworkInput{
			FrameworkName: aws.String(id),
		}

		output, err := conn.DescribeFramework(ctx, input)

		if errs.IsA[*awstypes.ResourceNotFoundException](err) {
			return output, err.Error(), nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.ToString(output.DeploymentStatus), nil
	}
}

func statusRecoveryPoint(ctx context.Context, conn *backup.Client, backupVaultName, recoveryPointARN string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindRecoveryPointByTwoPartKey(ctx, conn, backupVaultName, recoveryPointARN)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, string(output.Status), nil
	}
}
