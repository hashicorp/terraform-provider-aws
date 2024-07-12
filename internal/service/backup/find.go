// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package backup

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/backup"
	awstypes "github.com/aws/aws-sdk-go-v2/service/backup/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func findJobByID(ctx context.Context, conn *backup.Client, id string) (*backup.DescribeBackupJobOutput, error) {
	input := &backup.DescribeBackupJobInput{
		BackupJobId: aws.String(id),
	}

	output, err := conn.DescribeBackupJob(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func findRecoveryPointByTwoPartKey(ctx context.Context, conn *backup.Client, backupVaultName, recoveryPointARN string) (*backup.DescribeRecoveryPointOutput, error) {
	input := &backup.DescribeRecoveryPointInput{
		BackupVaultName:  aws.String(backupVaultName),
		RecoveryPointArn: aws.String(recoveryPointARN),
	}

	output, err := conn.DescribeRecoveryPoint(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func findVaultAccessPolicyByName(ctx context.Context, conn *backup.Client, name string) (*backup.GetBackupVaultAccessPolicyOutput, error) {
	input := &backup.GetBackupVaultAccessPolicyInput{
		BackupVaultName: aws.String(name),
	}

	output, err := conn.GetBackupVaultAccessPolicy(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) || tfawserr.ErrCodeEquals(err, errCodeAccessDeniedException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func findVaultByName(ctx context.Context, conn *backup.Client, name string) (*backup.DescribeBackupVaultOutput, error) {
	input := &backup.DescribeBackupVaultInput{
		BackupVaultName: aws.String(name),
	}

	output, err := conn.DescribeBackupVault(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) || tfawserr.ErrCodeEquals(err, errCodeAccessDeniedException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output, nil
}

func findFrameworkByName(ctx context.Context, conn *backup.Client, name string) (*backup.DescribeFrameworkOutput, error) {
	input := &backup.DescribeFrameworkInput{
		FrameworkName: aws.String(name),
	}

	output, err := conn.DescribeFramework(ctx, input)

	if errs.IsA[*awstypes.ResourceNotFoundException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}
