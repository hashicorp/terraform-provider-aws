// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package efs

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/efs"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindBackupPolicyByID(ctx context.Context, conn *efs.EFS, id string) (*efs.BackupPolicy, error) {
	input := &efs.DescribeBackupPolicyInput{
		FileSystemId: aws.String(id),
	}

	output, err := conn.DescribeBackupPolicyWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, efs.ErrCodeFileSystemNotFound) {
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

func FindFileSystemPolicyByID(ctx context.Context, conn *efs.EFS, id string) (*efs.DescribeFileSystemPolicyOutput, error) {
	input := &efs.DescribeFileSystemPolicyInput{
		FileSystemId: aws.String(id),
	}

	output, err := conn.DescribeFileSystemPolicyWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, efs.ErrCodeFileSystemNotFound) || tfawserr.ErrCodeEquals(err, efs.ErrCodePolicyNotFound) {
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
