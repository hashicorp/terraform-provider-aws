package ds

import (
	"context"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directoryservice"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func FindDirectoryByID(ctx context.Context, conn *directoryservice.DirectoryService, id string) (*directoryservice.DirectoryDescription, error) {
	input := &directoryservice.DescribeDirectoriesInput{
		DirectoryIds: aws.StringSlice([]string{id}),
	}
	var output []*directoryservice.DirectoryDescription

	err := describeDirectoriesPages(ctx, conn, input, func(page *directoryservice.DescribeDirectoriesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.DirectoryDescriptions {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, directoryservice.ErrCodeEntityDoesNotExistException) {
		return nil, &resource.NotFoundError{
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

	directory := output[0]

	if stage := aws.StringValue(directory.Stage); stage == directoryservice.DirectoryStageDeleted {
		return nil, &resource.NotFoundError{
			Message:     stage,
			LastRequest: input,
		}
	}

	return directory, nil
}

func FindDomainController(ctx context.Context, conn *directoryservice.DirectoryService, directoryID, domainControllerID string) (*directoryservice.DomainController, error) {
	input := &directoryservice.DescribeDomainControllersInput{
		DirectoryId:         aws.String(directoryID),
		DomainControllerIds: aws.StringSlice([]string{domainControllerID}),
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

	if status := aws.StringValue(domainController.Status); status == directoryservice.DomainControllerStatusDeleted {
		return nil, &resource.NotFoundError{
			Message:     status,
			LastRequest: input,
		}
	}

	return domainController, nil
}

func FindDomainControllers(ctx context.Context, conn *directoryservice.DirectoryService, input *directoryservice.DescribeDomainControllersInput) ([]*directoryservice.DomainController, error) {
	var output []*directoryservice.DomainController

	err := conn.DescribeDomainControllersPagesWithContext(ctx, input, func(page *directoryservice.DescribeDomainControllersOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, v := range page.DomainControllers {
			if v != nil {
				output = append(output, v)
			}
		}

		return !lastPage
	})

	if tfawserr.ErrCodeEquals(err, directoryservice.ErrCodeEntityDoesNotExistException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	return output, nil
}

func FindRadiusSettings(ctx context.Context, conn *directoryservice.DirectoryService, directoryID string) (*directoryservice.RadiusSettings, error) {
	output, err := FindDirectoryByID(ctx, conn, directoryID)

	if err != nil {
		return nil, err
	}

	if output.RadiusSettings == nil {
		return nil, tfresource.NewEmptyResultError(directoryID)
	}

	return output.RadiusSettings, nil
}

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
		return nil, &resource.NotFoundError{
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
		return nil, &resource.NotFoundError{
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

	output, err := conn.DescribeSharedDirectoriesWithContext(ctx, input)

	if tfawserr.ErrCodeEquals(err, directoryservice.ErrCodeEntityDoesNotExistException) {
		return nil, &resource.NotFoundError{
			LastError:   err,
			LastRequest: input,
		}
	}

	if err != nil {
		return nil, err
	}

	if output == nil || len(output.SharedDirectories) == 0 || output.SharedDirectories[0] == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	if count := len(output.SharedDirectories); count > 1 {
		return nil, tfresource.NewTooManyResultsError(count, input)
	}

	sharedDirectory := output.SharedDirectories[0]

	if status := aws.StringValue(sharedDirectory.ShareStatus); status == directoryservice.ShareStatusDeleted {
		return nil, &resource.NotFoundError{
			Message:     status,
			LastRequest: input,
		}
	}

	return sharedDirectory, nil
}
