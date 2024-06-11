// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ds

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/directoryservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/directoryservice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindDirectoryByID(ctx context.Context, conn *directoryservice.Client, id string) (*awstypes.DirectoryDescription, error) {
	input := &directoryservice.DescribeDirectoriesInput{
		DirectoryIds: []string{id},
	}
	var output []awstypes.DirectoryDescription

	err := describeDirectoriesPages(ctx, conn, input, func(page *directoryservice.DescribeDirectoriesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		output = append(output, page.DirectoryDescriptions...)

		return !lastPage
	})

	if errs.IsA[*awstypes.EntityDoesNotExistException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	directory, err := tfresource.AssertSingleValueResult(output)

	if err != nil {
		return nil, err
	}

	if directory.Stage == awstypes.DirectoryStageDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(directory.Stage),
			LastRequest: input,
		}
	}

	return directory, nil
}

func FindDomainController(ctx context.Context, conn *directoryservice.Client, directoryID, domainControllerID string) (*awstypes.DomainController, error) {
	input := &directoryservice.DescribeDomainControllersInput{
		DirectoryId:         aws.String(directoryID),
		DomainControllerIds: []string{domainControllerID},
	}

	output, err := FindDomainControllers(ctx, conn, input)

	if err != nil {
		return nil, err
	}

	if len(output) == 0 {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	domainController := output[0]

	if domainController.Status == awstypes.DomainControllerStatusDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(domainController.Status),
			LastRequest: input,
		}
	}

	return domainController, nil
}

func FindDomainControllers(ctx context.Context, conn *directoryservice.Client, input *directoryservice.DescribeDomainControllersInput) ([]*awstypes.DomainController, error) {
	var output []awstypes.DomainController

	pages := directoryservice.NewDescribeDomainControllersPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.EntityDoesNotExistException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.DomainControllers...)
	}

	return tfslices.ToPointers(output), nil
}

func FindRadiusSettings(ctx context.Context, conn *directoryservice.Client, directoryID string) (*awstypes.RadiusSettings, error) {
	output, err := FindDirectoryByID(ctx, conn, directoryID)

	if err != nil {
		return nil, err
	}

	if output.RadiusSettings == nil {
		return nil, tfresource.NewEmptyResultError(directoryID)
	}

	return output.RadiusSettings, nil
}

func FindRegion(ctx context.Context, conn *directoryservice.Client, directoryID, regionName string) (*awstypes.RegionDescription, error) {
	input := &directoryservice.DescribeRegionsInput{
		DirectoryId: aws.String(directoryID),
		RegionName:  aws.String(regionName),
	}
	var output []awstypes.RegionDescription

	err := describeRegionsPages(ctx, conn, input, func(page *directoryservice.DescribeRegionsOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		output = append(output, page.RegionsDescription...)

		return !lastPage
	})

	if errs.IsA[*awstypes.DirectoryDoesNotExistException](err) {
		return nil, &retry.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	region, err := tfresource.AssertSingleValueResult(output)

	if err != nil {
		return nil, err
	}

	if region.Status == awstypes.DirectoryStageDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(region.Status),
			LastRequest: input,
		}
	}

	return region, nil
}

func FindSharedDirectory(ctx context.Context, conn *directoryservice.Client, ownerDirectoryID, sharedDirectoryID string) (*awstypes.SharedDirectory, error) { // nosemgrep:ci.ds-in-func-name
	input := &directoryservice.DescribeSharedDirectoriesInput{
		OwnerDirectoryId:   aws.String(ownerDirectoryID),
		SharedDirectoryIds: []string{sharedDirectoryID},
	}

	var output []awstypes.SharedDirectory

	pages := directoryservice.NewDescribeSharedDirectoriesPaginator(conn, input)

	for pages.HasMorePages() {
		page, err := pages.NextPage(ctx)

		if errs.IsA[*awstypes.EntityDoesNotExistException](err) {
			return nil, &retry.NotFoundError{
				LastError:   err,
				LastRequest: input,
			}
		}

		if err != nil {
			return nil, err
		}

		output = append(output, page.SharedDirectories...)
	}

	sharedDirectory, _ := tfresource.AssertSingleValueResult(output)

	if sharedDirectory.ShareStatus == awstypes.ShareStatusDeleted {
		return nil, &retry.NotFoundError{
			Message:     string(sharedDirectory.ShareStatus),
			LastRequest: input,
		}
	}

	return sharedDirectory, nil
}
