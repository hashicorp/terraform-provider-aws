// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package efs

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/efs"
	awstypes "github.com/aws/aws-sdk-go-v2/service/efs/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindBackupPolicyByID(ctx context.Context, conn *efs.Client, id string) (*awstypes.BackupPolicy, error) {
	input := &efs.DescribeBackupPolicyInput{
		FileSystemId: aws.String(id),
	}

	output, err := conn.DescribeBackupPolicy(ctx, input)

	if errs.IsA[*awstypes.FileSystemNotFound](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || output.BackupPolicy == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.BackupPolicy, nil
}

func FindFileSystemPolicyByID(ctx context.Context, conn *efs.Client, id string) (*efs.DescribeFileSystemPolicyOutput, error) {
	input := &efs.DescribeFileSystemPolicyInput{
		FileSystemId: aws.String(id),
	}

	output, err := conn.DescribeFileSystemPolicy(ctx, input)

	if errs.IsA[*awstypes.FileSystemNotFound](err) || errs.IsA[*awstypes.PolicyNotFound](err) {
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
