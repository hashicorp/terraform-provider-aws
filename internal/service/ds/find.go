// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ds

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directoryservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindRegion(ctx context.Context, conn *directoryservice.DirectoryService, directoryID, regionName string) (*directoryservice.RegionDescription, error) {
	input := &directoryservice.DescribeRegionsInput{
		DirectoryId: aws.String(directoryID),
		RegionName:  aws.String(regionName),
	}
	var output []*directoryservice.RegionDescription

	err := describeRegionsPages(ctx, conn, input, func(page *directoryservice.DescribeRegionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.RegionsDescription {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, directoryservice.ErrCodeDirectoryDoesNotExistException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if len(output) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	region := output[0]

	if status := aws.StringValue(region.Status); status == directoryservice.DirectoryStageDeleted {
		return nil, &retry.NotFoundError{
			Message:     status,
			LastRequest: input,
		}
	}

	return region, nil
}

func FindSharedDirectory(ctx context.Context, conn *directoryservice.DirectoryService, ownerDirectoryID, sharedDirectoryID string) (*directoryservice.SharedDirectory, error) { // nosemgrep:ci.ds-in-func-name
	input := &directoryservice.DescribeSharedDirectoriesInput{
		OwnerDirectoryId:   aws.String(ownerDirectoryID),
		SharedDirectoryIds: aws.StringSlice([]string{sharedDirectoryID}),
	}

	var output []*directoryservice.SharedDirectory

	err := conn.DescribeSharedDirectoriesPagesWithContext(ctx, input, func(page *directoryservice.DescribeSharedDirectoriesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.SharedDirectories {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, directoryservice.ErrCodeEntityDoesNotExistException) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if len(output) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	sharedDirectory := output[0]

	if status := aws.StringValue(sharedDirectory.ShareStatus); status == directoryservice.ShareStatusDeleted {
		return nil, &retry.NotFoundError{
			Message:     status,
			LastRequest: input,
		}
	}

	return sharedDirectory, nil
}
