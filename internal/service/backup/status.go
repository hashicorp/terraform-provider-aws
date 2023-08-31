// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package backup

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/backup"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func statusJobState(ctx context.Context, conn *backup.Backup, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindJobByID(ctx, conn, id)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.State), nil
	}
}

func statusFramework(ctx context.Context, conn *backup.Backup, id string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &backup.DescribeFrameworkInput{
			FrameworkName: aws.String(id),
		}

		output, err := conn.DescribeFrameworkWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, backup.ErrCodeResourceNotFoundException) {
			return output, backup.ErrCodeResourceNotFoundException, nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.DeploymentStatus), nil
	}
}

func statusRecoveryPoint(ctx context.Context, conn *backup.Backup, backupVaultName, recoveryPointARN string) retry.StateRefreshFunc {
	return func() (interface{}, string, error) {
		output, err := FindRecoveryPointByTwoPartKey(ctx, conn, backupVaultName, recoveryPointARN)

		if tfresource.NotFound(err) {
			return nil, "", nil
		}

		if err != nil {
			return nil, "", err
		}

		return output, aws.StringValue(output.Status), nil
	}
}
