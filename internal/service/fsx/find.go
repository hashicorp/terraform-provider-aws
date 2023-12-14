// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fsx

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/fsx"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindBackupByID(ctx context.Context, conn *fsx.FSx, id string) (*fsx.Backup, error) {
	input := &fsx.DescribeBackupsInput{
		BackupIds: aws.StringSlice([]string{id}),
	}

	output, err := conn.DescribeBackupsWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, fsx.ErrCodeFileSystemNotFound) || tfawserr.ErrCodeEquals(err, fsx.ErrCodeBackupNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Backups) == 0 || output.Backups[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Backups[0], nil
}

func findFileCacheByID(ctx context.Context, conn *fsx.FSx, id string) (*fsx.FileCache, error) {
	input := &fsx.DescribeFileCachesInput{
		FileCacheIds: []*string{aws.String(id)},
	}
	var fileCaches []*fsx.FileCache

	err := conn.DescribeFileCachesPagesWithContext(ctx, input, func(page *fsx.DescribeFileCachesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}
		fileCaches = append(fileCaches, page.FileCaches...)

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, fsx.ErrCodeFileCacheNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}
	if err != nil {
		return nil, err
	}
	if len(fileCaches) == 0 || fileCaches[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}
	if count := len(fileCaches); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}
	return fileCaches[0], nil
}

func FindSnapshotByID(ctx context.Context, conn *fsx.FSx, id string) (*fsx.Snapshot, error) {
	input := &fsx.DescribeSnapshotsInput{
		SnapshotIds: aws.StringSlice([]string{id}),
	}

	output, err := conn.DescribeSnapshotsWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, fsx.ErrCodeVolumeNotFound) || tfawserr.ErrCodeEquals(err, fsx.ErrCodeSnapshotNotFound) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.Snapshots) == 0 || output.Snapshots[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Snapshots[0], nil
}

func FindSnapshots(ctx context.Context, conn *fsx.FSx, input *fsx.DescribeSnapshotsInput) ([]*fsx.Snapshot, error) {
	var output []*fsx.Snapshot

	err := conn.DescribeSnapshotsPagesWithContext(ctx, input, func(page *fsx.DescribeSnapshotsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.Snapshots {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, fsx.ErrCodeSnapshotNotFound) {
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
